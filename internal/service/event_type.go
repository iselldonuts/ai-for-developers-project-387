package service

import (
	"context"
	"strings"

	"github.com/iselldonuts/ai-for-developers-project-386/internal/dto"
	"github.com/iselldonuts/ai-for-developers-project-386/internal/errs"
	"github.com/iselldonuts/ai-for-developers-project-386/internal/model"
)

type EventTypeStore interface {
	ListEventTypes(ctx context.Context) ([]model.EventType, error)
	CreateEventType(ctx context.Context, input dto.CreateEventTypeInput) (model.EventType, error)
}

type EventTypeService struct {
	store     EventTypeStore
	txManager TxManager
}

func NewEventTypeService(store EventTypeStore, txManager TxManager) EventTypeService {
	return EventTypeService{
		store:     store,
		txManager: txManager,
	}
}

func (s EventTypeService) List(ctx context.Context) ([]model.EventType, error) {
	return s.store.ListEventTypes(ctx)
}

func (s EventTypeService) Create(ctx context.Context, input dto.CreateEventTypeInput) (model.EventType, error) {
	input.Title = strings.TrimSpace(input.Title)
	input.Description = strings.TrimSpace(input.Description)

	if input.Title == "" {
		return model.EventType{}, errs.ValidationError{Message: "title is required"}
	}

	if input.DurationMinutes <= 0 {
		return model.EventType{}, errs.ValidationError{Message: "durationMinutes must be greater than 0"}
	}

	var created model.EventType

	err := s.txManager.WithTx(ctx, func(txCtx context.Context) error {
		var err error
		created, err = s.store.CreateEventType(txCtx, input)
		return err
	})
	if err != nil {
		return model.EventType{}, err
	}

	return created, nil
}
