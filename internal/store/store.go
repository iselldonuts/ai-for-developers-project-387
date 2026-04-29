package store

import (
	"context"
	"time"

	"github.com/iselldonuts/ai-for-developers-project-386/internal/dto"
	"github.com/iselldonuts/ai-for-developers-project-386/internal/model"
)

type BookingStore interface {
	ListBookings(ctx context.Context) ([]model.Booking, error)
	ListBookingsInRange(ctx context.Context, from time.Time, to time.Time) ([]model.Booking, error)
	CreateBooking(ctx context.Context, input dto.CreateBookingInput, endsAt time.Time) (model.Booking, error)
}

type OwnerStore interface {
	GetOwnerProfile(ctx context.Context) (model.OwnerProfile, error)
}

type AvailabilityStore interface {
	ListAvailabilityWindows(ctx context.Context, ownerID string) ([]model.AvailabilityWindow, error)
	ReplaceAvailabilityWindows(ctx context.Context, ownerID string, windows []model.AvailabilityWindow) error
}

type EventTypeStore interface {
	ListEventTypes(ctx context.Context) ([]model.EventType, error)
	GetEventTypeByID(ctx context.Context, eventTypeID string) (model.EventType, error)
	CreateEventType(ctx context.Context, input dto.CreateEventTypeInput) (model.EventType, error)
}
