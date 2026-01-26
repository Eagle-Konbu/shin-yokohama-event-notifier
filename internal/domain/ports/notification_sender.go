package ports

import (
	"context"

	"github.com/Eagle-Konbu/shin-yokohama-event-notifier/internal/domain/notification"
)

type NotificationSender interface {
	Send(ctx context.Context, notif *notification.Notification) error
}
