package discord

import (
	"context"
	"fmt"

	"github.com/Eagle-Konbu/shin-yokohama-event-notifier/internal/domain/notification"
	"github.com/Eagle-Konbu/shin-yokohama-event-notifier/internal/domain/ports"
)

// WebhookAdapter implements the NotificationSender port using Discord webhooks
type WebhookAdapter struct {
	client     *WebhookClient
	webhookURL string
}

// NewWebhookAdapter creates a new Discord webhook adapter
func NewWebhookAdapter(webhookURL string) ports.NotificationSender {
	return &WebhookAdapter{
		client:     NewWebhookClient(),
		webhookURL: webhookURL,
	}
}

// Send implements the NotificationSender interface
func (a *WebhookAdapter) Send(ctx context.Context, notif *notification.Notification) error {
	embed := mapNotificationToEmbed(notif)

	payload := &WebhookPayload{
		Embeds: []Embed{embed},
	}

	if err := a.client.Execute(ctx, a.webhookURL, payload); err != nil {
		return fmt.Errorf("failed to send Discord webhook: %w", err)
	}

	return nil
}
