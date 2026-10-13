// Package db abre o pool do PostgreSQL e aplica as migracoes (goose). Sem ORM: o SQL e escrito a mao (ADR 0007).
package db

import (
	"context"
	"fmt"
	"io/fs"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

// maxConns e baixo de proposito: o Neon free limita conexoes e a VM de tst tem 1 GB de RAM.
const maxConns = 5

// Open abre o pool e confirma que o banco responde. A URL vem de DATABASE_URL; nunca a registre em log.
func Open(ctx context.Context, url string) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(url)
	if err != nil {
		return nil, fmt.Errorf("DATABASE_URL invalida: %w", err)
	}
	cfg.MaxConns = maxConns
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("abrir o pool do banco: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("o banco nao responde: %w", err)
	}
	return pool, nil
}

// Migrate aplica as migracoes pendentes do diretorio raiz de migrations. Idempotente.
// Politica expand/contract: uma versao so adiciona; remover coluna ou tabela vem em versao posterior.
func Migrate(ctx context.Context, pool *pgxpool.Pool, migrations fs.FS) error {
	conn := stdlib.OpenDBFromPool(pool)
	defer func() { _ = conn.Close() }()
	provider, err := goose.NewProvider(goose.DialectPostgres, conn, migrations)
	if err != nil {
		return fmt.Errorf("preparar as migracoes: %w", err)
	}
	if _, err := provider.Up(ctx); err != nil {
		return fmt.Errorf("aplicar as migracoes: %w", err)
	}
	return nil
}
