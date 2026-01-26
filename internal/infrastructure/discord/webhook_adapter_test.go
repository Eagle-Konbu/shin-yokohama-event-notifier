package discord

import (
	"context"
	"errors"
	"testing"

	"github.com/Eagle-Konbu/shin-yokohama-event-notifier/internal/domain/notification"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockWebhookClient struct {
	mock.Mock
}

func (m *MockWebhookClient) Execute(ctx context.Context, webhookURL string, payload *WebhookPayload) error {
	args := m.Called(ctx, webhookURL, payload)
	return args.Error(0)
}

type TestableWebhookAdapter struct {
	client     *MockWebhookClient
	webhookURL string
}

func (a *TestableWebhookAdapter) Send(ctx context.Context, notif *notification.Notification) error {
	embed := mapNotificationToEmbed(notif)
	payload := &WebhookPayload{
		Embeds: []Embed{embed},
	}
	if err := a.client.Execute(ctx, a.webhookURL, payload); err != nil {
		return err
	}
	return nil
}

func TestNewWebhookAdapter(t *testing.T) {
	webhookURL := "https://discord.com/api/webhooks/123/abc"
	adapter := NewWebhookAdapter(webhookURL)

	require.NotNil(t, adapter)
}

func TestWebhookAdapter_Send_Success(t *testing.T) {
	mockClient := new(MockWebhookClient)
	webhookURL := "https://discord.com/api/webhooks/123/abc"
	adapter := &TestableWebhookAdapter{
		client:     mockClient,
		webhookURL: webhookURL,
	}

	ctx := context.Background()
	notif := notification.NewNotification("Test Title", "Test Description", notification.ColorBlue)
	notif.AddField("Field1", "Value1", true)

	mockClient.On("Execute", ctx, webhookURL, mock.MatchedBy(func(p *WebhookPayload) bool {
		return len(p.Embeds) == 1 &&
			p.Embeds[0].Title == "Test Title" &&
			p.Embeds[0].Description == "Test Description" &&
			p.Embeds[0].Color == int(notification.ColorBlue) &&
			len(p.Embeds[0].Fields) == 1
	})).Return(nil)

	err := adapter.Send(ctx, notif)

	require.NoError(t, err)
	mockClient.AssertExpectations(t)
}

func TestWebhookAdapter_Send_ClientError(t *testing.T) {
	mockClient := new(MockWebhookClient)
	webhookURL := "https://discord.com/api/webhooks/123/abc"
	adapter := &TestableWebhookAdapter{
		client:     mockClient,
		webhookURL: webhookURL,
	}

	ctx := context.Background()
	notif := notification.NewNotification("Test", "Test", notification.ColorRed)
	expectedErr := errors.New("client error")

	mockClient.On("Execute", ctx, webhookURL, mock.Anything).Return(expectedErr)

	err := adapter.Send(ctx, notif)

	require.Error(t, err)
	assert.ErrorIs(t, err, expectedErr)
	mockClient.AssertExpectations(t)
}

func TestWebhookAdapter_Send_EmbedMapping(t *testing.T) {
	mockClient := new(MockWebhookClient)
	webhookURL := "https://discord.com/api/webhooks/123/abc"
	adapter := &TestableWebhookAdapter{
		client:     mockClient,
		webhookURL: webhookURL,
	}

	ctx := context.Background()
	notif := notification.NewNotification("Title", "Description", notification.ColorPurple)
	notif.AddField("Field1", "Value1", true)
	notif.AddField("Field2", "Value2", false)
	notif.AddField("Field3", "Value3", true)

	var capturedPayload *WebhookPayload
	mockClient.On("Execute", ctx, webhookURL, mock.Anything).Run(func(args mock.Arguments) {
		capturedPayload = args.Get(2).(*WebhookPayload)
	}).Return(nil)

	err := adapter.Send(ctx, notif)

	require.NoError(t, err)
	require.NotNil(t, capturedPayload)
	require.Len(t, capturedPayload.Embeds, 1)

	embed := capturedPayload.Embeds[0]
	assert.Equal(t, "Title", embed.Title)
	assert.Equal(t, "Description", embed.Description)
	assert.Equal(t, int(notification.ColorPurple), embed.Color)
	require.Len(t, embed.Fields, 3)
	assert.Equal(t, "Field1", embed.Fields[0].Name)
	assert.Equal(t, "Value1", embed.Fields[0].Value)
	assert.True(t, embed.Fields[0].Inline)
	assert.Equal(t, "Field2", embed.Fields[1].Name)
	assert.False(t, embed.Fields[1].Inline)
}

func TestWebhookAdapter_Send_ContextPropagation(t *testing.T) {
	mockClient := new(MockWebhookClient)
	webhookURL := "https://discord.com/api/webhooks/123/abc"
	adapter := &TestableWebhookAdapter{
		client:     mockClient,
		webhookURL: webhookURL,
	}

	type contextKey string
	const testKey contextKey = "testKey"
	ctx := context.WithValue(context.Background(), testKey, "testValue")
	notif := notification.NewNotification("Test", "Test", notification.ColorGreen)

	var capturedCtx context.Context
	mockClient.On("Execute", mock.Anything, webhookURL, mock.Anything).Run(func(args mock.Arguments) {
		capturedCtx = args.Get(0).(context.Context)
	}).Return(nil)

	err := adapter.Send(ctx, notif)

	require.NoError(t, err)
	require.NotNil(t, capturedCtx)
	assert.Equal(t, "testValue", capturedCtx.Value(testKey))
}
