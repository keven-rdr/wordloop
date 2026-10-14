//go:build integration

package postgres_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/keven-rdr/wordloop/api/internal/identity/accounts"
	"github.com/keven-rdr/wordloop/api/internal/identity/adapters/postgres"
	"github.com/keven-rdr/wordloop/api/internal/identity/passkeys"
	"github.com/keven-rdr/wordloop/api/internal/platform/token"
)

func TestChallengesAreSingleUseTypedAndExpire(t *testing.T) {
	e := newPkEnv(t)
	ctx := context.Background()
	save := func(id string, p passkeys.Purpose, expires time.Time) []byte {
		hash := token.Hash(id)
		err := e.store.Save(ctx, passkeys.Challenge{IDHash: hash, Purpose: p, UserID: e.user, InviteHash: []byte("inv"), State: []byte(`{"c":1}`), CreatedAt: e.at, ExpiresAt: expires})
		if err != nil {
			t.Fatal(err)
		}
		return hash
	}
	hash := save("a", passkeys.PurposeRegister, e.at.Add(5*time.Minute))
	if _, err := e.store.Take(ctx, hash, passkeys.PurposeLogin, e.at); !errors.Is(err, passkeys.ErrNotFound) {
		t.Fatalf("finalidade errada: %v", err)
	}
	got, err := e.store.Take(ctx, hash, passkeys.PurposeRegister, e.at)
	if err != nil || got.UserID != e.user || string(got.InviteHash) != "inv" || string(got.State) != `{"c":1}` {
		t.Fatalf("Take: %+v %v", got, err)
	}
	if _, err := e.store.Take(ctx, hash, passkeys.PurposeRegister, e.at); !errors.Is(err, passkeys.ErrNotFound) {
		t.Fatalf("desafio de uso unico: %v", err)
	}
	old := save("velho", passkeys.PurposeLogin, e.at.Add(-time.Second))
	if _, err := e.store.Take(ctx, old, passkeys.PurposeLogin, e.at); !errors.Is(err, passkeys.ErrNotFound) {
		t.Fatalf("desafio vencido: %v", err)
	}
	loginHash := token.Hash("anonimo")
	if err := e.store.Save(ctx, passkeys.Challenge{IDHash: loginHash, Purpose: passkeys.PurposeLogin, State: []byte("{}"), CreatedAt: e.at, ExpiresAt: e.at.Add(time.Minute)}); err != nil {
		t.Fatalf("desafio de login nao tem usuario: %v", err)
	}
	if ch, err := e.store.Take(ctx, loginHash, passkeys.PurposeLogin, e.at); err != nil || ch.UserID != uuid.Nil {
		t.Fatalf("desafio de login: %+v %v", ch, err)
	}
	var left int
	if err := e.pool.QueryRow(ctx, `SELECT count(*) FROM auth_challenge`).Scan(&left); err != nil || left != 0 {
		t.Fatalf("ao guardar, os vencidos sao purgados: restam %d (%v)", left, err)
	}
}

func TestOwnerAndHasCredentialsSeePasskeys(t *testing.T) {
	e := newPkEnv(t)
	ctx := context.Background()
	if _, err := e.store.Owner(ctx, uuid.New()); !errors.Is(err, passkeys.ErrNotFound) {
		t.Fatalf("dono inexistente: %v", err)
	}
	if o, err := e.store.Owner(ctx, e.user); err != nil || o.ID != e.user {
		t.Fatalf("Owner: %+v %v", o, err)
	}
	store := postgres.NewAccounts(e.pool)
	if _, err := e.pool.Exec(ctx, `UPDATE app_user SET email = 'voce@exemplo.com', role = 'admin' WHERE id = $1`, e.user); err != nil {
		t.Fatal(err)
	}
	if _, err := e.pool.Exec(ctx, `DELETE FROM auth_identity WHERE user_id = $1`, e.user); err != nil {
		t.Fatal(err)
	}
	if acc, _ := store.FindByEmail(ctx, "voce@exemplo.com"); acc.HasCredentials {
		t.Fatal("antes da passkey o admin ainda nao tem como entrar")
	}
	if err := e.store.Add(ctx, e.cred("a")); err != nil {
		t.Fatal(err)
	}
	acc, err := store.FindByEmail(ctx, "voce@exemplo.com")
	if err != nil || !acc.HasCredentials {
		t.Fatalf("com passkey a conta ja tem como entrar (create-admin nao pode reemitir convite): %+v %v", acc, err)
	}
	if _, err := accounts.NormalizeEmail("voce@exemplo.com"); err != nil {
		t.Fatal(err)
	}
}

func TestPasskeyStoreReportsDatabaseFailures(t *testing.T) {
	e := newPkEnv(t)
	e.pool.Close()
	ctx := context.Background()
	if _, err := e.store.List(ctx, e.user); err == nil {
		t.Error("List deve falhar com o banco fechado")
	}
	if _, _, err := e.store.Find(ctx, []byte("x")); err == nil || errors.Is(err, passkeys.ErrNotFound) {
		t.Errorf("Find: falha do banco nao e 'nao encontrado': %v", err)
	}
	if _, err := e.store.Owner(ctx, e.user); err == nil || errors.Is(err, passkeys.ErrNotFound) {
		t.Errorf("Owner: %v", err)
	}
	if _, err := e.store.PeekInvite(ctx, []byte("x"), e.at); err == nil || errors.Is(err, passkeys.ErrUnauthorized) {
		t.Errorf("PeekInvite: %v", err)
	}
	if err := e.store.Add(ctx, e.cred("a")); err == nil {
		t.Error("Add deve falhar")
	}
	if err := e.store.AddWithInvite(ctx, e.cred("a"), []byte("x"), e.at); err == nil {
		t.Error("AddWithInvite deve falhar")
	}
	if err := e.store.RecordLogin(ctx, e.cred("a"), e.at); err == nil {
		t.Error("RecordLogin deve falhar")
	}
	if err := e.store.Delete(ctx, e.user, uuid.New(), anyone); err == nil || errors.Is(err, passkeys.ErrNotFound) {
		t.Errorf("Delete: %v", err)
	}
	if err := e.store.Save(ctx, passkeys.Challenge{IDHash: []byte("x"), Purpose: passkeys.PurposeLogin, CreatedAt: e.at, ExpiresAt: e.at}); err == nil {
		t.Error("Save deve falhar")
	}
	if _, err := e.store.Take(ctx, []byte("x"), passkeys.PurposeLogin, e.at); err == nil || errors.Is(err, passkeys.ErrNotFound) {
		t.Errorf("Take: %v", err)
	}
}
