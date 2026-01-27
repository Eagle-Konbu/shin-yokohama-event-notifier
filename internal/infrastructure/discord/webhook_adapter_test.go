package discord

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Eagle-Konbu/shin-yokohama-event-notifier/internal/domain/notification"
)

func newTestWebhookAdapter(fn RoundTripFunc, webhookURL string) *WebhookAdapter {
	return &WebhookAdapter{
		client:     newTestClient(fn),
		webhookURL: webhookURL,
	}
}

func TestNewWebhookAdapter(t *testing.T) {
	webhookURL := "https://discord.com/api/webhooks/123/abc"
	adapter := NewWebhookAdapter(webhookURL)

	require.NotNil(t, adapter)
}

func TestWebhookAdapter_Send_Success(t *testing.T) {
	webhookURL := "https://discord.com/api/webhooks/123/abc"

	mockTransport := RoundTripFunc(func(req *http.Request) (*http.Response, error) {
		assert.Equal(t, webhookURL, req.URL.String())
		assert.Equal(t, http.MethodPost, req.Method)
		assert.Equal(t, "application/json", req.Header.Get("Content-Type"))

		body, err := io.ReadAll(req.Body)
		require.NoError(t, err)
		assert.Contains(t, string(body), "Test Title")
		assert.Contains(t, string(body), "Test Description")

		return &http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(bytes.NewBuffer([]byte(""))),
		}, nil
	})

	adapter := newTestWebhookAdapter(mockTransport, webhookURL)
	ctx := context.Background()
	notif := notification.NewNotification("Test Title", "Test Description", notification.ColorGreen)
	notif.AddField("Field1", "Value1", true)

	err := adapter.Send(ctx, notif)

	require.NoError(t, err)
}

func TestWebhookAdapter_Send_ClientError(t *testing.T) {
	webhookURL := "https://discord.com/api/webhooks/123/abc"
	expectedErr := errors.New("connection refused")

	mockTransport := RoundTripFunc(func(req *http.Request) (*http.Response, error) {
		return nil, expectedErr
	})

	adapter := newTestWebhookAdapter(mockTransport, webhookURL)
	ctx := context.Background()
	notif := notification.NewNotification("Test", "Test", notification.ColorRed)

	err := adapter.Send(ctx, notif)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to send Discord webhook")
	assert.ErrorIs(t, err, expectedErr)
}

func TestWebhookAdapter_Send_NonSuccessStatus(t *testing.T) {
	webhookURL := "https://discord.com/api/webhooks/123/abc"

	mockTransport := RoundTripFunc(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: 400,
			Body:       io.NopCloser(bytes.NewBuffer([]byte("Bad Request"))),
		}, nil
	})

	adapter := newTestWebhookAdapter(mockTransport, webhookURL)
	ctx := context.Background()
	notif := notification.NewNotification("Test", "Test", notification.ColorRed)

	err := adapter.Send(ctx, notif)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to send Discord webhook")
	assert.Contains(t, err.Error(), "400")
}

func TestWebhookAdapter_Send_EmbedMapping(t *testing.T) {
	webhookURL := "https://discord.com/api/webhooks/123/abc"

	var capturedBody []byte
	mockTransport := RoundTripFunc(func(req *http.Request) (*http.Response, error) {
		var err error
		capturedBody, err = io.ReadAll(req.Body)
		require.NoError(t, err)
		return &http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(bytes.NewBuffer([]byte(""))),
		}, nil
	})

	adapter := newTestWebhookAdapter(mockTransport, webhookURL)
	ctx := context.Background()
	notif := notification.NewNotification("Title", "Description", notification.ColorYellow)
	notif.AddField("Field1", "Value1", true)
	notif.AddField("Field2", "Value2", false)
	notif.AddField("Field3", "Value3", true)

	err := adapter.Send(ctx, notif)

	require.NoError(t, err)
	bodyStr := string(capturedBody)
	assert.Contains(t, bodyStr, "Title")
	assert.Contains(t, bodyStr, "Description")
	assert.Contains(t, bodyStr, "Field1")
	assert.Contains(t, bodyStr, "Value1")
	assert.Contains(t, bodyStr, "Field2")
	assert.Contains(t, bodyStr, "Value2")
	assert.Contains(t, bodyStr, "Field3")
	assert.Contains(t, bodyStr, "Value3")
}

func TestWebhookAdapter_Send_ContextPropagation(t *testing.T) {
	webhookURL := "https://discord.com/api/webhooks/123/abc"

	type contextKey string
	const testKey contextKey = "testKey"

	var capturedCtx context.Context
	mockTransport := RoundTripFunc(func(req *http.Request) (*http.Response, error) {
		capturedCtx = req.Context()
		return &http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(bytes.NewBuffer([]byte(""))),
		}, nil
	})

	adapter := newTestWebhookAdapter(mockTransport, webhookURL)
	ctx := context.WithValue(context.Background(), testKey, "testValue")
	notif := notification.NewNotification("Test", "Test", notification.ColorGreen)

	err := adapter.Send(ctx, notif)

	require.NoError(t, err)
	require.NotNil(t, capturedCtx)
	assert.Equal(t, "testValue", capturedCtx.Value(testKey))
}

func TestWebhookAdapter_Send_ContextCancellation(t *testing.T) {
	webhookURL := "https://discord.com/api/webhooks/123/abc"

	mockTransport := RoundTripFunc(func(req *http.Request) (*http.Response, error) {
		return nil, req.Context().Err()
	})

	adapter := newTestWebhookAdapter(mockTransport, webhookURL)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	notif := notification.NewNotification("Test", "Test", notification.ColorGreen)

	err := adapter.Send(ctx, notif)

	require.Error(t, err)
	assert.ErrorIs(t, err, context.Canceled)
}
