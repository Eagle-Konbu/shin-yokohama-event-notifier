package service

import (
	"context"

	"github.com/Eagle-Konbu/shin-yokohama-event-notifier/internal/domain/notification"
	"github.com/stretchr/testify/mock"
)

// MockNotificationSender is a manual mock for testing
type MockNotificationSender struct {
	mock.Mock
}

// Send implements the NotificationSender interface
func (m *MockNotificationSender) Send(ctx context.Context, notif *notification.Notification) error {
	args := m.Called(ctx, notif)
	return args.Error(0)
}
