package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/keven-rdr/wordloop/api/internal/identity/adapters/postgres/db"
	"github.com/keven-rdr/wordloop/api/internal/identity/app"
)

// Sessions implementa app.Sessions sobre o Postgres.
type Sessions struct{ q *db.Queries }

var _ app.Sessions = (*Sessions)(nil)

// NewSessions recebe o pool (ou uma transacao).
func NewSessions(conn db.DBTX) *Sessions { return &Sessions{q: db.New(conn)} }

// Create grava a sessao (so o hash do id).
func (r *Sessions) Create(ctx context.Context, s app.NewSession) error {
	err := r.q.CreateSession(ctx, db.CreateSessionParams{
		IDHash: s.IDHash, UserID: s.UserID, RefreshTokenEnc: s.RefreshTokenEnc,
		ExpiresAt: pgtype.Timestamptz{Time: s.ExpiresAt, Valid: true}, UserAgent: text(s.UserAgent),
	})
	if err != nil {
		return fmt.Errorf("criar a sessao: %w", err)
	}
	return nil
}

// Get le a sessao pelo hash do id; app.ErrNotFound se nao existe.
func (r *Sessions) Get(ctx context.Context, idHash []byte) (app.SessionRecord, error) {
	row, err := r.q.GetSession(ctx, idHash)
	if errors.Is(err, pgx.ErrNoRows) {
		return app.SessionRecord{}, app.ErrNotFound
	}
	if err != nil {
		return app.SessionRecord{}, fmt.Errorf("ler a sessao: %w", err)
	}
	return app.SessionRecord{UserID: row.UserID, ExpiresAt: row.ExpiresAt.Time.In(time.UTC)}, nil
}

// Touch marca o ultimo uso da sessao.
func (r *Sessions) Touch(ctx context.Context, idHash []byte) error {
	if err := r.q.TouchSession(ctx, idHash); err != nil {
		return fmt.Errorf("atualizar a sessao: %w", err)
	}
	return nil
}

// Delete apaga a sessao.
func (r *Sessions) Delete(ctx context.Context, idHash []byte) error {
	if err := r.q.DeleteSession(ctx, idHash); err != nil {
		return fmt.Errorf("apagar a sessao: %w", err)
	}
	return nil
}
