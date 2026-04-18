package main

import (
	"context"
	"log"

	"github.com/aws/aws-lambda-go/lambda"

	"github.com/Eagle-Konbu/shin-yokohama-event-notifier/cmd/shared"

	lambdaHandler "github.com/Eagle-Konbu/shin-yokohama-event-notifier/internal/infrastructure/lambda"
)

func main() {
	ctx := context.Background()
	eventService, err := shared.BuildEventService(ctx)
	if err != nil {
		log.Fatalf("Failed to initialize app: %v", err)
	}

	handler := lambdaHandler.NewWeeklyHandler(eventService)

	lambda.Start(handler.HandleRequest)
}
