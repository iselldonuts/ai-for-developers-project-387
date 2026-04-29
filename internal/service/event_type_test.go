package service

import (
	"context"
	"testing"

	"github.com/iselldonuts/ai-for-developers-project-386/internal/dto"
	"github.com/iselldonuts/ai-for-developers-project-386/internal/model"
)

type eventTypeStoreStub struct {
	createInput dto.CreateEventTypeInput
	createFn    func(ctx context.Context, input dto.CreateEventTypeInput) (model.EventType, error)
}

func (s *eventTypeStoreStub) ListEventTypes(_ context.Context) ([]model.EventType, error) {
	return nil, nil
}

func (s *eventTypeStoreStub) GetEventTypeByID(_ context.Context, eventTypeID string) (model.EventType, error) {
	return model.EventType{ID: eventTypeID, DurationMinutes: 30}, nil
}

func (s *eventTypeStoreStub) CreateEventType(ctx context.Context, input dto.CreateEventTypeInput) (model.EventType, error) {
	s.createInput = input
	return s.createFn(ctx, input)
}

type txManagerStub struct{}

func (txManagerStub) WithTx(ctx context.Context, fn func(ctx context.Context) error) error {
	return fn(ctx)
}

func TestEventTypeServiceCreateTrimsInput(t *testing.T) {
	t.Parallel()

	store := &eventTypeStoreStub{
		createFn: func(_ context.Context, input dto.CreateEventTypeInput) (model.EventType, error) {
			return model.EventType{
				ID:              "evt_1",
				Title:           input.Title,
				Description:     input.Description,
				DurationMinutes: input.DurationMinutes,
			}, nil
		},
	}

	service := NewEventTypeService(store, txManagerStub{})

	item, err := service.Create(context.Background(), dto.CreateEventTypeInput{
		Title:           "  Lesson  ",
		Description:     "  Intro  ",
		DurationMinutes: 30,
	})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	if store.createInput.Title != "Lesson" {
		t.Fatalf("expected trimmed title, got %q", store.createInput.Title)
	}

	if store.createInput.Description != "Intro" {
		t.Fatalf("expected trimmed description, got %q", store.createInput.Description)
	}

	if item.DurationMinutes != 30 {
		t.Fatalf("expected duration 30, got %d", item.DurationMinutes)
	}
}
