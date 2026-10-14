package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/keven-rdr/wordloop/api/internal/identity/adapters/postgres/db"
	"github.com/keven-rdr/wordloop/api/internal/identity/passwords"
)

type Passwords struct {
	pool *pgxpool.Pool
	q    *db.Queries
}

var _ passwords.Store = (*Passwords)(nil)

func NewPasswords(pool *pgxpool.Pool) *Passwords { return &Passwords{pool: pool, q: db.New(pool)} }

func (r *Passwords) FindByEmail(ctx context.Context, email string) (passwords.Account, error) {
	row, err := r.q.FindPasswordAccount(ctx, email)
	if errors.Is(err, pgx.ErrNoRows) {
		return passwords.Account{}, passwords.ErrNotFound
	}
	if err != nil {
		return passwords.Account{}, fmt.Errorf("buscar a conta: %w", err)
	}
	return passwords.Account{ID: row.ID, Hash: row.PasswordHash.String, LockedUntil: lockTime(row.LockedUntil.Time, row.LockedUntil.Valid)}, nil
}

func (r *Passwords) BeginAttempt(ctx context.Context, id uuid.UUID, at time.Time) (bool, error) {
	_, err := r.q.BeginPasswordAttempt(ctx, db.BeginPasswordAttemptParams{ID: id, At: ts(at)})
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("contar a tentativa: %w", err)
	}
	return true, nil
}

func lockTime(t time.Time, valid bool) time.Time {
	if !valid {
		return time.Time{}
	}
	return t
}

func (r *Passwords) RegisterSuccess(ctx context.Context, id uuid.UUID) error {
	return r.q.ResetPasswordFailures(ctx, id)
}

func (r *Passwords) PeekInvite(ctx context.Context, inviteHash []byte, now time.Time) (passwords.Owner, error) {
	row, err := r.q.PeekInviteOwner(ctx, db.PeekInviteOwnerParams{IDHash: inviteHash, ExpiresAt: ts(now)})
	if errors.Is(err, pgx.ErrNoRows) {
		return passwords.Owner{}, passwords.ErrNotFound
	}
	if err != nil {
		return passwords.Owner{}, fmt.Errorf("ler o convite: %w", err)
	}
	return passwords.Owner{ID: row.UserID, Email: row.Email.String}, nil
}

func (r *Passwords) SetWithInvite(ctx context.Context, inviteHash []byte, userID uuid.UUID, hash string, now time.Time) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	q := r.q.WithTx(tx)
	owner, err := q.ConsumeInvite(ctx, db.ConsumeInviteParams{IDHash: inviteHash, UsedAt: ts(now)})
	if errors.Is(err, pgx.ErrNoRows) || (err == nil && owner != userID) {
		return passwords.ErrUnauthorized
	}
	if err != nil {
		return fmt.Errorf("consumir o convite: %w", err)
	}
	if err := q.SetPasswordHash(ctx, db.SetPasswordHashParams{ID: userID, PasswordHash: text(hash)}); err != nil {
		return fmt.Errorf("gravar a senha: %w", err)
	}
	if err := q.DeleteSessionsByUser(ctx, userID); err != nil {
		return fmt.Errorf("encerrar as sessoes antigas: %w", err)
	}
	return tx.Commit(ctx)
}
