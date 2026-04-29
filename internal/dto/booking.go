package dto

import "time"

type CreateBookingInput struct {
	EventTypeID string
	GuestName   string
	GuestEmail  string
	StartsAt    time.Time
}

type UpcomingBookingsResult struct {
	Items []BookingListItem
}

type BookingListItem struct {
	BookingID       string
	EventTypeTitle  string
	GuestName       string
	GuestEmail      string
	StartsAt        time.Time
	DurationMinutes int
}
