package service

import (
	"context"
	"errors"
	"testing"

	"github.com/Eagle-Konbu/shin-yokohama-event-notifier/internal/domain/notification"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestNewEventNotificationService(t *testing.T) {
	mockSender := new(MockNotificationSender)
	service := NewEventNotificationService(mockSender)

	require.NotNil(t, service)
	assert.NotNil(t, service.notificationSender)
}

func TestEventNotificationService_ProcessScheduledEvent_Success(t *testing.T) {
	mockSender := new(MockNotificationSender)
	service := NewEventNotificationService(mockSender)
	ctx := context.Background()
	eventData := "test-event-data"

	mockSender.On("Send", ctx, mock.MatchedBy(func(n *notification.Notification) bool {
		return n.Title() == "Scheduled Event Notification" &&
			n.Color() == notification.ColorBlue &&
			len(n.Fields()) == 2 &&
			n.Fields()[0].Name == "Source" &&
			n.Fields()[0].Value == "EventBridge" &&
			n.Fields()[1].Name == "Type" &&
			n.Fields()[1].Value == "Scheduled"
	})).Return(nil)

	err := service.ProcessScheduledEvent(ctx, eventData)

	require.NoError(t, err)
	mockSender.AssertExpectations(t)
	mockSender.AssertCalled(t, "Send", ctx, mock.Anything)
}

func TestEventNotificationService_ProcessScheduledEvent_DescriptionContainsEventData(t *testing.T) {
	mockSender := new(MockNotificationSender)
	service := NewEventNotificationService(mockSender)
	ctx := context.Background()
	eventData := "important-event-12345"

	var capturedNotif *notification.Notification
	mockSender.On("Send", ctx, mock.Anything).Run(func(args mock.Arguments) {
		capturedNotif = args.Get(1).(*notification.Notification)
	}).Return(nil)

	err := service.ProcessScheduledEvent(ctx, eventData)

	require.NoError(t, err)
	require.NotNil(t, capturedNotif)
	assert.Contains(t, capturedNotif.Description(), eventData)
	assert.Contains(t, capturedNotif.Description(), "Event triggered:")
}

func TestEventNotificationService_ProcessScheduledEvent_SenderError(t *testing.T) {
	mockSender := new(MockNotificationSender)
	service := NewEventNotificationService(mockSender)
	ctx := context.Background()
	eventData := "test-event"
	expectedErr := errors.New("sender error")

	mockSender.On("Send", ctx, mock.Anything).Return(expectedErr)

	err := service.ProcessScheduledEvent(ctx, eventData)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to send notification")
	assert.ErrorIs(t, err, expectedErr)
	mockSender.AssertExpectations(t)
}

func TestEventNotificationService_ProcessScheduledEvent_ContextPropagation(t *testing.T) {
	mockSender := new(MockNotificationSender)
	service := NewEventNotificationService(mockSender)

	type contextKey string
	const testKey contextKey = "testKey"
	ctx := context.WithValue(context.Background(), testKey, "testValue")
	eventData := "test-event"

	var capturedCtx context.Context
	mockSender.On("Send", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
		capturedCtx = args.Get(0).(context.Context)
	}).Return(nil)

	err := service.ProcessScheduledEvent(ctx, eventData)

	require.NoError(t, err)
	require.NotNil(t, capturedCtx)
	assert.Equal(t, "testValue", capturedCtx.Value(testKey))
}
