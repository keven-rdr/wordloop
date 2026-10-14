//go:build integration

package postgres_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/keven-rdr/wordloop/api/internal/identity/accounts"
	"github.com/keven-rdr/wordloop/api/internal/identity/adapters/postgres"
	"github.com/keven-rdr/wordloop/api/internal/identity/domain"
)

func TestUpsertLinksRowsCreatedBeforeTheIdentityTable(t *testing.T) {
	pool := setupPool(t)
	ctx := context.Background()
	var legacyID uuid.UUID
	err := pool.QueryRow(ctx, `INSERT INTO app_user (idp_issuer, idp_subject, display_name) VALUES ('https://idp.test', 'antigo', 'Antigo') RETURNING id`).Scan(&legacyID)
	if err != nil {
		t.Fatal(err)
	}
	users := postgres.NewUsers(pool)
	got, err := users.Upsert(ctx, domain.Identity{Issuer: "https://idp.test", Subject: "antigo"})
	if err != nil || got.ID != legacyID || got.DisplayName != "Antigo" {
		t.Fatalf("o login deve reaproveitar a linha antiga: %+v %v", got, err)
	}
	var links int
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM auth_identity WHERE user_id = $1`, legacyID).Scan(&links); err != nil || links != 1 {
		t.Fatalf("o vinculo deve ser criado na hora: %d %v", links, err)
	}
}

func TestSimultaneousFirstLoginsCreateASingleProfile(t *testing.T) {
	pool := setupPool(t)
	users := postgres.NewUsers(pool)
	ident := domain.Identity{Issuer: "https://idp.test", Subject: "corrida", DisplayName: "Corrida"}
	ids := make(chan uuid.UUID, 8)
	var wg sync.WaitGroup
	for range 8 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			u, err := users.Upsert(context.Background(), ident)
			if err != nil {
				t.Error(err)
				return
			}
			ids <- u.ID
		}()
	}
	wg.Wait()
	close(ids)
	first := uuid.Nil
	for id := range ids {
		if first == uuid.Nil {
			first = id
		}
		if id != first {
			t.Fatalf("todos os logins devem resolver para o mesmo perfil: %v e %v", first, id)
		}
	}
	var profiles, links int
	if err := pool.QueryRow(context.Background(), `SELECT (SELECT count(*) FROM app_user), (SELECT count(*) FROM auth_identity)`).Scan(&profiles, &links); err != nil {
		t.Fatal(err)
	}
	if profiles != 1 || links != 1 {
		t.Fatalf("esperava 1 perfil e 1 vinculo, vieram %d e %d", profiles, links)
	}
}

func TestAccountsStore(t *testing.T) {
	pool := setupPool(t)
	ctx := context.Background()
	store := postgres.NewAccounts(pool)
	at := time.Now().Truncate(time.Millisecond)

	if _, err := store.FindByEmail(ctx, "a@b.co"); !errors.Is(err, accounts.ErrNotFound) {
		t.Fatalf("conta inexistente: esperava ErrNotFound, veio %v", err)
	}
	id, err := store.CreateAdmin(ctx, "a@b.co", "Admin", at)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.CreateAdmin(ctx, "A@B.CO", "Outro", at); !errors.Is(err, accounts.ErrEmailTaken) {
		t.Fatalf("e-mail repetido (maiusculas): esperava ErrEmailTaken, veio %v", err)
	}
	acc, err := store.FindByEmail(ctx, "A@b.co")
	if err != nil || acc.ID != id || acc.Role != domain.RoleAdmin || acc.HasCredentials {
		t.Fatalf("FindByEmail: %+v %v", acc, err)
	}
	if err := store.ReplaceInvite(ctx, id, []byte("hash-1"), at, at.Add(time.Minute)); err != nil {
		t.Fatal(err)
	}
	if _, err := pool.Exec(ctx, `UPDATE auth_token SET used_at = now() WHERE id_hash = 'hash-1'`); err != nil {
		t.Fatal(err)
	}
	if err := store.ReplaceInvite(ctx, id, []byte("hash-2"), at, at.Add(time.Minute)); err != nil {
		t.Fatal(err)
	}
	if err := store.ReplaceInvite(ctx, id, []byte("hash-3"), at, at.Add(time.Minute)); err != nil {
		t.Fatal(err)
	}
	var total, unused int
	err = pool.QueryRow(ctx, `SELECT count(*), count(*) FILTER (WHERE used_at IS NULL) FROM auth_token WHERE user_id = $1 AND purpose = 'invite'`, id).Scan(&total, &unused)
	if err != nil || total != 2 || unused != 1 {
		t.Fatalf("o convite usado fica, o nao usado anterior some e so o novo vale: total=%d sem uso=%d (%v)", total, unused, err)
	}

	if _, err := pool.Exec(ctx, `INSERT INTO auth_identity (user_id, provider, subject) VALUES ($1, 'https://idp.test', 'x')`, id); err != nil {
		t.Fatal(err)
	}
	if acc, _ := store.FindByEmail(ctx, "a@b.co"); !acc.HasCredentials {
		t.Fatal("com vinculo de provedor a conta ja tem como entrar")
	}
}

func TestStoresReportDatabaseFailures(t *testing.T) {
	pool := setupPool(t)
	pool.Close()
	ctx := context.Background()
	store := postgres.NewAccounts(pool)
	if _, err := store.FindByEmail(ctx, "a@b.co"); err == nil || errors.Is(err, accounts.ErrNotFound) {
		t.Fatalf("falha do banco nao e 'conta nao encontrada': %v", err)
	}
	if _, err := store.CreateAdmin(ctx, "a@b.co", "A", time.Now()); err == nil || errors.Is(err, accounts.ErrEmailTaken) {
		t.Fatalf("falha do banco nao e 'e-mail ja cadastrado': %v", err)
	}
	if err := store.ReplaceInvite(ctx, uuid.New(), []byte("h"), time.Now(), time.Now().Add(time.Minute)); err == nil {
		t.Fatal("falha do banco deve voltar como erro")
	}
	if _, err := postgres.NewUsers(pool).Upsert(ctx, domain.Identity{Issuer: "i", Subject: "s"}); err == nil {
		t.Fatal("falha do banco no login deve voltar como erro")
	}
}

func TestSetMFARequiredPersistsAndShowsUpInTheProfile(t *testing.T) {
	pool := setupPool(t)
	ctx := context.Background()
	store, users := postgres.NewAccounts(pool), postgres.NewUsers(pool)
	id, err := store.CreateAdmin(ctx, "dono@exemplo.com", "Dono", time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if got, err := users.Get(ctx, id); err != nil || got.MFARequired {
		t.Fatalf("o 2FA vem desligado: %+v %v", got, err)
	}
	if err := store.SetMFARequired(ctx, id, true); err != nil {
		t.Fatal(err)
	}
	if got, err := users.Get(ctx, id); err != nil || !got.MFARequired {
		t.Fatalf("depois de ligar: %+v %v", got, err)
	}
	if err := store.SetMFARequired(ctx, uuid.New(), true); !errors.Is(err, accounts.ErrNotFound) {
		t.Fatalf("conta que nao existe: esperava ErrNotFound, veio %v", err)
	}
}
