package scraper

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Eagle-Konbu/shin-yokohama-event-notifier/internal/domain/event"
)

func TestNewNissanStadiumScraper(t *testing.T) {
	scraper := NewNissanStadiumScraper()

	require.NotNil(t, scraper)
}

func TestNissanStadiumScraper_FetchEvents_NotImplemented(t *testing.T) {
	scraper := NewNissanStadiumScraper()
	ctx := context.Background()

	events, err := scraper.FetchEvents(ctx)

	require.Error(t, err)
	assert.Equal(t, "not implemented", err.Error())
	assert.Nil(t, events)
}

func TestNissanStadiumScraper_VenueID(t *testing.T) {
	scraper := NewNissanStadiumScraper()

	vid := scraper.VenueID()

	assert.Equal(t, event.VenueIDNissanStadium, vid)
}
