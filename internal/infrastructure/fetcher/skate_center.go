package fetcher

import (
	"context"
	"errors"
	"time"

	"github.com/Eagle-Konbu/shin-yokohama-event-notifier/internal/domain/event"
	"github.com/Eagle-Konbu/shin-yokohama-event-notifier/internal/domain/ports"
)

type SkateCenterFetcher struct{}

func NewSkateCenterFetcher() ports.EventFetcher {
	return &SkateCenterFetcher{}
}

func (s *SkateCenterFetcher) FetchEvents(ctx context.Context) ([]event.Event, error) {
	time.Sleep(time.Second * 2)
	return nil, errors.New("not implemented")
}

func (s *SkateCenterFetcher) VenueID() event.VenueID {
	return event.VenueIDSkateCenter
}
