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
)

// RoundTripFunc is a custom http.RoundTripper for testing
type RoundTripFunc func(req *http.Request) (*http.Response, error)

// RoundTrip implements the http.RoundTripper interface
func (f RoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

// newTestClient creates a WebhookClient with a mock transport
func newTestClient(fn RoundTripFunc) *WebhookClient {
	return &WebhookClient{
		httpClient: &http.Client{
			Transport: fn,
		},
	}
}

func TestWebhookClient_Execute_Success(t *testing.T) {
	webhookURL := "https://discord.com/api/webhooks/123/abc"
	payload := &WebhookPayload{
		Embeds: []Embed{
			{Title: "Test", Description: "Test Description"},
		},
	}

	mockTransport := RoundTripFunc(func(req *http.Request) (*http.Response, error) {
		// Verify request method
		assert.Equal(t, http.MethodPost, req.Method)

		// Verify URL
		assert.Equal(t, webhookURL, req.URL.String())

		// Verify Content-Type header
		assert.Equal(t, "application/json", req.Header.Get("Content-Type"))

		// Verify body can be read
		body, err := io.ReadAll(req.Body)
		require.NoError(t, err)
		assert.Contains(t, string(body), "Test")

		// Return success response
		return &http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(bytes.NewBuffer([]byte(""))),
		}, nil
	})

	client := newTestClient(mockTransport)
	err := client.Execute(context.Background(), webhookURL, payload)

	require.NoError(t, err)
}

func TestWebhookClient_Execute_NoContent(t *testing.T) {
	mockTransport := RoundTripFunc(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: 204,
			Body:       io.NopCloser(bytes.NewBuffer([]byte(""))),
		}, nil
	})

	client := newTestClient(mockTransport)
	err := client.Execute(context.Background(), "https://discord.com/api/webhooks/123/abc", &WebhookPayload{})

	require.NoError(t, err)
}

func TestWebhookClient_Execute_BadRequest(t *testing.T) {
	mockTransport := RoundTripFunc(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: 400,
			Body:       io.NopCloser(bytes.NewBuffer([]byte("Bad Request"))),
		}, nil
	})

	client := newTestClient(mockTransport)
	err := client.Execute(context.Background(), "https://discord.com/api/webhooks/123/abc", &WebhookPayload{})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "webhook returned non-success status: 400")
}

func TestWebhookClient_Execute_RateLimited(t *testing.T) {
	mockTransport := RoundTripFunc(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: 429,
			Body:       io.NopCloser(bytes.NewBuffer([]byte("Rate Limited"))),
		}, nil
	})

	client := newTestClient(mockTransport)
	err := client.Execute(context.Background(), "https://discord.com/api/webhooks/123/abc", &WebhookPayload{})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "webhook returned non-success status: 429")
}

func TestWebhookClient_Execute_NetworkError(t *testing.T) {
	expectedErr := errors.New("connection refused")
	mockTransport := RoundTripFunc(func(req *http.Request) (*http.Response, error) {
		return nil, expectedErr
	})

	client := newTestClient(mockTransport)
	err := client.Execute(context.Background(), "https://discord.com/api/webhooks/123/abc", &WebhookPayload{})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to execute webhook request")
	assert.ErrorIs(t, err, expectedErr)
}

func TestWebhookClient_Execute_ContextCancellation(t *testing.T) {
	mockTransport := RoundTripFunc(func(req *http.Request) (*http.Response, error) {
		// Check if request context is cancelled
		select {
		case <-req.Context().Done():
			return nil, req.Context().Err()
		default:
			return &http.Response{StatusCode: 200, Body: io.NopCloser(bytes.NewBuffer([]byte("")))}, nil
		}
	})

	client := newTestClient(mockTransport)
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	err := client.Execute(ctx, "https://discord.com/api/webhooks/123/abc", &WebhookPayload{})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to execute webhook request")
}

func TestWebhookClient_Execute_RequestConstruction(t *testing.T) {
	webhookURL := "https://discord.com/api/webhooks/123/abc"
	payload := &WebhookPayload{
		Content: "Test Content",
		Embeds: []Embed{
			{Title: "Test Title"},
		},
	}

	var capturedReq *http.Request
	mockTransport := RoundTripFunc(func(req *http.Request) (*http.Response, error) {
		capturedReq = req
		return &http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(bytes.NewBuffer([]byte(""))),
		}, nil
	})

	client := newTestClient(mockTransport)
	ctx := context.Background()
	err := client.Execute(ctx, webhookURL, payload)

	require.NoError(t, err)
	require.NotNil(t, capturedReq)

	// Verify URL
	assert.Equal(t, webhookURL, capturedReq.URL.String())

	// Verify method
	assert.Equal(t, http.MethodPost, capturedReq.Method)

	// Verify Content-Type header
	assert.Equal(t, "application/json", capturedReq.Header.Get("Content-Type"))

	// Verify body contains payload data
	body, err := io.ReadAll(capturedReq.Body)
	require.NoError(t, err)
	assert.Contains(t, string(body), "Test Content")
	assert.Contains(t, string(body), "Test Title")
}

func TestNewWebhookClient(t *testing.T) {
	client := NewWebhookClient()

	require.NotNil(t, client)
	require.NotNil(t, client.httpClient)
	assert.NotZero(t, client.httpClient.Timeout)
}
