package scraper

import (
	"context"
	"errors"

	"github.com/Eagle-Konbu/shin-yokohama-event-notifier/internal/domain/event"
	"github.com/Eagle-Konbu/shin-yokohama-event-notifier/internal/domain/ports"
)

type YokohamaArenaScraper struct{}

func NewYokohamaArenaScraper() ports.EventFetcher {
	return &YokohamaArenaScraper{}
}

func (s *YokohamaArenaScraper) FetchEvents(ctx context.Context) ([]event.Event, error) {
	return nil, errors.New("not implemented")
}
