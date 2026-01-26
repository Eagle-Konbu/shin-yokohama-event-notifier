package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadConfig_Success(t *testing.T) {
	t.Setenv("DISCORD_WEBHOOK_URL", "https://discord.com/api/webhooks/123/abc")

	cfg, err := LoadConfig()

	require.NoError(t, err)
	require.NotNil(t, cfg)
	assert.Equal(t, "https://discord.com/api/webhooks/123/abc", cfg.DiscordWebhookURL)
}

func TestLoadConfig_MissingEnvVar(t *testing.T) {
	// Explicitly unset the environment variable (t.Setenv with empty string)
	t.Setenv("DISCORD_WEBHOOK_URL", "")

	cfg, err := LoadConfig()

	require.Error(t, err)
	assert.Nil(t, cfg)
	assert.Contains(t, err.Error(), "DISCORD_WEBHOOK_URL")
	assert.Contains(t, err.Error(), "required")
}

func TestLoadConfig_EmptyEnvVar(t *testing.T) {
	t.Setenv("DISCORD_WEBHOOK_URL", "")

	cfg, err := LoadConfig()

	require.Error(t, err)
	assert.Nil(t, cfg)
	assert.Contains(t, err.Error(), "DISCORD_WEBHOOK_URL")
}

func TestLoadConfig_ValidURL(t *testing.T) {
	testCases := []struct {
		name       string
		webhookURL string
	}{
		{
			name:       "standard webhook URL",
			webhookURL: "https://discord.com/api/webhooks/123456789/abcdefgh",
		},
		{
			name:       "webhook URL with query params",
			webhookURL: "https://discord.com/api/webhooks/123456789/abcdefgh?wait=true",
		},
		{
			name:       "discordapp.com domain",
			webhookURL: "https://discordapp.com/api/webhooks/123456789/abcdefgh",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("DISCORD_WEBHOOK_URL", tc.webhookURL)

			cfg, err := LoadConfig()

			require.NoError(t, err)
			assert.Equal(t, tc.webhookURL, cfg.DiscordWebhookURL)
		})
	}
}
