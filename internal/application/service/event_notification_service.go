package service

import (
	"context"
	"fmt"

	"github.com/Eagle-Konbu/shin-yokohama-event-notifier/internal/domain/notification"
	"github.com/Eagle-Konbu/shin-yokohama-event-notifier/internal/domain/ports"
)

// EventNotificationService handles the application logic for processing events
type EventNotificationService struct {
	notificationSender ports.NotificationSender
}

// NewEventNotificationService creates a new event notification service
func NewEventNotificationService(sender ports.NotificationSender) *EventNotificationService {
	return &EventNotificationService{
		notificationSender: sender,
	}
}

// ProcessScheduledEvent processes a scheduled event and sends a notification
func (s *EventNotificationService) ProcessScheduledEvent(ctx context.Context, eventData string) error {
	// Business logic: Parse event, create notification
	notif := notification.NewNotification(
		"Scheduled Event Notification",
		fmt.Sprintf("Event triggered: %s", eventData),
		notification.ColorBlue,
	)

	notif.AddField("Source", "EventBridge", true)
	notif.AddField("Type", "Scheduled", true)

	// Send via port
	if err := s.notificationSender.Send(ctx, notif); err != nil {
		return fmt.Errorf("failed to send notification: %w", err)
	}

	return nil
}
