package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/aws/aws-sdk-go/service/elbv2"
	awsProvider "github.com/briankopp/hasta-la-vista/pkg/cloudproviders/aws"
	"github.com/briankopp/hasta-la-vista/pkg/utils"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// instanceDetail struct is used for decoding the CW event
type instanceDetail struct {
	InstanceID     string `json:"instance-id"`
	InstanceAction string `json:"instance-action"`
}

// HandleSpotTerminationRequest is the lambda handler for Spot Termination CW Events
func HandleSpotTerminationRequest(ctx context.Context, req events.CloudWatchEvent) error {
	if req.DetailType != "EC2 Spot Instance Interruption Warning" {
		log.Warn().
			Str("detail-type", req.DetailType).
			Msg("received unexpected detail-type request")
		return errors.New("received unexpected detail-type request")
	}

	var details instanceDetail
	err := json.Unmarshal(req.Detail, &details)
	if err != nil {
		log.Error().
			Err(err).
			Str("details", fmt.Sprintf("%v", req.Detail)).
			Msg("Unable to decode the instance details")
		return err
	}

	if details.InstanceAction != "terminate" {
		log.Warn().
			Str("instanceAction", details.InstanceAction).
			Msg("instance-action not terminate")
		return errors.New("instance-action not terminate")
	}

	// Acquire AWS client
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

	err = provider.DrainNodeFromLoadBalancer(details.InstanceID)
	if err != nil {
		log.Error().
			Err(err).
			Msg("Error draining node from load balencers")
		return err
	}

	log.Info().Str("instanceId", details.InstanceID).Msg("Successfully drained node from load balancer")
	return nil
}

func setupLogger() {
	output := zerolog.ConsoleWriter{
		NoColor:    true,
		Out:        os.Stdout,
		TimeFormat: time.RFC3339,
	}

	output.FormatLevel = func(i interface{}) string {
		return strings.ToUpper(fmt.Sprintf("| %-6s|", i))
	}

	output.FormatMessage = func(i interface{}) string {
		return fmt.Sprintf("%s |", i)
	}

	output.FormatFieldName = func(i interface{}) string {
		return fmt.Sprintf("%s=", i)
	}

	output.FormatFieldValue = func(i interface{}) string {
		return fmt.Sprintf("%s", i)
	}

	log.Logger = zerolog.New(output).With().Timestamp().Logger()
	utils.SetLogLevel()
}

func main() {
	setupLogger()
	lambda.Start(HandleSpotTerminationRequest)
}
