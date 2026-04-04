package lambda

import (
	"context"
	"fmt"

	"github.com/Eagle-Konbu/shin-yokohama-event-notifier/internal/application/service"
)

type WeeklyHandler struct {
	eventService *service.EventNotificationService
}

func NewWeeklyHandler(eventService *service.EventNotificationService) *WeeklyHandler {
	return &WeeklyHandler{
		eventService: eventService,
	}
}

func (h *WeeklyHandler) HandleRequest(ctx context.Context) error {
	if err := h.eventService.NotifyWeeklyEvents(ctx); err != nil {
		return fmt.Errorf("failed to notify weekly events: %w", err)
	}

	return nil
}
