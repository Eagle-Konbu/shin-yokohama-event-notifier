package ports

import (
	"context"
	"time"

	"github.com/Eagle-Konbu/shin-yokohama-event-notifier/internal/domain/event"
)

//go:generate mockgen -source=event_fetcher.go -destination=mock_ports/mock_event_fetcher.go -package=mock_ports
type EventFetcher interface {
	FetchEvents(ctx context.Context, from, to time.Time) ([]event.Event, error)
	VenueID() event.VenueID
}
