package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	setLogLevel()
	appSecret := getAppSecret()
	clusterName := getClusterName()

	svr := &http.Server{Addr: fmt.Sprintf(":%v", getPort())}
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

		nodeID := request.URL.Query().Get("id")
		nodeIP := request.URL.Query().Get("ip")
		vpcID := request.URL.Query().Get("vpcid")
		err := handleDeregistration(nodeID, nodeIP, clusterName, vpcID)
		if err != nil {
			response.WriteHeader(500)
			return
		}

		fmt.Fprint(response, "OK")
	})

	done := make(chan os.Signal, 1)

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
