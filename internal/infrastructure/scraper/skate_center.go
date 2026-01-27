package scraper

import (
	"context"
	"errors"

	"github.com/Eagle-Konbu/shin-yokohama-event-notifier/internal/domain/event"
	"github.com/Eagle-Konbu/shin-yokohama-event-notifier/internal/domain/ports"
)

type SkateCenterScraper struct{}

func NewSkateCenterScraper() ports.EventFetcher {
	return &SkateCenterScraper{}
}

func (s *SkateCenterScraper) FetchEvents(ctx context.Context) ([]event.Event, error) {
	return nil, errors.New("not implemented")
}

func (s *SkateCenterScraper) VenueID() event.VenueID {
	return event.VenueIDSkateCenter
}
