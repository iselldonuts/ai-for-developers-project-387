package migrations

import (
	"context"
	"database/sql"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

func Up(ctx context.Context, databaseURL string, migrationsDir string) error {
	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return err
	}
	defer db.Close()

	goose.SetDialect("postgres")

	return goose.UpContext(ctx, db, migrationsDir)
}
