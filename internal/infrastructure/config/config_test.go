package config

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockSecretsManagerClient struct {
	getSecretValueFunc func(ctx context.Context, params *secretsmanager.GetSecretValueInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.GetSecretValueOutput, error)
}

func (m *mockSecretsManagerClient) GetSecretValue(ctx context.Context, params *secretsmanager.GetSecretValueInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.GetSecretValueOutput, error) {
	return m.getSecretValueFunc(ctx, params, optFns...)
}

func TestLoadConfig_Success(t *testing.T) {
	t.Setenv("SECRET_ARN", "arn:aws:secretsmanager:ap-northeast-1:123456789012:secret:test-secret")

	mockClient := &mockSecretsManagerClient{
		getSecretValueFunc: func(ctx context.Context, params *secretsmanager.GetSecretValueInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.GetSecretValueOutput, error) {
			return &secretsmanager.GetSecretValueOutput{
				SecretString: aws.String("https://discord.com/api/webhooks/123/abc"),
			}, nil
		},
	}

	cfg, err := LoadConfigWithClient(context.Background(), mockClient)

	require.NoError(t, err)
	require.NotNil(t, cfg)
	assert.Equal(t, "https://discord.com/api/webhooks/123/abc", cfg.DiscordWebhookURL)
}

func TestLoadConfig_MissingEnvVar(t *testing.T) {
	t.Setenv("SECRET_ARN", "")

	cfg, err := LoadConfigWithClient(context.Background(), nil)

	require.Error(t, err)
	assert.Nil(t, cfg)
	assert.Contains(t, err.Error(), "SECRET_ARN")
	assert.Contains(t, err.Error(), "required")
}

func TestLoadConfig_SecretsManagerError(t *testing.T) {
	t.Setenv("SECRET_ARN", "arn:aws:secretsmanager:ap-northeast-1:123456789012:secret:test-secret")

	mockClient := &mockSecretsManagerClient{
		getSecretValueFunc: func(ctx context.Context, params *secretsmanager.GetSecretValueInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.GetSecretValueOutput, error) {
			return nil, errors.New("access denied")
		},
	}

	cfg, err := LoadConfigWithClient(context.Background(), mockClient)

	require.Error(t, err)
	assert.Nil(t, cfg)
	assert.Contains(t, err.Error(), "failed to get secret value")
}

func TestLoadConfig_EmptySecretValue(t *testing.T) {
	t.Setenv("SECRET_ARN", "arn:aws:secretsmanager:ap-northeast-1:123456789012:secret:test-secret")

	mockClient := &mockSecretsManagerClient{
		getSecretValueFunc: func(ctx context.Context, params *secretsmanager.GetSecretValueInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.GetSecretValueOutput, error) {
			return &secretsmanager.GetSecretValueOutput{
				SecretString: nil,
			}, nil
		},
	}

	cfg, err := LoadConfigWithClient(context.Background(), mockClient)

	require.Error(t, err)
	assert.Nil(t, cfg)
	assert.Contains(t, err.Error(), "secret value is empty")
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
			t.Setenv("SECRET_ARN", "arn:aws:secretsmanager:ap-northeast-1:123456789012:secret:test-secret")

			mockClient := &mockSecretsManagerClient{
				getSecretValueFunc: func(ctx context.Context, params *secretsmanager.GetSecretValueInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.GetSecretValueOutput, error) {
					return &secretsmanager.GetSecretValueOutput{
						SecretString: aws.String(tc.webhookURL),
					}, nil
				},
			}

			cfg, err := LoadConfigWithClient(context.Background(), mockClient)

			require.NoError(t, err)
			assert.Equal(t, tc.webhookURL, cfg.DiscordWebhookURL)
		})
	}
}

func TestLoadConfig_AWSConfigError(t *testing.T) {
	t.Setenv("SECRET_ARN", "arn:aws:secretsmanager:ap-northeast-1:123456789012:secret:test-secret")

	originalLoadAWSConfig := loadAWSConfig
	t.Cleanup(func() { loadAWSConfig = originalLoadAWSConfig })

	loadAWSConfig = func(ctx context.Context, optFns ...func(*config.LoadOptions) error) (aws.Config, error) {
		return aws.Config{}, errors.New("failed to load config")
	}

	cfg, err := LoadConfigWithClient(context.Background(), nil)

	require.Error(t, err)
	assert.Nil(t, cfg)
	assert.Contains(t, err.Error(), "failed to load AWS config")
}

func TestLoadConfig_WithNilClient(t *testing.T) {
	t.Setenv("SECRET_ARN", "arn:aws:secretsmanager:ap-northeast-1:123456789012:secret:test-secret")

	originalLoadAWSConfig := loadAWSConfig
	originalNewSecretsManagerClient := newSecretsManagerClient
	t.Cleanup(func() {
		loadAWSConfig = originalLoadAWSConfig
		newSecretsManagerClient = originalNewSecretsManagerClient
	})

	loadAWSConfig = func(ctx context.Context, optFns ...func(*config.LoadOptions) error) (aws.Config, error) {
		return aws.Config{}, nil
	}

	newSecretsManagerClient = func(cfg aws.Config) SecretsManagerClient {
		return &mockSecretsManagerClient{
			getSecretValueFunc: func(ctx context.Context, params *secretsmanager.GetSecretValueInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.GetSecretValueOutput, error) {
				return &secretsmanager.GetSecretValueOutput{
					SecretString: aws.String("https://discord.com/api/webhooks/123/abc"),
				}, nil
			},
		}
	}

	cfg, err := LoadConfigWithClient(context.Background(), nil)

	require.NoError(t, err)
	require.NotNil(t, cfg)
	assert.Equal(t, "https://discord.com/api/webhooks/123/abc", cfg.DiscordWebhookURL)
}

func TestLoadConfig_CallsLoadConfigWithClient(t *testing.T) {
	t.Setenv("SECRET_ARN", "arn:aws:secretsmanager:ap-northeast-1:123456789012:secret:test-secret")

	originalLoadAWSConfig := loadAWSConfig
	originalNewSecretsManagerClient := newSecretsManagerClient
	t.Cleanup(func() {
		loadAWSConfig = originalLoadAWSConfig
		newSecretsManagerClient = originalNewSecretsManagerClient
	})

	loadAWSConfig = func(ctx context.Context, optFns ...func(*config.LoadOptions) error) (aws.Config, error) {
		return aws.Config{}, nil
	}

	newSecretsManagerClient = func(cfg aws.Config) SecretsManagerClient {
		return &mockSecretsManagerClient{
			getSecretValueFunc: func(ctx context.Context, params *secretsmanager.GetSecretValueInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.GetSecretValueOutput, error) {
				return &secretsmanager.GetSecretValueOutput{
					SecretString: aws.String("https://discord.com/api/webhooks/123/abc"),
				}, nil
			},
		}
	}

	cfg, err := LoadConfig(context.Background())

	require.NoError(t, err)
	require.NotNil(t, cfg)
	assert.Equal(t, "https://discord.com/api/webhooks/123/abc", cfg.DiscordWebhookURL)
}
