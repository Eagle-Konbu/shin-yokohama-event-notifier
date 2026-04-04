package lambda

import (
	"context"
	"fmt"

	"github.com/Eagle-Konbu/shin-yokohama-event-notifier/internal/application/service"
)

type DailyHandler struct {
	eventService *service.EventNotificationService
}

func NewDailyHandler(eventService *service.EventNotificationService) *DailyHandler {
	return &DailyHandler{
		eventService: eventService,
	}
}

func (h *DailyHandler) HandleRequest(ctx context.Context) error {
	if err := h.eventService.NotifyTodayEvents(ctx); err != nil {
		return fmt.Errorf("failed to notify today events: %w", err)
	}

	return nil
}
