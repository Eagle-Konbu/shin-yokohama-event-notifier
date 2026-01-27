package lambda

import (
	"context"
	"fmt"

	"github.com/Eagle-Konbu/shin-yokohama-event-notifier/internal/application/service"
)

type Handler struct {
	eventService *service.EventNotificationService
}

func NewHandler(eventService *service.EventNotificationService) *Handler {
	return &Handler{
		eventService: eventService,
	}
}

func (h *Handler) HandleRequest(ctx context.Context) error {
	if err := h.eventService.NotifyTodayEvents(ctx); err != nil {
		return fmt.Errorf("failed to notify today events: %w", err)
	}

	return nil
}
