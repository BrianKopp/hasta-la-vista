package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/aws/aws-sdk-go/service/elbv2"
	awsProvider "github.com/briankopp/hasta-la-vista/pkg/cloudproviders/aws"
	"github.com/briankopp/hasta-la-vista/pkg/utils"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// asgDetails struct is used for decoding the CW event
type asgDetails struct {
	LifecycleActionToken string `json:"LifecycleActionToken"`
	EC2InstanceID        string `json:"EC2InstanceId"`
	LifecycleTransition  string `json:"LifecycleTransition"`
	LifecycleHookName    string `json:"LifecycleHookName"`
	AutoscalingGroupName string `json:"AutoScalingGroupName"`
}

// HandleSpotTerminationRequest is the lambda handler for Spot Termination CW Events
func HandleSpotTerminationRequest(ctx context.Context, req events.CloudWatchEvent) {
	if req.DetailType != "EC2 Instance-launch Lifecycle Action" {
		log.Warn().
			Str("detail-type", req.DetailType).
			Msg("received unexpected detail-type request")
		return
	}

	var details asgDetails
	err := json.Unmarshal(req.Detail, &details)
	if err != nil {
		log.Error().
			Err(err).
			Str("details", fmt.Sprintf("%v", req.Detail)).
			Msg("Unable to decode the instance details")
		return
	}

	if details.LifecycleTransition != "autoscaling:EC2_INSTANCE_TERMINATING" {
		log.Warn().
			Str("transition", details.LifecycleTransition).
			Msg("Transition not terminate")
		return
	}

	// Acquire AWS client
	awsSession := session.Must(session.NewSession())
	config := aws.Config{Region: aws.String(utils.GetAWSRegion())}
	elbClient := elb.New(awsSession, &config)
	elbV2Client := elbv2.New(awsSession, &config)
	ec2Client := ec2.New(awsSession, &config)
	provider := &awsProvider.CloudProvider{
		ELB:   elbClient,
		ELBV2: elbV2Client,
		EC2:   ec2Client,
	}

	err = provider.DrainNodeFromLoadBalancer(details.EC2InstanceID)
	if err != nil {
		log.Error().
			Err(err).
			Msg("Error draining node from load balencers")
		return
	}

	log.Info().Str("instanceId", details.EC2InstanceID).Msg("Successfully drained node from load balancer")

	asgClient := autoscaling.New(awsSession, &config)
	_, err = asgClient.CompleteLifecycleAction(
		&autoscaling.CompleteLifecycleActionInput{
			LifecycleActionResult: aws.String("CONTINUE"),
			LifecycleActionToken:  &details.LifecycleActionToken,
			LifecycleHookName:     &details.LifecycleHookName,
			AutoScalingGroupName:  &details.AutoscalingGroupName,
		},
	)

	if err != nil {
		log.Error().
			Err(err).
			Msg("Error completing lifecycle action")
		return
	}

	return
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
