package ports

import (
	"context"

	"github.com/Eagle-Konbu/shin-yokohama-event-notifier/internal/domain/notification"
)

//go:generate mockgen -source=notification_sender.go -destination=mock_ports/mock_notification_sender.go -package=mock_ports
type NotificationSender interface {
	Send(ctx context.Context, notif *notification.Notification) error
}
