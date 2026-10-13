package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/keven-rdr/wordloop/api/internal/identity/adapters/postgres/db"
	"github.com/keven-rdr/wordloop/api/internal/identity/app"
	"github.com/keven-rdr/wordloop/api/internal/identity/domain"
)

// Users implementa app.Users sobre o Postgres.
type Users struct{ q *db.Queries }

var _ app.Users = (*Users)(nil)

// NewUsers recebe o pool (ou uma transacao).
func NewUsers(conn db.DBTX) *Users { return &Users{q: db.New(conn)} }

func text(s string) pgtype.Text { return pgtype.Text{String: s, Valid: s != ""} }

// Upsert cria o perfil no primeiro login e atualiza o ultimo acesso nos seguintes.
func (r *Users) Upsert(ctx context.Context, id domain.Identity) (domain.User, error) {
	row, err := r.q.UpsertUser(ctx, db.UpsertUserParams{
		IdpIssuer: text(id.Issuer), IdpSubject: text(id.Subject), DisplayName: text(id.DisplayName),
	})
	if err != nil {
		return domain.User{}, fmt.Errorf("upsert do perfil: %w", err)
	}
	return domain.User{ID: row.ID, DisplayName: row.DisplayName.String}, nil
}

// Get le o perfil; app.ErrNotFound se nao existe ou foi removido.
func (r *Users) Get(ctx context.Context, id uuid.UUID) (domain.User, error) {
	row, err := r.q.GetUser(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.User{}, app.ErrNotFound
	}
	if err != nil {
		return domain.User{}, fmt.Errorf("ler o perfil: %w", err)
	}
	return domain.User{ID: row.ID, DisplayName: row.DisplayName.String}, nil
}
