package event

import "time"

type Event struct {
	Date         time.Time
	Title        string
	HasStartTime bool
}
