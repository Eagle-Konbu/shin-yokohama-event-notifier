package main

import (
	"log"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/Eagle-Konbu/shin-yokohama-event-notifier/internal/application/service"
	"github.com/Eagle-Konbu/shin-yokohama-event-notifier/internal/infrastructure/config"
	"github.com/Eagle-Konbu/shin-yokohama-event-notifier/internal/infrastructure/discord"
	lambdaHandler "github.com/Eagle-Konbu/shin-yokohama-event-notifier/internal/infrastructure/lambda"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Wire dependencies (manual dependency injection)
	// Infrastructure layer: Create Discord webhook adapter
	discordSender := discord.NewWebhookAdapter(cfg.DiscordWebhookURL)

	// Application layer: Create service with injected port implementation
	eventService := service.NewEventNotificationService(discordSender)

	// Infrastructure layer: Create Lambda handler
	handler := lambdaHandler.NewHandler(eventService)

	// Start Lambda runtime
	lambda.Start(handler.HandleRequest)
}
