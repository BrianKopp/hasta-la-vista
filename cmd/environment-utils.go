package main

import (
	"os"
	"strconv"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// setLogLevel assigns the log level based on the LOGLEVEL environment variable
func setLogLevel() {
	lvl, keyExists := os.LookupEnv("LOGLEVEL")
	if !keyExists {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
		return
	}

	if lvl == "INFO" || lvl == "info" {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
		return
	}

	if lvl == "DEBUG" || lvl == "debug" {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
		return
	}

	if lvl == "WARN" || lvl == "warn" {
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
		return
	}

	if lvl == "ERROR" || lvl == "error" {
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
		return
	}
}

// getPort gets the port from the PORT environment variable, default 8080
func getPort() int {
	portStr, exists := os.LookupEnv("PORT")
	if !exists {
		log.Info().Msg("PORT environment variable not found, defaulting port to 8080")
		return 8080
	}

	port, err := strconv.Atoi(portStr)
	if err != nil {
		log.Warn().
			Str("port", portStr).
			Msg("Error converting PORT environment variable to number, defaulting port to 8080")
		return 8080
	}

	log.Info().Int("port", port).Msg("PORT environment variable successfully resolved")
	return port
}

func getAppSecret() string {
	appSecret, exists := os.LookupEnv("SECRET")
	if !exists || appSecret != "" {
		log.Fatal().Msg("SECRET environment variable not found, exiting")
		os.Exit(1)
	}

	return appSecret
}

func getClusterName() string {
	clusterName, exists := os.LookupEnv("CLUSTERNAME")
	if !exists || clusterName != "" {
		log.Fatal().Msg("CLUSTERNAME environment variable not found, exiting")
		os.Exit(1)
	}

	return clusterName
}

func getAWSRegion() string {
	awsRegion, exists := os.LookupEnv("AWS_REGION")
	if !exists || awsRegion != "" {
		log.Fatal().Msg("AWS_REGION environment variable not found, exiting")
		os.Exit(1)
	}

	return awsRegion
}
