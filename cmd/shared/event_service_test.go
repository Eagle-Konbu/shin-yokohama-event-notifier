package shared

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Eagle-Konbu/shin-yokohama-event-notifier/internal/infrastructure/config"
)

func TestBuildEventService_Success(t *testing.T) {
	original := loadConfig
	t.Cleanup(func() { loadConfig = original })

	loadConfig = func(_ context.Context) (*config.Config, error) {
		return &config.Config{
			DiscordWebhookURL: "https://discord.com/api/webhooks/123/abc",
		}, nil
	}

	svc, err := BuildEventService(context.Background())

	require.NoError(t, err)
	assert.NotNil(t, svc)
}

func TestBuildEventService_ConfigError(t *testing.T) {
	original := loadConfig
	t.Cleanup(func() { loadConfig = original })

	loadConfig = func(_ context.Context) (*config.Config, error) {
		return nil, errors.New("config error")
	}

	svc, err := BuildEventService(context.Background())

	require.Error(t, err)
	assert.Nil(t, svc)
	assert.Contains(t, err.Error(), "failed to load config")
}
