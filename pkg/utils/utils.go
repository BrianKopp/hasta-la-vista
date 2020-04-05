package utils

import (
	"os"
	"strconv"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// SetLogLevel assigns the log level based on the LOGLEVEL environment variable
func SetLogLevel() {
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

// GetPort gets the port from the PORT environment variable, default 8080
func GetPort() int {
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

// GetAppSecret gets the app secret from the SECRET environment variable
func GetAppSecret() string {
	appSecret, exists := os.LookupEnv("SECRET")
	if !exists || appSecret == "" {
		log.Fatal().Msg("SECRET environment variable not found, exiting")
		os.Exit(1)
	}

	return appSecret
}

// GetClusterName gets the cluster name from the CLUSTERNAME environment variable
func GetClusterName() string {
	clusterName, exists := os.LookupEnv("CLUSTERNAME")
	if !exists || clusterName == "" {
		log.Fatal().Msg("CLUSTERNAME environment variable not found, exiting")
		os.Exit(1)
	}

	return clusterName
}

// GetCloudProviderType gets the cloud proivder type from the CLOUD_PROVIDER environment variable
func GetCloudProviderType() string {
	cloudProvider, exists := os.LookupEnv("CLOUD_PROVIDER")
	if !exists || cloudProvider == "" {
		log.Fatal().Msg("CLOUD_PROVIDER environment variable not found, exiting")
		os.Exit(1)
	}

	return cloudProvider
}

// GetAWSRegion gets the AWS region from the AWS_REGION environment variable
func GetAWSRegion() string {
	awsRegion, exists := os.LookupEnv("AWS_REGION")
	if !exists || awsRegion == "" {
		log.Fatal().Msg("AWS_REGION environment variable not found, exiting")
		os.Exit(1)
	}

	return awsRegion
}
