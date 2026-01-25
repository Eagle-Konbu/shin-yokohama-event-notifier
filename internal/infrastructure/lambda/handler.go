package lambda

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Eagle-Konbu/shin-yokohama-event-notifier/internal/application/service"
)

// Handler is the Lambda inbound adapter
type Handler struct {
	eventService *service.EventNotificationService
}

// NewHandler creates a new Lambda handler
func NewHandler(eventService *service.EventNotificationService) *Handler {
	return &Handler{
		eventService: eventService,
	}
}

// HandleRequest processes Lambda invocations
func (h *Handler) HandleRequest(ctx context.Context, event json.RawMessage) error {
	eventData := string(event)

	if err := h.eventService.ProcessScheduledEvent(ctx, eventData); err != nil {
		return fmt.Errorf("failed to process event: %w", err)
	}

	return nil
}
