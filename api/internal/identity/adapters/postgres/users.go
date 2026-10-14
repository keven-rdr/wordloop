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

func (r *Users) Upsert(ctx context.Context, id domain.Identity) (domain.User, error) {
	uid, err := r.resolve(ctx, id)
	if err != nil {
		return domain.User{}, fmt.Errorf("resolver o perfil: %w", err)
	}
	row, err := r.q.TouchUser(ctx, db.TouchUserParams{ID: uid, DisplayName: id.DisplayName})
	if err != nil {
		return domain.User{}, fmt.Errorf("atualizar o perfil: %w", err)
	}
	return domain.User{ID: row.ID, DisplayName: row.DisplayName.String, Role: domain.ParseRole(row.Role)}, nil
}

func (r *Users) resolve(ctx context.Context, id domain.Identity) (uuid.UUID, error) {
	for range 2 {
		uid, err := r.q.GetUserIDByIdentity(ctx, db.GetUserIDByIdentityParams{Provider: id.Issuer, Subject: id.Subject})
		if err == nil {
			return uid, nil
		}
		if !errors.Is(err, pgx.ErrNoRows) {
			return uuid.Nil, err
		}
		if uid, ok, err := r.linkLegacy(ctx, id); err != nil || ok {
			return uid, err
		}
		uid, err = r.q.CreateUserWithIdentity(ctx, db.CreateUserWithIdentityParams{
			IdpIssuer: text(id.Issuer), IdpSubject: text(id.Subject), DisplayName: text(id.DisplayName),
		})
		if err == nil || !isUniqueViolation(err) {
			return uid, err
		}
	}
	return uuid.Nil, errors.New("conflito repetido ao criar o perfil")
}

func (r *Users) linkLegacy(ctx context.Context, id domain.Identity) (uuid.UUID, bool, error) {
	uid, err := r.q.GetUserIDByLegacyIdentity(ctx, db.GetUserIDByLegacyIdentityParams{IdpIssuer: text(id.Issuer), IdpSubject: text(id.Subject)})
	if errors.Is(err, pgx.ErrNoRows) {
		return uuid.Nil, false, nil
	}
	if err != nil {
		return uuid.Nil, false, err
	}
	err = r.q.LinkIdentity(ctx, db.LinkIdentityParams{UserID: uid, Provider: id.Issuer, Subject: id.Subject})
	return uid, err == nil, err
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
	return domain.User{ID: row.ID, DisplayName: row.DisplayName.String, Role: domain.ParseRole(row.Role)}, nil
}
