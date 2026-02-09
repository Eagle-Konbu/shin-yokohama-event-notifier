package event

import "time"

type Schedule struct {
	StartTime *time.Time
	OpenTime  *time.Time
}

type Event struct {
	Date      time.Time
	Title     string
	Schedules []Schedule
}
