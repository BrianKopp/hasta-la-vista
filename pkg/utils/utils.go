package utils

import (
	"os"

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

// GetClusterName gets the cluster name from the CLUSTERNAME environment variable
func GetClusterName() string {
	clusterName, exists := os.LookupEnv("CLUSTERNAME")
	if !exists || clusterName == "" {
		log.Fatal().Msg("CLUSTERNAME environment variable not found, exiting")
		os.Exit(1)
	}

	return clusterName
}

// GetVPCID gets the VPC ID from the VPCID environment variable
func GetVPCID() string {
	vpcID, exists := os.LookupEnv("VPCID")
	if !exists || vpcID == "" {
		log.Fatal().Msg("VPCID environment variable not found, exiting")
		os.Exit(1)
	}

	return vpcID
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
