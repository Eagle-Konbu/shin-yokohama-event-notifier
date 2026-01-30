package scraper

import (
	"context"
	"errors"
	"time"

	"github.com/Eagle-Konbu/shin-yokohama-event-notifier/internal/domain/event"
	"github.com/Eagle-Konbu/shin-yokohama-event-notifier/internal/domain/ports"
)

type YokohamaArenaScraper struct{}

func NewYokohamaArenaScraper() ports.EventFetcher {
	return &YokohamaArenaScraper{}
}

func (s *YokohamaArenaScraper) FetchEvents(ctx context.Context) ([]event.Event, error) {
	time.Sleep(time.Second * 2)
	return nil, errors.New("not implemented")
}

func (s *YokohamaArenaScraper) VenueID() event.VenueID {
	return event.VenueIDYokohamaArena
}
