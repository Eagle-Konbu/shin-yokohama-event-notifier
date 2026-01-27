package main

import (
	"context"
	"log"

	"github.com/aws/aws-lambda-go/lambda"

	"github.com/Eagle-Konbu/shin-yokohama-event-notifier/internal/application/service"
	"github.com/Eagle-Konbu/shin-yokohama-event-notifier/internal/domain/ports"
	"github.com/Eagle-Konbu/shin-yokohama-event-notifier/internal/infrastructure/config"
	"github.com/Eagle-Konbu/shin-yokohama-event-notifier/internal/infrastructure/discord"
	"github.com/Eagle-Konbu/shin-yokohama-event-notifier/internal/infrastructure/scraper"

	lambdaHandler "github.com/Eagle-Konbu/shin-yokohama-event-notifier/internal/infrastructure/lambda"
)

func main() {
	ctx := context.Background()
	cfg, err := config.LoadConfig(ctx)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	fetchers := []ports.EventFetcher{
		scraper.NewYokohamaArenaScraper(),
		scraper.NewNissanStadiumScraper(),
		scraper.NewSkateCenterScraper(),
	}

	discordSender := discord.NewWebhookAdapter(cfg.DiscordWebhookURL)
	eventService := service.NewEventNotificationService(discordSender, fetchers)
	handler := lambdaHandler.NewHandler(eventService)

	lambda.Start(handler.HandleRequest)
}
