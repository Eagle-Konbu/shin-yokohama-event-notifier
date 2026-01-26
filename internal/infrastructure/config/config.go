package config

import (
	"context"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
)

type Config struct {
	DiscordWebhookURL string
}

// SecretsManagerClient defines the interface for Secrets Manager operations.
type SecretsManagerClient interface {
	GetSecretValue(ctx context.Context, params *secretsmanager.GetSecretValueInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.GetSecretValueOutput, error)
}

// LoadConfig loads configuration from AWS Secrets Manager.
func LoadConfig(ctx context.Context) (*Config, error) {
	return LoadConfigWithClient(ctx, nil)
}

// LoadConfigWithClient loads configuration using the provided Secrets Manager client.
// If client is nil, a default client is created.
func LoadConfigWithClient(ctx context.Context, client SecretsManagerClient) (*Config, error) {
	secretARN := os.Getenv("SECRET_ARN")
	if secretARN == "" {
		return nil, fmt.Errorf("SECRET_ARN environment variable is required")
	}

	if client == nil {
		cfg, err := config.LoadDefaultConfig(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to load AWS config: %w", err)
		}
		client = secretsmanager.NewFromConfig(cfg)
	}

	result, err := client.GetSecretValue(ctx, &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(secretARN),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get secret value: %w", err)
	}

	if result.SecretString == nil {
		return nil, fmt.Errorf("secret value is empty")
	}

	return &Config{
		DiscordWebhookURL: *result.SecretString,
	}, nil
}
