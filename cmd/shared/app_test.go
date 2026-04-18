package shared

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
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

func TestBuildEventServiceWithClient_Success(t *testing.T) {
	t.Setenv("SECRET_ARN", "arn:aws:secretsmanager:ap-northeast-1:123456789012:secret:test-secret")

	mockClient := &mockSecretsManagerClient{
		getSecretValueFunc: func(_ context.Context, _ *secretsmanager.GetSecretValueInput, _ ...func(*secretsmanager.Options)) (*secretsmanager.GetSecretValueOutput, error) {
			return &secretsmanager.GetSecretValueOutput{
				SecretString: aws.String("https://discord.com/api/webhooks/123/abc"),
			}, nil
		},
	}

	svc, err := BuildEventServiceWithClient(context.Background(), mockClient)

	require.NoError(t, err)
	assert.NotNil(t, svc)
}

func TestBuildEventServiceWithClient_MissingSecretARN(t *testing.T) {
	t.Setenv("SECRET_ARN", "")

	svc, err := BuildEventServiceWithClient(context.Background(), nil)

	require.Error(t, err)
	assert.Nil(t, svc)
	assert.Contains(t, err.Error(), "failed to load config")
}

func TestBuildEventServiceWithClient_SecretsManagerError(t *testing.T) {
	t.Setenv("SECRET_ARN", "arn:aws:secretsmanager:ap-northeast-1:123456789012:secret:test-secret")

	mockClient := &mockSecretsManagerClient{
		getSecretValueFunc: func(_ context.Context, _ *secretsmanager.GetSecretValueInput, _ ...func(*secretsmanager.Options)) (*secretsmanager.GetSecretValueOutput, error) {
			return nil, errors.New("access denied")
		},
	}

	svc, err := BuildEventServiceWithClient(context.Background(), mockClient)

	require.Error(t, err)
	assert.Nil(t, svc)
	assert.Contains(t, err.Error(), "failed to load config")
}
