package scraper

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
