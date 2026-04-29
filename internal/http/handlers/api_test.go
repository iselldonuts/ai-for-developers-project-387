package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/iselldonuts/ai-for-developers-project-386/internal/dto"
	"github.com/iselldonuts/ai-for-developers-project-386/internal/model"
	"github.com/iselldonuts/ai-for-developers-project-386/internal/service"
)

type eventTypeStoreStub struct {
	createFn func(ctx context.Context, input dto.CreateEventTypeInput) (model.EventType, error)
}

func (s *eventTypeStoreStub) ListEventTypes(_ context.Context) ([]model.EventType, error) {
	return nil, nil
}

func (s *eventTypeStoreStub) GetEventTypeByID(_ context.Context, eventTypeID string) (model.EventType, error) {
	return model.EventType{
		ID:              eventTypeID,
		Title:           "Lesson",
		Description:     "",
		DurationMinutes: 30,
	}, nil
}

func (s *eventTypeStoreStub) CreateEventType(ctx context.Context, input dto.CreateEventTypeInput) (model.EventType, error) {
	return s.createFn(ctx, input)
}

type availabilityStoreStub struct {
	windows []model.AvailabilityWindow
}

func (s *availabilityStoreStub) ListAvailabilityWindows(_ context.Context, _ string) ([]model.AvailabilityWindow, error) {
	return s.windows, nil
}

func (s *availabilityStoreStub) ReplaceAvailabilityWindows(_ context.Context, _ string, windows []model.AvailabilityWindow) error {
	s.windows = windows
	return nil
}

type bookingStoreStub struct {
	bookings []model.Booking
}

func (s *bookingStoreStub) ListBookings(_ context.Context) ([]model.Booking, error) {
	return s.bookings, nil
}

func (s *bookingStoreStub) ListBookingsInRange(_ context.Context, from time.Time, to time.Time) ([]model.Booking, error) {
	items := make([]model.Booking, 0)

	for _, item := range s.bookings {
		if item.StartsAt.Before(to) && item.EndsAt.After(from) {
			items = append(items, item)
		}
	}

	return items, nil
}

func (s *bookingStoreStub) CreateBooking(_ context.Context, input dto.CreateBookingInput, endsAt time.Time) (model.Booking, error) {
	item := model.Booking{
		ID:          "bkg_1",
		EventTypeID: input.EventTypeID,
		GuestName:   input.GuestName,
		GuestEmail:  input.GuestEmail,
		StartsAt:    input.StartsAt,
		EndsAt:      endsAt,
	}
	s.bookings = append(s.bookings, item)
	return item, nil
}

type ownerStoreStub struct {
	owner model.OwnerProfile
}

func (s ownerStoreStub) GetOwnerProfile(_ context.Context) (model.OwnerProfile, error) {
	return s.owner, nil
}

type txManagerStub struct{}

func (txManagerStub) WithTx(ctx context.Context, fn func(ctx context.Context) error) error {
	return fn(ctx)
}

func TestOwnerEventTypesApiCreate(t *testing.T) {
	t.Parallel()

	eventTypeStore := &eventTypeStoreStub{
		createFn: func(_ context.Context, input dto.CreateEventTypeInput) (model.EventType, error) {
			return model.EventType{
				ID:              "evt_1",
				Title:           input.Title,
				Description:     input.Description,
				DurationMinutes: input.DurationMinutes,
			}, nil
		},
	}

	handler := NewAPIHandler(
		service.NewAvailabilityService(
			ownerStoreStub{owner: model.OwnerProfile{ID: "owner_1", DisplayName: "Owner", Timezone: "Europe/Moscow"}},
			&availabilityStoreStub{},
			txManagerStub{},
		),
		service.NewBookingService(
			&bookingStoreStub{},
			eventTypeStore,
			ownerStoreStub{owner: model.OwnerProfile{ID: "owner_1", DisplayName: "Owner", Timezone: "Europe/Moscow"}},
			&availabilityStoreStub{},
			txManagerStub{},
		),
		service.NewEventTypeService(eventTypeStore, txManagerStub{}),
	)

	request := httptest.NewRequest(http.MethodPost, "/owner/event-types", strings.NewReader(`{"title":"Lesson","description":"Intro","durationMinutes":30}`))
	response := httptest.NewRecorder()

	handler.OwnerEventTypesApiCreate(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", response.Code)
	}
}

func TestOwnerAvailabilityApiUpdateAvailabilityRejectsDuplicateDays(t *testing.T) {
	t.Parallel()

	handler := NewAPIHandler(
		service.NewAvailabilityService(
			ownerStoreStub{owner: model.OwnerProfile{ID: "owner_1", DisplayName: "Owner", Timezone: "Europe/Moscow"}},
			&availabilityStoreStub{},
			txManagerStub{},
		),
		service.NewBookingService(
			&bookingStoreStub{},
			&eventTypeStoreStub{},
			ownerStoreStub{owner: model.OwnerProfile{ID: "owner_1", DisplayName: "Owner", Timezone: "Europe/Moscow"}},
			&availabilityStoreStub{},
			txManagerStub{},
		),
		service.NewEventTypeService(&eventTypeStoreStub{}, txManagerStub{}),
	)

	request := httptest.NewRequest(http.MethodPut, "/owner/availability", strings.NewReader(`{"windows":[{"dayOfWeek":"monday","startTime":"09:00","endTime":"17:00"},{"dayOfWeek":"monday","startTime":"10:00","endTime":"18:00"}]}`))
	response := httptest.NewRecorder()

	handler.OwnerAvailabilityApiUpdateAvailability(response, request)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", response.Code)
	}
}

func TestPublicBookingsApiCreateReturnsConflictWhenSlotIsTaken(t *testing.T) {
	t.Parallel()

	bookingStore := &bookingStoreStub{
		bookings: []model.Booking{
			{
				ID:          "bkg_existing",
				EventTypeID: "evt_1",
				GuestName:   "Alex",
				GuestEmail:  "alex@example.com",
				StartsAt:    time.Date(2026, time.April, 13, 9, 0, 0, 0, time.UTC),
				EndsAt:      time.Date(2026, time.April, 13, 9, 30, 0, 0, time.UTC),
			},
		},
	}

	handler := NewAPIHandler(
		service.NewAvailabilityService(
			ownerStoreStub{owner: model.OwnerProfile{ID: "owner_1", DisplayName: "Owner", Timezone: "UTC"}},
			&availabilityStoreStub{
				windows: []model.AvailabilityWindow{
					{DayOfWeek: model.DayOfWeekMonday, StartTime: "09:00", EndTime: "17:00"},
				},
			},
			txManagerStub{},
		),
		service.NewBookingService(
			bookingStore,
			&eventTypeStoreStub{},
			ownerStoreStub{owner: model.OwnerProfile{ID: "owner_1", DisplayName: "Owner", Timezone: "UTC"}},
			&availabilityStoreStub{
				windows: []model.AvailabilityWindow{
					{DayOfWeek: model.DayOfWeekMonday, StartTime: "09:00", EndTime: "17:00"},
				},
			},
			txManagerStub{},
		),
		service.NewEventTypeService(&eventTypeStoreStub{}, txManagerStub{}),
	)

	request := httptest.NewRequest(http.MethodPost, "/bookings", strings.NewReader(`{"eventTypeId":"evt_1","guestName":"Jamie","guestEmail":"jamie@example.com","startsAt":"2026-04-13T09:00:00Z"}`))
	response := httptest.NewRecorder()

	handler.PublicBookingsApiCreate(response, request)

	if response.Code != http.StatusConflict {
		t.Fatalf("expected status 409, got %d", response.Code)
	}
}
