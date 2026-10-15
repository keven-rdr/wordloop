//go:build integration

package postgres_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/keven-rdr/wordloop/api/internal/identity/adapters/postgres"
	"github.com/keven-rdr/wordloop/api/internal/identity/app"
	"github.com/keven-rdr/wordloop/api/internal/identity/domain"
	"github.com/keven-rdr/wordloop/api/internal/platform/db"
	"github.com/keven-rdr/wordloop/api/internal/platform/db/dbtest"
	"github.com/keven-rdr/wordloop/api/migrations"
)

func setup(t *testing.T) (*postgres.Users, *postgres.Sessions) {
	t.Helper()
	pool := setupPool(t)
	return postgres.NewUsers(pool), postgres.NewSessions(pool)
}

func setupPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	ctx := context.Background()
	url, _ := dbtest.Postgres(t)
	pool, err := db.Open(ctx, url)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	if err := db.Migrate(ctx, pool, migrations.FS); err != nil {
		t.Fatal(err)
	}
	return pool
}

func TestUpsertCreatesOnceAndKeepsIdentityStable(t *testing.T) {
	users, _ := setup(t)
	ctx := context.Background()
	ident := domain.Identity{Issuer: "https://idp.test", Subject: "abc", Email: "abc" + "@exemplo.com", EmailVerified: true, DisplayName: "Aluno"}

	first, err := users.Upsert(ctx, ident)
	if err != nil || first.ID == uuid.Nil || first.DisplayName != "Aluno" {
		t.Fatalf("primeiro upsert: %+v %v", first, err)
	}
	ident.DisplayName = ""
	again, err := users.Upsert(ctx, ident)
	if err != nil || again.ID != first.ID || again.DisplayName != "Aluno" {
		t.Fatalf("o segundo login deve reaproveitar o perfil e manter o nome: %+v %v", again, err)
	}
	ident.DisplayName = "Aluno Novo"
	if renamed, _ := users.Upsert(ctx, ident); renamed.DisplayName != "Aluno Novo" || renamed.ID != first.ID {
		t.Fatalf("o nome novo deve atualizar o perfil: %+v", renamed)
	}
	other, err := users.Upsert(ctx, domain.Identity{Issuer: "https://idp.test", Subject: "outro", Email: "outro" + "@exemplo.com", EmailVerified: true})
	if err != nil || other.ID == first.ID {
		t.Fatalf("outro subject deve ser outro perfil: %+v %v", other, err)
	}
	sameSubjectOtherIssuer, err := users.Upsert(ctx, domain.Identity{Issuer: "https://outro.idp", Subject: "abc", Email: "abc@outro.idp.test", EmailVerified: true})
	if err != nil || sameSubjectOtherIssuer.ID == first.ID {
		t.Fatalf("o mesmo subject em outro emissor e outra pessoa: %+v %v", sameSubjectOtherIssuer, err)
	}
	got, err := users.Get(ctx, first.ID)
	if err != nil || got.ID != first.ID {
		t.Fatalf("Get: %+v %v", got, err)
	}
	if _, err := users.Get(ctx, uuid.New()); !errors.Is(err, app.ErrNotFound) {
		t.Fatalf("Get de perfil inexistente: esperava ErrNotFound, veio %v", err)
	}
}

func TestSessionLifecycle(t *testing.T) {
	users, sessions := setup(t)
	ctx := context.Background()
	user, _ := users.Upsert(ctx, domain.Identity{Issuer: "i", Subject: "s", Email: "s" + "@exemplo.com", EmailVerified: true})
	hash := []byte("0123456789abcdef0123456789abcdef")
	created := time.Now().Truncate(time.Millisecond)
	expires := created.Add(time.Hour)

	if _, err := sessions.Get(ctx, hash); !errors.Is(err, app.ErrNotFound) {
		t.Fatalf("sessao inexistente: esperava ErrNotFound, veio %v", err)
	}
	if err := sessions.Create(ctx, app.NewSession{
		IDHash: hash, UserID: user.ID, RefreshTokenEnc: "sel", CreatedAt: created, ExpiresAt: expires, AuthenticatedAt: created, MFA: true, UserAgent: "ua",
	}); err != nil {
		t.Fatal(err)
	}
	rec, err := sessions.Get(ctx, hash)
	if err != nil || rec.UserID != user.ID || !rec.ExpiresAt.Equal(expires) || !rec.CreatedAt.Equal(created) || !rec.LastUsedAt.Equal(created) || !rec.MFA || rec.PublicID == uuid.Nil {
		t.Fatalf("Get: %+v %v (esperava expiracao %v)", rec, err, expires)
	}
	later := created.Add(10 * time.Minute)
	if err := sessions.Touch(ctx, hash, later); err != nil {
		t.Fatal(err)
	}
	if touched, _ := sessions.Get(ctx, hash); !touched.LastUsedAt.Equal(later) || !touched.CreatedAt.Equal(created) {
		t.Fatalf("Touch deve mudar so o ultimo uso: %+v", touched)
	}
	if err := sessions.Delete(ctx, hash); err != nil {
		t.Fatal(err)
	}
	if _, err := sessions.Get(ctx, hash); !errors.Is(err, app.ErrNotFound) {
		t.Fatalf("depois de apagar: esperava ErrNotFound, veio %v", err)
	}
	if err := sessions.Delete(ctx, hash); err != nil {
		t.Fatalf("apagar duas vezes nao pode falhar: %v", err)
	}
	orphan := app.NewSession{IDHash: []byte("x"), UserID: uuid.New(), RefreshTokenEnc: "sel", CreatedAt: created, ExpiresAt: expires, AuthenticatedAt: created}
	if err := sessions.Create(ctx, orphan); err == nil {
		t.Fatal("sessao de perfil inexistente deve violar a chave estrangeira")
	}
}
