package lambda

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/Eagle-Konbu/shin-yokohama-event-notifier/internal/application/service"
	"github.com/Eagle-Konbu/shin-yokohama-event-notifier/internal/domain/notification"
)

type MockNotificationSender struct {
	mock.Mock
}

func (m *MockNotificationSender) Send(ctx context.Context, notif *notification.Notification) error {
	args := m.Called(ctx, notif)
	return args.Error(0)
}

func TestNewHandler(t *testing.T) {
	mockSender := new(MockNotificationSender)
	svc := service.NewEventNotificationService(mockSender)
	handler := NewHandler(svc)

	require.NotNil(t, handler)
	assert.NotNil(t, handler.eventService)
}

func TestHandler_HandleRequest_Success(t *testing.T) {
	mockSender := new(MockNotificationSender)
	svc := service.NewEventNotificationService(mockSender)
	handler := NewHandler(svc)

	ctx := context.Background()
	event := json.RawMessage(`{"key": "value", "source": "test"}`)

	mockSender.On("Send", ctx, mock.Anything).Return(nil)

	err := handler.HandleRequest(ctx, event)

	require.NoError(t, err)
	mockSender.AssertExpectations(t)
}

func TestHandler_HandleRequest_ServiceError(t *testing.T) {
	mockSender := new(MockNotificationSender)
	svc := service.NewEventNotificationService(mockSender)
	handler := NewHandler(svc)

	ctx := context.Background()
	event := json.RawMessage(`{"key": "value"}`)
	expectedErr := errors.New("sender error")

	mockSender.On("Send", ctx, mock.Anything).Return(expectedErr)

	err := handler.HandleRequest(ctx, event)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to process event")
	assert.ErrorIs(t, err, expectedErr)
	mockSender.AssertExpectations(t)
}

func TestHandler_HandleRequest_EmptyEvent(t *testing.T) {
	mockSender := new(MockNotificationSender)
	svc := service.NewEventNotificationService(mockSender)
	handler := NewHandler(svc)

	ctx := context.Background()
	event := json.RawMessage(`{}`)

	mockSender.On("Send", ctx, mock.Anything).Return(nil)

	err := handler.HandleRequest(ctx, event)

	require.NoError(t, err)
	mockSender.AssertExpectations(t)
}

func TestHandler_HandleRequest_ComplexEventBridgeJSON(t *testing.T) {
	mockSender := new(MockNotificationSender)
	svc := service.NewEventNotificationService(mockSender)
	handler := NewHandler(svc)

	ctx := context.Background()
	event := json.RawMessage(`{
		"version": "0",
		"id": "53dc4d37-cffa-4f76-80c9-8b7d4a4d2eaa",
		"detail-type": "Scheduled Event",
		"source": "aws.events",
		"account": "123456789012",
		"time": "2024-01-01T00:00:00Z",
		"region": "us-east-1",
		"resources": ["arn:aws:events:us-east-1:123456789012:rule/my-schedule"],
		"detail": {}
	}`)

	var capturedNotif *notification.Notification
	mockSender.On("Send", ctx, mock.Anything).Run(func(args mock.Arguments) {
		capturedNotif = args.Get(1).(*notification.Notification)
	}).Return(nil)

	err := handler.HandleRequest(ctx, event)

	require.NoError(t, err)
	require.NotNil(t, capturedNotif)
	assert.Contains(t, capturedNotif.Description(), "Scheduled Event")
}

func TestHandler_HandleRequest_ContextPropagation(t *testing.T) {
	mockSender := new(MockNotificationSender)
	svc := service.NewEventNotificationService(mockSender)
	handler := NewHandler(svc)

	type contextKey string
	const testKey contextKey = "testKey"
	ctx := context.WithValue(context.Background(), testKey, "testValue")
	event := json.RawMessage(`{"test": "data"}`)

	var capturedCtx context.Context
	mockSender.On("Send", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		capturedCtx = args.Get(0).(context.Context)
	}).Return(nil)

	err := handler.HandleRequest(ctx, event)

	require.NoError(t, err)
	require.NotNil(t, capturedCtx)
	assert.Equal(t, "testValue", capturedCtx.Value(testKey))
}
