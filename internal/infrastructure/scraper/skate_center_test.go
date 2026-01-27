package scraper

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Eagle-Konbu/shin-yokohama-event-notifier/internal/domain/event"
)

func TestNewSkateCenterScraper(t *testing.T) {
	scraper := NewSkateCenterScraper()

	require.NotNil(t, scraper)
}

func TestSkateCenterScraper_FetchEvents_NotImplemented(t *testing.T) {
	scraper := NewSkateCenterScraper()
	ctx := context.Background()

	events, err := scraper.FetchEvents(ctx)

	require.Error(t, err)
	assert.Equal(t, "not implemented", err.Error())
	assert.Nil(t, events)
}

func TestSkateCenterScraper_VenueID(t *testing.T) {
	scraper := NewSkateCenterScraper()

	vid := scraper.VenueID()

	assert.Equal(t, event.VenueIDSkateCenter, vid)
}
