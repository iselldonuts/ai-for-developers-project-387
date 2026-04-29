package postgres

import (
	"context"
	"time"

	"github.com/iselldonuts/ai-for-developers-project-386/internal/dto"
	"github.com/iselldonuts/ai-for-developers-project-386/internal/model"
	"github.com/iselldonuts/ai-for-developers-project-386/internal/platform/id"
	"github.com/iselldonuts/ai-for-developers-project-386/internal/platform/postgres/txmanager"
)

type BookingStore struct {
	txManager txmanager.Manager
}

func NewBookingStore(txManager txmanager.Manager) BookingStore {
	return BookingStore{txManager: txManager}
}

func (s BookingStore) ListBookings(ctx context.Context) ([]model.Booking, error) {
	querier := s.txManager.GetQuerier(ctx)

	rows, err := querier.Query(ctx, `
		select id, event_type_id, guest_name, guest_email, starts_at, ends_at
		from bookings
		order by starts_at, id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]model.Booking, 0)

	for rows.Next() {
		var item model.Booking

		if err := rows.Scan(
			&item.ID,
			&item.EventTypeID,
			&item.GuestName,
			&item.GuestEmail,
			&item.StartsAt,
			&item.EndsAt,
		); err != nil {
			return nil, err
		}

		items = append(items, item)
	}

	return items, rows.Err()
}

func (s BookingStore) ListBookingsInRange(ctx context.Context, from time.Time, to time.Time) ([]model.Booking, error) {
	querier := s.txManager.GetQuerier(ctx)

	rows, err := querier.Query(ctx, `
		select id, event_type_id, guest_name, guest_email, starts_at, ends_at
		from bookings
		where starts_at < $2 and ends_at > $1
		order by starts_at, id
	`, from.UTC(), to.UTC())
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]model.Booking, 0)

	for rows.Next() {
		var item model.Booking

		if err := rows.Scan(
			&item.ID,
			&item.EventTypeID,
			&item.GuestName,
			&item.GuestEmail,
			&item.StartsAt,
			&item.EndsAt,
		); err != nil {
			return nil, err
		}

		items = append(items, item)
	}

	return items, rows.Err()
}

func (s BookingStore) CreateBooking(ctx context.Context, input dto.CreateBookingInput, endsAt time.Time) (model.Booking, error) {
	querier := s.txManager.GetQuerier(ctx)

	item := model.Booking{
		ID:          id.New("bkg"),
		EventTypeID: input.EventTypeID,
		GuestName:   input.GuestName,
		GuestEmail:  input.GuestEmail,
		StartsAt:    input.StartsAt.UTC(),
		EndsAt:      endsAt.UTC(),
	}

	if _, err := querier.Exec(ctx, `
		insert into bookings (id, event_type_id, guest_name, guest_email, starts_at, ends_at)
		values ($1, $2, $3, $4, $5, $6)
	`, item.ID, item.EventTypeID, item.GuestName, item.GuestEmail, item.StartsAt, item.EndsAt); err != nil {
		return model.Booking{}, err
	}

	return item, nil
}
