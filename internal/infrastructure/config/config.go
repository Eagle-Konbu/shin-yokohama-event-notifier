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

type SecretsManagerClient interface {
	GetSecretValue(ctx context.Context, params *secretsmanager.GetSecretValueInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.GetSecretValueOutput, error)
}

var loadAWSConfig = config.LoadDefaultConfig

var newSecretsManagerClient = func(cfg aws.Config) SecretsManagerClient {
	return secretsmanager.NewFromConfig(cfg)
}

func LoadConfig(ctx context.Context) (*Config, error) {
	return LoadConfigWithClient(ctx, nil)
}

func LoadConfigWithClient(ctx context.Context, client SecretsManagerClient) (*Config, error) {
	secretARN := os.Getenv("SECRET_ARN")
	if secretARN == "" {
		return nil, fmt.Errorf("SECRET_ARN environment variable is required")
	}

	if client == nil {
		cfg, err := loadAWSConfig(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to load AWS config: %w", err)
		}
		client = newSecretsManagerClient(cfg)
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
