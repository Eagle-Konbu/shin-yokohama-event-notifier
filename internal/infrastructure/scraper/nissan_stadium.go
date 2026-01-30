package scraper

import (
	"context"
	"errors"
	"time"

	"github.com/Eagle-Konbu/shin-yokohama-event-notifier/internal/domain/event"
	"github.com/Eagle-Konbu/shin-yokohama-event-notifier/internal/domain/ports"
)

type NissanStadiumScraper struct{}

func NewNissanStadiumScraper() ports.EventFetcher {
	return &NissanStadiumScraper{}
}

func (s *NissanStadiumScraper) FetchEvents(ctx context.Context) ([]event.Event, error) {
	time.Sleep(time.Second * 2) // Simulate network delay
	return nil, errors.New("not implemented")
}

func (s *NissanStadiumScraper) VenueID() event.VenueID {
	return event.VenueIDNissanStadium
}
