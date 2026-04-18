package ports

import (
	"context"
	"time"

	"github.com/Eagle-Konbu/shin-yokohama-event-notifier/internal/domain/event"
)

type EventFetcher interface {
	FetchEvents(ctx context.Context, from, to time.Time) ([]event.Event, error)
	VenueID() event.VenueID
}
