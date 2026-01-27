package ports

import (
	"context"

	"github.com/Eagle-Konbu/shin-yokohama-event-notifier/internal/domain/event"
)

type EventFetcher interface {
	FetchEvents(ctx context.Context) ([]event.Event, error)
}
