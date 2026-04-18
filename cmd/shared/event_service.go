package shared

import (
	"context"
	"fmt"

	"github.com/Eagle-Konbu/shin-yokohama-event-notifier/internal/application/service"
	"github.com/Eagle-Konbu/shin-yokohama-event-notifier/internal/domain/ports"
	"github.com/Eagle-Konbu/shin-yokohama-event-notifier/internal/infrastructure/config"
	"github.com/Eagle-Konbu/shin-yokohama-event-notifier/internal/infrastructure/discord"
	"github.com/Eagle-Konbu/shin-yokohama-event-notifier/internal/infrastructure/fetcher"
)

func BuildEventService(ctx context.Context) (*service.EventNotificationService, error) {
	return BuildEventServiceWithClient(ctx, nil)
}

func BuildEventServiceWithClient(ctx context.Context, client config.SecretsManagerClient) (*service.EventNotificationService, error) {
	cfg, err := config.LoadConfigWithClient(ctx, client)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	fetchers := []ports.EventFetcher{
		fetcher.NewYokohamaArenaFetcher(),
		fetcher.NewNissanStadiumFetcher(),
		fetcher.NewSkateCenterFetcher(),
	}

	discordSender := discord.NewWebhookAdapter(cfg.DiscordWebhookURL)
	eventService := service.NewEventNotificationService(discordSender, fetchers)

	return eventService, nil
}
