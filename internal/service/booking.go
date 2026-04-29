package service

import (
	"context"
	"net/mail"
	"slices"
	"strings"
	"time"

	"github.com/iselldonuts/ai-for-developers-project-386/internal/dto"
	"github.com/iselldonuts/ai-for-developers-project-386/internal/errs"
	"github.com/iselldonuts/ai-for-developers-project-386/internal/model"
)

type BookingStore interface {
	ListBookings(ctx context.Context) ([]model.Booking, error)
	ListBookingsInRange(ctx context.Context, from time.Time, to time.Time) ([]model.Booking, error)
	CreateBooking(ctx context.Context, input dto.CreateBookingInput, endsAt time.Time) (model.Booking, error)
}

type BookingEventTypeStore interface {
	ListEventTypes(ctx context.Context) ([]model.EventType, error)
	GetEventTypeByID(ctx context.Context, eventTypeID string) (model.EventType, error)
}

type BookingOwnerStore interface {
	GetOwnerProfile(ctx context.Context) (model.OwnerProfile, error)
}

type BookingAvailabilityStore interface {
	ListAvailabilityWindows(ctx context.Context, ownerID string) ([]model.AvailabilityWindow, error)
}

type BookingService struct {
	bookingStore      BookingStore
	eventTypeStore    BookingEventTypeStore
	ownerStore        BookingOwnerStore
	availabilityStore BookingAvailabilityStore
	txManager         TxManager
}

func NewBookingService(
	bookingStore BookingStore,
	eventTypeStore BookingEventTypeStore,
	ownerStore BookingOwnerStore,
	availabilityStore BookingAvailabilityStore,
	txManager TxManager,
) BookingService {
	return BookingService{
		bookingStore:      bookingStore,
		eventTypeStore:    eventTypeStore,
		ownerStore:        ownerStore,
		availabilityStore: availabilityStore,
		txManager:         txManager,
	}
}

func (s BookingService) List(ctx context.Context) ([]model.Booking, error) {
	return s.bookingStore.ListBookings(ctx)
}

func (s BookingService) ListUpcoming(ctx context.Context, from time.Time) (dto.UpcomingBookingsResult, error) {
	bookings, err := s.bookingStore.ListBookings(ctx)
	if err != nil {
		return dto.UpcomingBookingsResult{}, err
	}

	eventTypes, err := s.eventTypeStore.ListEventTypes(ctx)
	if err != nil {
		return dto.UpcomingBookingsResult{}, err
	}

	titlesByID := make(map[string]string, len(eventTypes))
	for _, eventType := range eventTypes {
		titlesByID[eventType.ID] = eventType.Title
	}

	from = from.UTC()
	items := make([]dto.BookingListItem, 0, len(bookings))
	for _, booking := range bookings {
		if booking.EndsAt.UTC().Before(from) {
			continue
		}

		items = append(items, dto.BookingListItem{
			BookingID:      booking.ID,
			EventTypeID:    booking.EventTypeID,
			EventTypeTitle: titlesByID[booking.EventTypeID],
			GuestName:      booking.GuestName,
			GuestEmail:     booking.GuestEmail,
			StartsAt:       booking.StartsAt.UTC(),
			EndsAt:         booking.EndsAt.UTC(),
		})
	}

	slices.SortFunc(items, func(left dto.BookingListItem, right dto.BookingListItem) int {
		if comparison := left.StartsAt.Compare(right.StartsAt); comparison != 0 {
			return comparison
		}

		return strings.Compare(left.BookingID, right.BookingID)
	})

	return dto.UpcomingBookingsResult{Items: items}, nil
}

func (s BookingService) ListSlots(ctx context.Context, eventTypeID string, from time.Time, to time.Time) ([]model.Slot, error) {
	if strings.TrimSpace(eventTypeID) == "" {
		return nil, errs.ValidationError{Message: "eventTypeId is required"}
	}

	from = from.UTC()
	to = to.UTC()

	if !from.Before(to) {
		return nil, errs.ValidationError{Message: "from must be earlier than to"}
	}

	eventType, owner, windows, bookings, err := s.loadSchedulingContext(ctx, eventTypeID, from, to)
	if err != nil {
		return nil, err
	}

	location, err := time.LoadLocation(owner.Timezone)
	if err != nil {
		return nil, err
	}

	windowsByDay := make(map[model.DayOfWeek]model.AvailabilityWindow, len(windows))
	for _, window := range windows {
		windowsByDay[window.DayOfWeek] = window
	}

	slots := make([]model.Slot, 0)
	cursor := startOfDayInLocation(from.In(location))
	toLocal := to.In(location)

	for day := cursor; day.Before(toLocal); day = day.AddDate(0, 0, 1) {
		window, ok := windowsByDay[dayOfWeekFromTime(day)]
		if !ok {
			continue
		}

		windowStart, err := combineLocalDateAndTime(day, window.StartTime, location)
		if err != nil {
			return nil, err
		}

		windowEnd, err := combineLocalDateAndTime(day, window.EndTime, location)
		if err != nil {
			return nil, err
		}

		for slotStart := windowStart; slotStart.Add(time.Duration(eventType.DurationMinutes)*time.Minute).Equal(windowEnd) || slotStart.Add(time.Duration(eventType.DurationMinutes)*time.Minute).Before(windowEnd); slotStart = slotStart.Add(time.Duration(eventType.DurationMinutes) * time.Minute) {
			slotEnd := slotStart.Add(time.Duration(eventType.DurationMinutes) * time.Minute)
			slotStartUTC := slotStart.UTC()
			slotEndUTC := slotEnd.UTC()

			if slotStartUTC.Before(from) || slotEndUTC.After(to) {
				continue
			}

			if hasBookingConflict(bookings, slotStartUTC, slotEndUTC) {
				continue
			}

			slots = append(slots, model.Slot{
				EventTypeID:     eventType.ID,
				StartsAt:        slotStartUTC,
				EndsAt:          slotEndUTC,
				DurationMinutes: eventType.DurationMinutes,
			})
		}
	}

	return slots, nil
}

func (s BookingService) Create(ctx context.Context, input dto.CreateBookingInput) (model.Booking, error) {
	input.EventTypeID = strings.TrimSpace(input.EventTypeID)
	input.GuestName = strings.TrimSpace(input.GuestName)
	input.GuestEmail = strings.TrimSpace(input.GuestEmail)
	input.StartsAt = input.StartsAt.UTC()

	if input.EventTypeID == "" {
		return model.Booking{}, errs.ValidationError{Message: "eventTypeId is required"}
	}
	if input.GuestName == "" {
		return model.Booking{}, errs.ValidationError{Message: "guestName is required"}
	}
	if input.GuestEmail == "" {
		return model.Booking{}, errs.ValidationError{Message: "guestEmail is required"}
	}
	if _, err := mail.ParseAddress(input.GuestEmail); err != nil {
		return model.Booking{}, errs.ValidationError{Message: "guestEmail must be a valid email"}
	}
	if input.StartsAt.IsZero() {
		return model.Booking{}, errs.ValidationError{Message: "startsAt is required"}
	}

	var created model.Booking

	err := s.txManager.WithTx(ctx, func(txCtx context.Context) error {
		eventType, owner, _, bookings, err := s.loadSchedulingContext(txCtx, input.EventTypeID, input.StartsAt, input.StartsAt.AddDate(0, 0, 1))
		if err != nil {
			return err
		}

		endsAt := input.StartsAt.Add(time.Duration(eventType.DurationMinutes) * time.Minute)
		location, err := time.LoadLocation(owner.Timezone)
		if err != nil {
			return err
		}

		dayStartLocal := startOfDayInLocation(input.StartsAt.In(location))
		dayEndLocal := dayStartLocal.AddDate(0, 0, 1)
		slots, err := s.ListSlots(txCtx, input.EventTypeID, dayStartLocal.UTC(), dayEndLocal.UTC())
		if err != nil {
			return err
		}

		slotAvailable := false
		for _, slot := range slots {
			if slot.StartsAt.Equal(input.StartsAt) {
				slotAvailable = true
				break
			}
		}

		if !slotAvailable {
			return errs.ErrConflict
		}

		if hasBookingConflict(bookings, input.StartsAt, endsAt) {
			return errs.ErrConflict
		}

		created, err = s.bookingStore.CreateBooking(txCtx, input, endsAt)
		return err
	})
	if err != nil {
		return model.Booking{}, err
	}

	return created, nil
}

func (s BookingService) loadSchedulingContext(
	ctx context.Context,
	eventTypeID string,
	from time.Time,
	to time.Time,
) (model.EventType, model.OwnerProfile, []model.AvailabilityWindow, []model.Booking, error) {
	eventType, err := s.eventTypeStore.GetEventTypeByID(ctx, eventTypeID)
	if err != nil {
		return model.EventType{}, model.OwnerProfile{}, nil, nil, err
	}

	owner, err := s.ownerStore.GetOwnerProfile(ctx)
	if err != nil {
		return model.EventType{}, model.OwnerProfile{}, nil, nil, err
	}

	windows, err := s.availabilityStore.ListAvailabilityWindows(ctx, owner.ID)
	if err != nil {
		return model.EventType{}, model.OwnerProfile{}, nil, nil, err
	}

	bookings, err := s.bookingStore.ListBookingsInRange(ctx, from, to)
	if err != nil {
		return model.EventType{}, model.OwnerProfile{}, nil, nil, err
	}

	slices.SortFunc(bookings, func(left model.Booking, right model.Booking) int {
		return left.StartsAt.Compare(right.StartsAt)
	})

	return eventType, owner, windows, bookings, nil
}

func hasBookingConflict(bookings []model.Booking, startsAt time.Time, endsAt time.Time) bool {
	for _, booking := range bookings {
		if startsAt.Before(booking.EndsAt) && endsAt.After(booking.StartsAt) {
			return true
		}
	}

	return false
}

func startOfDayInLocation(value time.Time) time.Time {
	year, month, day := value.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, value.Location())
}

func dayOfWeekFromTime(value time.Time) model.DayOfWeek {
	switch value.Weekday() {
	case time.Monday:
		return model.DayOfWeekMonday
	case time.Tuesday:
		return model.DayOfWeekTuesday
	case time.Wednesday:
		return model.DayOfWeekWednesday
	case time.Thursday:
		return model.DayOfWeekThursday
	case time.Friday:
		return model.DayOfWeekFriday
	case time.Saturday:
		return model.DayOfWeekSaturday
	default:
		return model.DayOfWeekSunday
	}
}

func combineLocalDateAndTime(day time.Time, hhmm string, location *time.Location) (time.Time, error) {
	clock, err := time.Parse("15:04", hhmm)
	if err != nil {
		return time.Time{}, errs.ValidationError{Message: "invalid availability time"}
	}

	year, month, date := day.Date()
	return time.Date(year, month, date, clock.Hour(), clock.Minute(), 0, 0, location), nil
}
