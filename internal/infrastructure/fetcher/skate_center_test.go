package fetcher

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Eagle-Konbu/shin-yokohama-event-notifier/internal/domain/event"
)

func TestNewSkateCenterFetcher(t *testing.T) {
	scraper := NewSkateCenterFetcher()

	require.NotNil(t, scraper)
}

func TestSkateCenterFetcher_FetchEvents_NotImplemented(t *testing.T) {
	scraper := NewSkateCenterFetcher()
	ctx := context.Background()

	events, err := scraper.FetchEvents(ctx)

	require.Error(t, err)
	assert.Equal(t, "not implemented", err.Error())
	assert.Nil(t, events)
}

func TestSkateCenterFetcher_VenueID(t *testing.T) {
	scraper := NewSkateCenterFetcher()

	vid := scraper.VenueID()

	assert.Equal(t, event.VenueIDSkateCenter, vid)
}
