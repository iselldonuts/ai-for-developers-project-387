package postgres

import (
	"context"
	"errors"

	"github.com/iselldonuts/ai-for-developers-project-386/internal/errs"
	"github.com/iselldonuts/ai-for-developers-project-386/internal/model"
	"github.com/iselldonuts/ai-for-developers-project-386/internal/platform/postgres/txmanager"
	"github.com/jackc/pgx/v5"
)

type OwnerStore struct {
	txManager txmanager.Manager
}

func NewOwnerStore(txManager txmanager.Manager) OwnerStore {
	return OwnerStore{txManager: txManager}
}

func (s OwnerStore) GetOwnerProfile(ctx context.Context) (model.OwnerProfile, error) {
	querier := s.txManager.GetQuerier(ctx)

	var owner model.OwnerProfile

	err := querier.QueryRow(ctx, `
		select id, display_name, timezone
		from owner_profiles
		order by id
		limit 1
	`).Scan(&owner.ID, &owner.DisplayName, &owner.Timezone)
	if errors.Is(err, pgx.ErrNoRows) {
		return model.OwnerProfile{}, errs.ErrNotFound
	}
	if err != nil {
		return model.OwnerProfile{}, err
	}

	return owner, nil
}
