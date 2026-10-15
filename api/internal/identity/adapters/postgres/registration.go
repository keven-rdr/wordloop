package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/keven-rdr/wordloop/api/internal/identity/adapters/postgres/db"
	"github.com/keven-rdr/wordloop/api/internal/identity/domain"
	"github.com/keven-rdr/wordloop/api/internal/identity/registration"
)

type Registration struct{ q *db.Queries }

var _ registration.Store = (*Registration)(nil)

func NewRegistration(conn db.DBTX) *Registration { return &Registration{q: db.New(conn)} }

func (r *Registration) FindByEmail(ctx context.Context, email string) (registration.Account, error) {
	row, err := r.q.FindRegistrationAccount(ctx, email)
	if errors.Is(err, pgx.ErrNoRows) {
		return registration.Account{}, registration.ErrNotFound
	}
	if err != nil {
		return registration.Account{}, fmt.Errorf("buscar a conta: %w", err)
	}
	return registration.Account{ID: row.ID, Role: domain.ParseRole(row.Role), Verified: row.Verified}, nil
}

func (r *Registration) CreatePending(ctx context.Context, email, displayName string) (uuid.UUID, error) {
	id, err := r.q.CreatePendingUser(ctx, db.CreatePendingUserParams{DisplayName: text(displayName), Email: text(email)})
	if isUniqueViolation(err) {
		return uuid.Nil, registration.ErrEmailTaken
	}
	if err != nil {
		return uuid.Nil, fmt.Errorf("criar a conta pendente: %w", err)
	}
	return id, nil
}

func (r *Registration) RecentTokens(ctx context.Context, userID uuid.UUID, since time.Time) (int, error) {
	n, err := r.q.CountRecentTokens(ctx, db.CountRecentTokensParams{UserID: userID, CreatedAt: ts(since)})
	if err != nil {
		return 0, fmt.Errorf("contar os tokens recentes: %w", err)
	}
	return int(n), nil
}

func (r *Registration) ReplaceToken(ctx context.Context, userID uuid.UUID, purpose domain.TokenPurpose, hash []byte, createdAt, expiresAt time.Time) error {
	err := r.q.ReplaceToken(ctx, db.ReplaceTokenParams{UserID: userID, Purpose: string(purpose), IDHash: hash, CreatedAt: ts(createdAt), ExpiresAt: ts(expiresAt)})
	if err != nil {
		return fmt.Errorf("gravar o token: %w", err)
	}
	return nil
}
