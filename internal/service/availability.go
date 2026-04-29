package service

import (
	"context"
	"fmt"
	"time"

	"github.com/iselldonuts/ai-for-developers-project-386/internal/dto"
	"github.com/iselldonuts/ai-for-developers-project-386/internal/errs"
	"github.com/iselldonuts/ai-for-developers-project-386/internal/model"
)

type OwnerStore interface {
	GetOwnerProfile(ctx context.Context) (model.OwnerProfile, error)
}

type AvailabilityStore interface {
	ListAvailabilityWindows(ctx context.Context, ownerID string) ([]model.AvailabilityWindow, error)
	ReplaceAvailabilityWindows(ctx context.Context, ownerID string, windows []model.AvailabilityWindow) error
}

type AvailabilitySnapshot struct {
	Owner   model.OwnerProfile
	Windows []model.AvailabilityWindow
}

type AvailabilityService struct {
	ownerStore        OwnerStore
	availabilityStore AvailabilityStore
	txManager         TxManager
}

func NewAvailabilityService(ownerStore OwnerStore, availabilityStore AvailabilityStore, txManager TxManager) AvailabilityService {
	return AvailabilityService{
		ownerStore:        ownerStore,
		availabilityStore: availabilityStore,
		txManager:         txManager,
	}
}

func (s AvailabilityService) Get(ctx context.Context) (AvailabilitySnapshot, error) {
	owner, err := s.ownerStore.GetOwnerProfile(ctx)
	if err != nil {
		return AvailabilitySnapshot{}, err
	}

	windows, err := s.availabilityStore.ListAvailabilityWindows(ctx, owner.ID)
	if err != nil {
		return AvailabilitySnapshot{}, err
	}

	return AvailabilitySnapshot{
		Owner:   owner,
		Windows: windows,
	}, nil
}

func (s AvailabilityService) Replace(ctx context.Context, input dto.ReplaceAvailabilityInput) (AvailabilitySnapshot, error) {
	if err := validateAvailabilityWindows(input.Windows); err != nil {
		return AvailabilitySnapshot{}, err
	}

	var snapshot AvailabilitySnapshot

	err := s.txManager.WithTx(ctx, func(txCtx context.Context) error {
		owner, err := s.ownerStore.GetOwnerProfile(txCtx)
		if err != nil {
			return err
		}

		if err := s.availabilityStore.ReplaceAvailabilityWindows(txCtx, owner.ID, input.Windows); err != nil {
			return err
		}

		windows, err := s.availabilityStore.ListAvailabilityWindows(txCtx, owner.ID)
		if err != nil {
			return err
		}

		snapshot = AvailabilitySnapshot{
			Owner:   owner,
			Windows: windows,
		}

		return nil
	})
	if err != nil {
		return AvailabilitySnapshot{}, err
	}

	return snapshot, nil
}

func validateAvailabilityWindows(windows []model.AvailabilityWindow) error {
	seenDays := make(map[model.DayOfWeek]struct{}, len(windows))

	for _, window := range windows {
		if !window.DayOfWeek.Valid() {
			return errs.ValidationError{Message: fmt.Sprintf("invalid dayOfWeek: %s", window.DayOfWeek)}
		}

		if _, exists := seenDays[window.DayOfWeek]; exists {
			return errs.ValidationError{Message: fmt.Sprintf("duplicate dayOfWeek: %s", window.DayOfWeek)}
		}

		seenDays[window.DayOfWeek] = struct{}{}

		startTime, err := time.Parse("15:04", window.StartTime)
		if err != nil {
			return errs.ValidationError{Message: fmt.Sprintf("invalid startTime for %s", window.DayOfWeek)}
		}

		endTime, err := time.Parse("15:04", window.EndTime)
		if err != nil {
			return errs.ValidationError{Message: fmt.Sprintf("invalid endTime for %s", window.DayOfWeek)}
		}

		if !startTime.Before(endTime) {
			return errs.ValidationError{Message: fmt.Sprintf("startTime must be earlier than endTime for %s", window.DayOfWeek)}
		}
	}

	return nil
}
