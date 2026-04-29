package postgres

import (
	"context"
	"errors"

	"github.com/iselldonuts/ai-for-developers-project-386/internal/dto"
	"github.com/iselldonuts/ai-for-developers-project-386/internal/errs"
	"github.com/iselldonuts/ai-for-developers-project-386/internal/model"
	"github.com/iselldonuts/ai-for-developers-project-386/internal/platform/id"
	"github.com/iselldonuts/ai-for-developers-project-386/internal/platform/postgres/txmanager"
	"github.com/jackc/pgx/v5"
)

type EventTypeStore struct {
	txManager txmanager.Manager
}

func NewEventTypeStore(txManager txmanager.Manager) EventTypeStore {
	return EventTypeStore{txManager: txManager}
}

func (s EventTypeStore) ListEventTypes(ctx context.Context) ([]model.EventType, error) {
	querier := s.txManager.GetQuerier(ctx)

	rows, err := querier.Query(ctx, `
		select id, title, description, duration_minutes
		from event_types
		order by created_at, title
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]model.EventType, 0)

	for rows.Next() {
		var item model.EventType

		if err := rows.Scan(&item.ID, &item.Title, &item.Description, &item.DurationMinutes); err != nil {
			return nil, err
		}

		items = append(items, item)
	}

	return items, rows.Err()
}

func (s EventTypeStore) GetEventTypeByID(ctx context.Context, eventTypeID string) (model.EventType, error) {
	querier := s.txManager.GetQuerier(ctx)

	var item model.EventType

	err := querier.QueryRow(ctx, `
		select id, title, description, duration_minutes
		from event_types
		where id = $1
	`, eventTypeID).Scan(&item.ID, &item.Title, &item.Description, &item.DurationMinutes)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.EventType{}, errs.ErrNotFound
	}
	if err != nil {
		return model.EventType{}, err
	}

	return item, nil
}

func (s EventTypeStore) CreateEventType(ctx context.Context, input dto.CreateEventTypeInput) (model.EventType, error) {
	querier := s.txManager.GetQuerier(ctx)

	item := model.EventType{
		ID:              id.New("evt"),
		Title:           input.Title,
		Description:     input.Description,
		DurationMinutes: input.DurationMinutes,
	}

	if _, err := querier.Exec(ctx, `
		insert into event_types (id, title, description, duration_minutes)
		values ($1, $2, $3, $4)
	`, item.ID, item.Title, item.Description, item.DurationMinutes); err != nil {
		return model.EventType{}, err
	}

	return item, nil
}
