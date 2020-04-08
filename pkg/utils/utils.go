package utils

import (
	"os"
	"strconv"
	"time"

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

// GetAppSecret gets the app secret from the SECRET environment variable
func GetAppSecret() string {
	appSecret, exists := os.LookupEnv("SECRET")
	if !exists || appSecret == "" {
		log.Fatal().Msg("SECRET environment variable not found, exiting")
		os.Exit(1)
	}

	return appSecret
}

// GetCloudProviderType gets the cloud proivder type from the CLOUD_PROVIDER environment variable
func GetCloudProviderType() string {
	cloudProvider, exists := os.LookupEnv("CLOUDPROVIDER")
	if !exists || cloudProvider == "" {
		log.Fatal().Msg("CLOUDPROVIDER environment variable not found, exiting")
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

// GetTimeout gets the TIMEOUT environment variable in seconds
func GetTimeout() time.Duration {
	timeoutSeconds, exists := os.LookupEnv("TIMEOUT")
	if !exists || timeoutSeconds == "" {
		log.Info().Msg("No TIMEOUT environment variable found, defaulting to 60s")
		return 60 * time.Second
	}

	timeoutSecondsInt, err := strconv.Atoi(timeoutSeconds)
	if err != nil {
		log.Error().Err(err).Msg("Error parsing TIMEOUT environment variable, defaulting to 60s")
		return 60 * time.Second
	}

	return time.Duration(timeoutSecondsInt) * time.Second
}

// IsDryRun gets whether the lambda is a dry run. DRYRUN environment variable must be 1 if true, else false
func IsDryRun() bool {
	dryRun, exists := os.LookupEnv("DRYRUN")
	if exists && dryRun == "1" {
		log.Info().Msg("DRYRUN set to true")
		return true
	}

	log.Info().Msg("Running in normal mode (not DRYRUN)")
	return false
}
