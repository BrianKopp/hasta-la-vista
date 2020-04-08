package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/aws/aws-sdk-go/service/elbv2"
	awsProvider "github.com/briankopp/hasta-la-vista/pkg/cloudproviders/aws"
	"github.com/briankopp/hasta-la-vista/pkg/deregister"
	"github.com/briankopp/hasta-la-vista/pkg/utils"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func buildCloudProvider(whichProvider string) (deregister.CloudProvider, error) {
	if whichProvider == "aws" {
		log.Info().Msg("building cloud provider for AWS")
		awsSession := session.Must(session.NewSession())
		config := aws.Config{Region: aws.String(utils.GetAWSRegion())}
		elbClient := elb.New(awsSession, &config)
		elbV2Client := elbv2.New(awsSession, &config)
		ec2Client := ec2.New(awsSession, &config)
		timeout := utils.GetTimeout()
		provider := &awsProvider.CloudProvider{
			ELB:     elbClient,
			ELBV2:   elbV2Client,
			EC2:     ec2Client,
			Timeout: timeout,
			DryRun:  utils.IsDryRun(),
		}
		return provider, nil
	}

	return nil, errors.New("Unrecognized cloud provider")
}

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	utils.SetLogLevel()
	appSecret := utils.GetAppSecret()
	whichProvider := utils.GetCloudProviderType()
	provider, err := buildCloudProvider(whichProvider)
	if err != nil {
		log.Fatal().Err(err).Msg("error getting cloud provider")
		os.Exit(1)
	}

	svr := &http.Server{Addr: fmt.Sprintf(":%v", 80)}
	http.HandleFunc("/health", func(response http.ResponseWriter, request *http.Request) {
		fmt.Fprint(response, "OK")
	})
	http.HandleFunc("/drain", func(response http.ResponseWriter, request *http.Request) {
		if request.Method != "POST" {
			log.Warn().Str("Method", request.Method).Msg("received /drain unallowed method")
			response.Header().Set("Allow", "POST")
			return
		}

		password := request.URL.Query().Get("pw")
		if password == "" || password != appSecret {
			log.Error().Msg("provided password invalid")
			response.WriteHeader(403)
			return
		}

		nodeName := request.URL.Query().Get("node")
		err := provider.DrainNodeFromLoadBalancer(nodeName)
		if err != nil {
			response.WriteHeader(500)
			return
		}

		fmt.Fprint(response, "OK")
	})

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGTERM)
	signal.Notify(done, os.Interrupt, syscall.SIGINT)
	log.Info().Msg("HTTP server started and listening on port 80")

	// Wait for an OS signal
	<-done
	log.Info().Msg("received signal, shutting down server")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer func() {
		cancel()
	}()

	if err := svr.Shutdown(ctx); err != nil {
		log.Error().Err(err).Msg("error while shutting down server")
		os.Exit(1)
	}

	log.Info().Msg("successfully closed the http server")
	return
}
