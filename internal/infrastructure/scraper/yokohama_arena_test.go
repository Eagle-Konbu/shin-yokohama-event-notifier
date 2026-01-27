package scraper

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Eagle-Konbu/shin-yokohama-event-notifier/internal/domain/event"
)

func TestNewYokohamaArenaScraper(t *testing.T) {
	scraper := NewYokohamaArenaScraper()

	require.NotNil(t, scraper)
}

func TestYokohamaArenaScraper_FetchEvents_NotImplemented(t *testing.T) {
	scraper := NewYokohamaArenaScraper()
	ctx := context.Background()

	events, err := scraper.FetchEvents(ctx)

	require.Error(t, err)
	assert.Equal(t, "not implemented", err.Error())
	assert.Nil(t, events)
}

func TestYokohamaArenaScraper_VenueID(t *testing.T) {
	scraper := NewYokohamaArenaScraper()

	vid := scraper.VenueID()

	assert.Equal(t, event.VenueIDYokohamaArena, vid)
}
