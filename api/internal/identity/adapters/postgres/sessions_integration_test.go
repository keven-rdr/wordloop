//go:build integration

package postgres_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/keven-rdr/wordloop/api/internal/identity/adapters/postgres"
	"github.com/keven-rdr/wordloop/api/internal/identity/app"
	"github.com/keven-rdr/wordloop/api/internal/identity/domain"
)

func newSession(userID uuid.UUID, id string, at time.Time) app.NewSession {
	return app.NewSession{
		IDHash: []byte("hash-" + id), UserID: userID, RefreshTokenEnc: "sel", CreatedAt: at, ExpiresAt: at.Add(time.Hour), AuthenticatedAt: at,
	}
}

func TestNewUserIsAlwaysARegularUserAndLoginNeverChangesRole(t *testing.T) {
	pool := setupPool(t)
	users := postgres.NewUsers(pool)
	ctx := context.Background()
	ident := domain.Identity{Issuer: "i", Subject: "s", DisplayName: "A"}
	user, err := users.Upsert(ctx, ident)
	if err != nil || user.Role != domain.RoleUser {
		t.Fatalf("o papel inicial deve ser user: %+v %v", user, err)
	}
	// Quem vira admin so o faz por SQL do dono; um novo login (upsert) nao pode desfazer nem conceder isso.
	if _, err := pool.Exec(ctx, "UPDATE app_user SET role = 'admin' WHERE id = $1", user.ID); err != nil {
		t.Fatal(err)
	}
	again, err := users.Upsert(ctx, ident)
	if err != nil || again.Role != domain.RoleAdmin {
		t.Fatalf("o login nao deve mexer no papel: %+v %v", again, err)
	}
	if got, _ := users.Get(ctx, user.ID); got.Role != domain.RoleAdmin {
		t.Fatalf("Get deve devolver o papel do banco: %+v", got)
	}
}

func TestListAndDeleteAreScopedToTheOwner(t *testing.T) {
	users, sessions := setup(t)
	ctx := context.Background()
	alice, _ := users.Upsert(ctx, domain.Identity{Issuer: "i", Subject: "alice"})
	bob, _ := users.Upsert(ctx, domain.Identity{Issuer: "i", Subject: "bob"})
	at := time.Now().Truncate(time.Millisecond)
	for _, n := range []app.NewSession{newSession(alice.ID, "a1", at), newSession(alice.ID, "a2", at.Add(time.Minute)), newSession(bob.ID, "b1", at)} {
		if err := sessions.Create(ctx, n); err != nil {
			t.Fatal(err)
		}
	}
	mine, err := sessions.ListByUser(ctx, alice.ID)
	if err != nil || len(mine) != 2 || !mine[0].LastUsedAt.After(mine[1].LastUsedAt) {
		t.Fatalf("a lista da Alice deve ter 2 sessoes, da mais recente para a mais antiga: %+v %v", mine, err)
	}
	bobs, _ := sessions.ListByUser(ctx, bob.ID)
	if len(bobs) != 1 {
		t.Fatalf("a lista do Bob deve ter 1 sessao, veio %d", len(bobs))
	}

	// IDOR: o Bob tenta apagar a sessao da Alice pelo id publico dela.
	if err := sessions.DeleteByPublicID(ctx, bob.ID, mine[0].PublicID); !errors.Is(err, app.ErrNotFound) {
		t.Fatalf("apagar sessao alheia: esperava ErrNotFound, veio %v", err)
	}
	if again, _ := sessions.ListByUser(ctx, alice.ID); len(again) != 2 {
		t.Fatal("a sessao da Alice nao pode ser apagada pelo Bob")
	}
	if err := sessions.DeleteByPublicID(ctx, alice.ID, mine[0].PublicID); err != nil {
		t.Fatalf("a dona deve poder apagar a propria sessao: %v", err)
	}
	if err := sessions.DeleteByPublicID(ctx, alice.ID, uuid.New()); !errors.Is(err, app.ErrNotFound) {
		t.Fatalf("id inexistente: esperava ErrNotFound, veio %v", err)
	}
}

func TestDeleteAllByUserLeavesOthersAlone(t *testing.T) {
	users, sessions := setup(t)
	ctx := context.Background()
	alice, _ := users.Upsert(ctx, domain.Identity{Issuer: "i", Subject: "alice"})
	bob, _ := users.Upsert(ctx, domain.Identity{Issuer: "i", Subject: "bob"})
	at := time.Now().Truncate(time.Millisecond)
	for _, n := range []app.NewSession{newSession(alice.ID, "a1", at), newSession(alice.ID, "a2", at), newSession(bob.ID, "b1", at)} {
		if err := sessions.Create(ctx, n); err != nil {
			t.Fatal(err)
		}
	}
	if err := sessions.DeleteAllByUser(ctx, alice.ID); err != nil {
		t.Fatal(err)
	}
	if a, _ := sessions.ListByUser(ctx, alice.ID); len(a) != 0 {
		t.Fatalf("a Alice nao deve ter sessoes, tem %d", len(a))
	}
	if b, _ := sessions.ListByUser(ctx, bob.ID); len(b) != 1 {
		t.Fatalf("o Bob deve manter a sessao, tem %d", len(b))
	}
}
