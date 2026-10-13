package main

import (
	"context"
	"errors"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/keven-rdr/wordloop/api/internal/platform/db"
	"github.com/keven-rdr/wordloop/api/migrations"
)

// openDatabase conecta ao Postgres de DATABASE_URL. O banco e obrigatorio em todos os ambientes (dev, tst e prd).
func openDatabase(ctx context.Context) (*pgxpool.Pool, error) {
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		return nil, errors.New("DATABASE_URL nao definida")
	}
	return db.Open(ctx, url)
}

// migrateUp aplica as migracoes pendentes e sai. O deploy o roda ANTES de subir a versao nova.
func migrateUp(ctx context.Context) error {
	pool, err := openDatabase(ctx)
	if err != nil {
		return err
	}
	defer pool.Close()
	return db.Migrate(ctx, pool, migrations.FS)
}
