package event

import "time"

type Event struct {
	Title string
	Date  time.Time
	Venue Venue
	URL   string
}

type Venue string

const (
	VenueNissanStadium Venue = "日産スタジアム"
	VenueYokohamaArena Venue = "横浜アリーナ"
	VenueSkateCenter   Venue = "KOSÉ新横浜スケートセンター"
)
