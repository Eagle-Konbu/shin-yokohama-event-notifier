package event

import "time"

type Event struct {
	Date      time.Time
	StartTime *time.Time
	OpenTime  *time.Time
	Title     string
}
