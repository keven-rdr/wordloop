package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
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

func ts(t time.Time) pgtype.Timestamptz { return pgtype.Timestamptz{Time: t, Valid: true} }

// Create grava a sessao (so o hash do id).
func (r *Sessions) Create(ctx context.Context, s app.NewSession) error {
	err := r.q.CreateSession(ctx, db.CreateSessionParams{
		IDHash: s.IDHash, UserID: s.UserID, RefreshTokenEnc: s.RefreshTokenEnc, CreatedAt: ts(s.CreatedAt),
		ExpiresAt: ts(s.ExpiresAt), AuthenticatedAt: ts(s.AuthenticatedAt), Mfa: s.MFA, UserAgent: text(s.UserAgent),
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
	return record(row.PublicID, row.UserID, row.CreatedAt, row.LastUsedAt, row.ExpiresAt, row.AuthenticatedAt, row.Mfa, row.UserAgent), nil
}

func (r *Sessions) ListByUser(ctx context.Context, userID uuid.UUID) ([]app.SessionRecord, error) {
	rows, err := r.q.ListSessionsByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("listar as sessoes: %w", err)
	}
	out := make([]app.SessionRecord, 0, len(rows))
	for _, row := range rows {
		out = append(out, record(row.PublicID, row.UserID, row.CreatedAt, row.LastUsedAt, row.ExpiresAt, row.AuthenticatedAt, row.Mfa, row.UserAgent))
	}
	return out, nil
}

// Touch marca o ultimo uso da sessao.
func (r *Sessions) Touch(ctx context.Context, idHash []byte, at time.Time) error {
	if err := r.q.TouchSession(ctx, db.TouchSessionParams{IDHash: idHash, LastUsedAt: ts(at)}); err != nil {
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

func (r *Sessions) DeleteByPublicID(ctx context.Context, userID, publicID uuid.UUID) error {
	n, err := r.q.DeleteSessionByPublicID(ctx, db.DeleteSessionByPublicIDParams{UserID: userID, PublicID: publicID})
	if err != nil {
		return fmt.Errorf("apagar a sessao: %w", err)
	}
	if n == 0 {
		return app.ErrNotFound
	}
	return nil
}

func (r *Sessions) DeleteAllByUser(ctx context.Context, userID uuid.UUID) error {
	if err := r.q.DeleteSessionsByUser(ctx, userID); err != nil {
		return fmt.Errorf("apagar as sessoes: %w", err)
	}
	return nil
}

func record(publicID, userID uuid.UUID, created, lastUsed, expires, authenticated pgtype.Timestamptz, mfa bool, userAgent pgtype.Text) app.SessionRecord {
	return app.SessionRecord{
		PublicID: publicID, UserID: userID, CreatedAt: created.Time.UTC(), LastUsedAt: lastUsed.Time.UTC(),
		ExpiresAt: expires.Time.UTC(), AuthenticatedAt: authenticated.Time.UTC(), MFA: mfa, UserAgent: userAgent.String,
	}
}
