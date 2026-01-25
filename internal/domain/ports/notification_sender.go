package ports

import (
	"context"

	"github.com/Eagle-Konbu/shin-yokohama-event-notifier/internal/domain/notification"
)

// NotificationSender is the outbound port for sending notifications
// This interface belongs to the domain layer and will be implemented by infrastructure adapters
type NotificationSender interface {
	// Send sends a notification and returns an error if the operation fails
	Send(ctx context.Context, notif *notification.Notification) error
}
