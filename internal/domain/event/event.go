package event

import "time"

type Event struct {
	Title string
	Date  time.Time
	Venue VenueID
}
