package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/keven-rdr/wordloop/api/internal/identity/accounts"
	"github.com/keven-rdr/wordloop/api/internal/identity/adapters/postgres/db"
	"github.com/keven-rdr/wordloop/api/internal/identity/domain"
)

type Accounts struct{ q *db.Queries }

var _ accounts.Store = (*Accounts)(nil)

func NewAccounts(conn db.DBTX) *Accounts { return &Accounts{q: db.New(conn)} }

func (r *Accounts) FindByEmail(ctx context.Context, email string) (accounts.Account, error) {
	row, err := r.q.FindAccountByEmail(ctx, email)
	if errors.Is(err, pgx.ErrNoRows) {
		return accounts.Account{}, accounts.ErrNotFound
	}
	if err != nil {
		return accounts.Account{}, fmt.Errorf("buscar a conta: %w", err)
	}
	return accounts.Account{ID: row.ID, Role: domain.ParseRole(row.Role), HasCredentials: row.HasCredentials}, nil
}

func (r *Accounts) CreateAdmin(ctx context.Context, email, name string, verifiedAt time.Time) (uuid.UUID, error) {
	id, err := r.q.CreateAdminUser(ctx, db.CreateAdminUserParams{DisplayName: text(name), Email: text(email), EmailVerifiedAt: ts(verifiedAt)})
	if isUniqueViolation(err) {
		return uuid.Nil, accounts.ErrEmailTaken
	}
	if err != nil {
		return uuid.Nil, fmt.Errorf("criar o admin: %w", err)
	}
	return id, nil
}

func (r *Accounts) ReplaceInvite(ctx context.Context, userID uuid.UUID, idHash []byte, createdAt, expiresAt time.Time) error {
	err := r.q.ReplaceInvite(ctx, db.ReplaceInviteParams{UserID: userID, IDHash: idHash, CreatedAt: ts(createdAt), ExpiresAt: ts(expiresAt)})
	if err != nil {
		return fmt.Errorf("gravar o convite: %w", err)
	}
	return nil
}
