package txmanager

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Querier interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

type Manager struct {
	pool *pgxpool.Pool
}

type txContextKey struct{}

func New(pool *pgxpool.Pool) Manager {
	return Manager{pool: pool}
}

func (m Manager) WithTx(ctx context.Context, fn func(ctx context.Context) error) error {
	if tx := txFromContext(ctx); tx != nil {
		return fn(ctx)
	}

	tx, err := m.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}

	txCtx := context.WithValue(ctx, txContextKey{}, tx)

	if err := fn(txCtx); err != nil {
		_ = tx.Rollback(ctx)
		return err
	}

	return tx.Commit(ctx)
}

func (m Manager) GetQuerier(ctx context.Context) Querier {
	if tx := txFromContext(ctx); tx != nil {
		return tx
	}

	return m.pool
}

func txFromContext(ctx context.Context) pgx.Tx {
	tx, _ := ctx.Value(txContextKey{}).(pgx.Tx)
	return tx
}
