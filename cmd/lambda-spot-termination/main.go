package main

import (
	"context"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/rs/zerolog/log"
)

func HandleRequest(ctx context.Context, req events.CloudWatchEvent) {
	if req.DetailType != "EC2 Spot Instance Interruption Warning" {
		log.Warn().
			Str("detail-type", req.DetailType).
			Msg("received unexpected detail-type request")
		return
	}

	return
}

func main() {
	lambda.Start(HandleRequest)
}
