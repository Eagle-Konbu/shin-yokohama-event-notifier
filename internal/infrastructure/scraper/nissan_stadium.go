package scraper

import (
	"context"
	"errors"

	"github.com/Eagle-Konbu/shin-yokohama-event-notifier/internal/domain/event"
	"github.com/Eagle-Konbu/shin-yokohama-event-notifier/internal/domain/ports"
)

type NissanStadiumScraper struct{}

func NewNissanStadiumScraper() ports.EventFetcher {
	return &NissanStadiumScraper{}
}

func (s *NissanStadiumScraper) FetchEvents(ctx context.Context) ([]event.Event, error) {
	return nil, errors.New("not implemented")
}
