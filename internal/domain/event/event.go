package event

import "time"

type TimeSlot struct {
	StartTime *time.Time
	OpenTime  *time.Time
}

type Event struct {
	Date      time.Time
	Title     string
	TimeSlots []TimeSlot
}
