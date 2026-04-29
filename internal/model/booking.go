package model

import "time"

type Booking struct {
	ID          string
	EventTypeID string
	GuestName   string
	GuestEmail  string
	StartsAt    time.Time
	EndsAt      time.Time
}

type Slot struct {
	EventTypeID     string
	StartsAt        time.Time
	EndsAt          time.Time
	DurationMinutes int
}
