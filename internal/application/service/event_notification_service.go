package service

import (
	"context"
	"fmt"

	"github.com/Eagle-Konbu/shin-yokohama-event-notifier/internal/domain/notification"
	"github.com/Eagle-Konbu/shin-yokohama-event-notifier/internal/domain/ports"
)

type EventNotificationService struct {
	notificationSender ports.NotificationSender
}

func NewEventNotificationService(sender ports.NotificationSender) *EventNotificationService {
	return &EventNotificationService{
		notificationSender: sender,
	}
}

func (s *EventNotificationService) ProcessScheduledEvent(ctx context.Context, eventData string) error {
	notif := notification.NewNotification(
		"Scheduled Event Notification",
		fmt.Sprintf("Event triggered: %s", eventData),
		notification.ColorBlue,
	)

	notif.AddField("Source", "EventBridge", true)
	notif.AddField("Type", "Scheduled", true)

	if err := s.notificationSender.Send(ctx, notif); err != nil {
		return fmt.Errorf("failed to send notification: %w", err)
	}

	return nil
}
