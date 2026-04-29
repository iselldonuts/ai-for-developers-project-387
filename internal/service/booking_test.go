package service

import (
	"context"
	"testing"
	"time"

	"github.com/iselldonuts/ai-for-developers-project-386/internal/dto"
	"github.com/iselldonuts/ai-for-developers-project-386/internal/errs"
	"github.com/iselldonuts/ai-for-developers-project-386/internal/model"
)

type bookingStoreServiceStub struct {
	bookings []model.Booking
}

func (s *bookingStoreServiceStub) ListBookings(_ context.Context) ([]model.Booking, error) {
	return s.bookings, nil
}

func (s *bookingStoreServiceStub) ListBookingsInRange(_ context.Context, from time.Time, to time.Time) ([]model.Booking, error) {
	items := make([]model.Booking, 0)

	for _, item := range s.bookings {
		if item.StartsAt.Before(to) && item.EndsAt.After(from) {
			items = append(items, item)
		}
	}

	return items, nil
}

func (s *bookingStoreServiceStub) CreateBooking(_ context.Context, input dto.CreateBookingInput, endsAt time.Time) (model.Booking, error) {
	item := model.Booking{
		ID:          "bkg_new",
		EventTypeID: input.EventTypeID,
		GuestName:   input.GuestName,
		GuestEmail:  input.GuestEmail,
		StartsAt:    input.StartsAt,
		EndsAt:      endsAt,
	}
	s.bookings = append(s.bookings, item)
	return item, nil
}

type bookingEventTypeStoreStub struct {
	item model.EventType
	err  error
}

func (s bookingEventTypeStoreStub) GetEventTypeByID(_ context.Context, _ string) (model.EventType, error) {
	return s.item, s.err
}

type bookingOwnerStoreStub struct {
	owner model.OwnerProfile
}

func (s bookingOwnerStoreStub) GetOwnerProfile(_ context.Context) (model.OwnerProfile, error) {
	return s.owner, nil
}

type bookingAvailabilityStoreStub struct {
	windows []model.AvailabilityWindow
}

func (s *bookingAvailabilityStoreStub) ListAvailabilityWindows(_ context.Context, _ string) ([]model.AvailabilityWindow, error) {
	return s.windows, nil
}

func TestBookingServiceListSlotsExcludesBookedIntervals(t *testing.T) {
	t.Parallel()

	service := NewBookingService(
		&bookingStoreServiceStub{
			bookings: []model.Booking{
				{
					ID:          "bkg_1",
					EventTypeID: "evt_1",
					GuestName:   "Guest",
					GuestEmail:  "guest@example.com",
					StartsAt:    time.Date(2026, time.April, 13, 9, 30, 0, 0, time.UTC),
					EndsAt:      time.Date(2026, time.April, 13, 10, 0, 0, 0, time.UTC),
				},
			},
		},
		bookingEventTypeStoreStub{item: model.EventType{ID: "evt_1", DurationMinutes: 30}},
		bookingOwnerStoreStub{owner: model.OwnerProfile{ID: "owner_1", DisplayName: "Owner", Timezone: "UTC"}},
		&bookingAvailabilityStoreStub{
			windows: []model.AvailabilityWindow{
				{DayOfWeek: model.DayOfWeekMonday, StartTime: "09:00", EndTime: "11:00"},
			},
		},
		txManagerStub{},
	)

	slots, err := service.ListSlots(
		context.Background(),
		"evt_1",
		time.Date(2026, time.April, 13, 0, 0, 0, 0, time.UTC),
		time.Date(2026, time.April, 14, 0, 0, 0, 0, time.UTC),
	)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(slots) != 3 {
		t.Fatalf("expected 3 free slots, got %d", len(slots))
	}

	if slots[0].StartsAt != time.Date(2026, time.April, 13, 9, 0, 0, 0, time.UTC) {
		t.Fatalf("unexpected first slot start: %s", slots[0].StartsAt)
	}
	if slots[1].StartsAt != time.Date(2026, time.April, 13, 10, 0, 0, 0, time.UTC) {
		t.Fatalf("unexpected second slot start: %s", slots[1].StartsAt)
	}
}

func TestBookingServiceCreateReturnsConflictForOverlappingBooking(t *testing.T) {
	t.Parallel()

	service := NewBookingService(
		&bookingStoreServiceStub{
			bookings: []model.Booking{
				{
					ID:          "bkg_1",
					EventTypeID: "evt_1",
					GuestName:   "Guest",
					GuestEmail:  "guest@example.com",
					StartsAt:    time.Date(2026, time.April, 13, 9, 0, 0, 0, time.UTC),
					EndsAt:      time.Date(2026, time.April, 13, 9, 30, 0, 0, time.UTC),
				},
			},
		},
		bookingEventTypeStoreStub{item: model.EventType{ID: "evt_1", DurationMinutes: 30}},
		bookingOwnerStoreStub{owner: model.OwnerProfile{ID: "owner_1", DisplayName: "Owner", Timezone: "UTC"}},
		&bookingAvailabilityStoreStub{
			windows: []model.AvailabilityWindow{
				{DayOfWeek: model.DayOfWeekMonday, StartTime: "09:00", EndTime: "17:00"},
			},
		},
		txManagerStub{},
	)

	_, err := service.Create(context.Background(), dto.CreateBookingInput{
		EventTypeID: "evt_1",
		GuestName:   "Jamie",
		GuestEmail:  "jamie@example.com",
		StartsAt:    time.Date(2026, time.April, 13, 9, 0, 0, 0, time.UTC),
	})
	if err == nil {
		t.Fatal("expected conflict error, got nil")
	}
	if err != errs.ErrConflict {
		t.Fatalf("expected ErrConflict, got %v", err)
	}
}
