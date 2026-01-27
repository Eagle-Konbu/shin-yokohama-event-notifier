package event

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVenueID_Constants(t *testing.T) {
	assert.Equal(t, VenueID("yokohama_arena"), VenueIDYokohamaArena)
	assert.Equal(t, VenueID("nissan_stadium"), VenueIDNissanStadium)
	assert.Equal(t, VenueID("skate_center"), VenueIDSkateCenter)
}

func TestNewAllVenues(t *testing.T) {
	venues := NewAllVenues()

	require.Len(t, venues, 3)

	t.Run("YokohamaArena", func(t *testing.T) {
		venue := venues[0]
		assert.Equal(t, VenueIDYokohamaArena, venue.ID)
		assert.Equal(t, "横浜アリーナ", venue.DisplayName)
		assert.Equal(t, "🏟️", venue.Emoji)
		assert.Empty(t, venue.Events)
	})

	t.Run("NissanStadium", func(t *testing.T) {
		venue := venues[1]
		assert.Equal(t, VenueIDNissanStadium, venue.ID)
		assert.Equal(t, "日産スタジアム", venue.DisplayName)
		assert.Equal(t, "⚽", venue.Emoji)
		assert.Empty(t, venue.Events)
	})

	t.Run("SkateCenter", func(t *testing.T) {
		venue := venues[2]
		assert.Equal(t, VenueIDSkateCenter, venue.ID)
		assert.Equal(t, "KOSÉ新横浜スケートセンター", venue.DisplayName)
		assert.Equal(t, "⛸️", venue.Emoji)
		assert.Empty(t, venue.Events)
	})
}
