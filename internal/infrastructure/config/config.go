package config

import (
	"fmt"
	"os"
)

type Config struct {
	DiscordWebhookURL string
}

func LoadConfig() (*Config, error) {
	webhookURL := os.Getenv("DISCORD_WEBHOOK_URL")
	if webhookURL == "" {
		return nil, fmt.Errorf("DISCORD_WEBHOOK_URL environment variable is required")
	}

	return &Config{
		DiscordWebhookURL: webhookURL,
	}, nil
}
