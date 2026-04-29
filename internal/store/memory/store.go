package memory

import (
	"context"
	"sync"
	"time"

	"github.com/iselldonuts/ai-for-developers-project-386/internal/dto"
	"github.com/iselldonuts/ai-for-developers-project-386/internal/errs"
	"github.com/iselldonuts/ai-for-developers-project-386/internal/model"
	"github.com/iselldonuts/ai-for-developers-project-386/internal/platform/id"
)

type Store struct {
	mu         sync.RWMutex
	owner      model.OwnerProfile
	windows    []model.AvailabilityWindow
	eventTypes []model.EventType
	bookings   []model.Booking
}

func New() *Store {
	return &Store{
		owner: model.OwnerProfile{
			ID:          "owner-default",
			DisplayName: "Owner",
			Timezone:    "Europe/Moscow",
		},
		windows: []model.AvailabilityWindow{
			{DayOfWeek: model.DayOfWeekMonday, StartTime: "09:00", EndTime: "17:00"},
			{DayOfWeek: model.DayOfWeekTuesday, StartTime: "09:00", EndTime: "17:00"},
			{DayOfWeek: model.DayOfWeekWednesday, StartTime: "09:00", EndTime: "17:00"},
			{DayOfWeek: model.DayOfWeekThursday, StartTime: "09:00", EndTime: "17:00"},
			{DayOfWeek: model.DayOfWeekFriday, StartTime: "09:00", EndTime: "17:00"},
		},
		eventTypes: []model.EventType{
			{ID: "event-type-intro", Title: "Знакомство", Description: "", DurationMinutes: 15},
			{ID: "event-type-lesson", Title: "Урок", Description: "", DurationMinutes: 30},
		},
	}
}

func (s *Store) GetOwnerProfile(_ context.Context) (model.OwnerProfile, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.owner, nil
}

func (s *Store) ListAvailabilityWindows(_ context.Context, _ string) ([]model.AvailabilityWindow, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return append([]model.AvailabilityWindow(nil), s.windows...), nil
}

func (s *Store) ReplaceAvailabilityWindows(_ context.Context, _ string, windows []model.AvailabilityWindow) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.windows = append([]model.AvailabilityWindow(nil), windows...)
	return nil
}

func (s *Store) ListEventTypes(_ context.Context) ([]model.EventType, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return append([]model.EventType(nil), s.eventTypes...), nil
}

func (s *Store) GetEventTypeByID(_ context.Context, eventTypeID string) (model.EventType, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, item := range s.eventTypes {
		if item.ID == eventTypeID {
			return item, nil
		}
	}

	return model.EventType{}, errs.ErrNotFound
}

func (s *Store) CreateEventType(_ context.Context, input dto.CreateEventTypeInput) (model.EventType, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	item := model.EventType{
		ID:              id.New("event-type"),
		Title:           input.Title,
		Description:     input.Description,
		DurationMinutes: input.DurationMinutes,
	}
	s.eventTypes = append(s.eventTypes, item)

	return item, nil
}

func (s *Store) ListBookings(_ context.Context) ([]model.Booking, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return append([]model.Booking(nil), s.bookings...), nil
}

func (s *Store) ListBookingsInRange(_ context.Context, from time.Time, to time.Time) ([]model.Booking, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	items := make([]model.Booking, 0, len(s.bookings))
	for _, booking := range s.bookings {
		if booking.StartsAt.Before(to) && booking.EndsAt.After(from) {
			items = append(items, booking)
		}
	}

	return items, nil
}

func (s *Store) CreateBooking(_ context.Context, input dto.CreateBookingInput, endsAt time.Time) (model.Booking, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	item := model.Booking{
		ID:          id.New("booking"),
		EventTypeID: input.EventTypeID,
		GuestName:   input.GuestName,
		GuestEmail:  input.GuestEmail,
		StartsAt:    input.StartsAt,
		EndsAt:      endsAt,
	}
	s.bookings = append(s.bookings, item)

	return item, nil
}
