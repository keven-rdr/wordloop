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
	"github.com/keven-rdr/wordloop/api/internal/identity/domain"
	"github.com/keven-rdr/wordloop/api/internal/identity/passkeys"
	"github.com/keven-rdr/wordloop/api/internal/platform/token"
)

type pkEnv struct {
	pool  *pgxpool.Pool
	store *postgres.Passkeys
	user  uuid.UUID
	at    time.Time
}

func newPkEnv(t *testing.T) pkEnv {
	t.Helper()
	pool := setupPool(t)
	ctx := context.Background()
	u, err := postgres.NewUsers(pool).Upsert(ctx, domain.Identity{Issuer: "i", Subject: "s", DisplayName: "Voce"})
	if err != nil {
		t.Fatal(err)
	}
	return pkEnv{pool: pool, store: postgres.NewPasskeys(pool), user: u.ID, at: time.Now().Truncate(time.Millisecond)}
}

func (e pkEnv) cred(id string) passkeys.Credential {
	return passkeys.Credential{
		CredentialID: []byte("cred-" + id), UserID: e.user, PublicKey: []byte("pk-" + id), SignCount: 3, Transports: []string{"internal", "hybrid"},
		UserVerified: true, BackupEligible: true, AAGUID: []byte("0123456789abcdef"), Name: "Passkey " + id, CreatedAt: e.at,
	}
}

func (e pkEnv) invite(t *testing.T, tok string, expires time.Time) []byte {
	t.Helper()
	hash := token.Hash(tok)
	err := postgres.NewAccounts(e.pool).ReplaceInvite(context.Background(), e.user, hash, e.at, expires)
	if err != nil {
		t.Fatal(err)
	}
	return hash
}

func TestPasskeyStoreRoundTrip(t *testing.T) {
	e := newPkEnv(t)
	ctx := context.Background()
	if err := e.store.Add(ctx, e.cred("a")); err != nil {
		t.Fatal(err)
	}
	if err := e.store.Add(ctx, e.cred("a")); !errors.Is(err, passkeys.ErrDuplicate) {
		t.Fatalf("credencial repetida: %v", err)
	}
	list, err := e.store.List(ctx, e.user)
	if err != nil || len(list) != 1 || list[0].Name != "Passkey a" || list[0].SignCount != 3 || len(list[0].Transports) != 2 || !list[0].BackupEligible {
		t.Fatalf("lista: %+v %v", list, err)
	}
	found, owner, err := e.store.Find(ctx, []byte("cred-a"))
	if err != nil || found.ID != list[0].ID || owner.ID != e.user || owner.DisplayName != "Voce" {
		t.Fatalf("Find: %+v %+v %v", found, owner, err)
	}
	if _, _, err := e.store.Find(ctx, []byte("nao-existe")); !errors.Is(err, passkeys.ErrNotFound) {
		t.Fatalf("credencial inexistente: %v", err)
	}
	later := e.at.Add(time.Hour)
	updated := e.cred("a")
	updated.SignCount = 9
	if err := e.store.RecordLogin(ctx, updated, later); err != nil {
		t.Fatal(err)
	}
	got, _, _ := e.store.Find(ctx, []byte("cred-a"))
	if got.SignCount != 9 || !got.LastUsedAt.Equal(later) {
		t.Fatalf("RecordLogin: %+v", got)
	}
	var sawKeys, sawIdentities int
	if err := e.store.Delete(ctx, uuid.New(), got.ID, anyone); !errors.Is(err, passkeys.ErrNotFound) {
		t.Fatalf("apagar passkey de outra pessoa: %v", err)
	}
	if err := e.store.Delete(ctx, e.user, uuid.New(), anyone); !errors.Is(err, passkeys.ErrNotFound) {
		t.Fatalf("passkey inexistente: %v", err)
	}
	refuse := func(keys, identities int) bool { sawKeys, sawIdentities = keys, identities; return false }
	if err := e.store.Delete(ctx, e.user, got.ID, refuse); !errors.Is(err, passkeys.ErrLastMethod) || sawKeys != 1 || sawIdentities != 1 {
		t.Fatalf("a regra deve ver 1 passkey e 1 vinculo e poder recusar: %v keys=%d ids=%d", err, sawKeys, sawIdentities)
	}
	if err := e.store.Delete(ctx, e.user, got.ID, anyone); err != nil {
		t.Fatal(err)
	}
	if list, _ := e.store.List(ctx, e.user); len(list) != 0 {
		t.Fatal("a passkey deveria ter sido apagada")
	}
}

func anyone(int, int) bool { return true }

func TestConcurrentDeletesCannotLeaveAnAdminWithoutPasskeys(t *testing.T) {
	e := newPkEnv(t)
	ctx := context.Background()
	for _, id := range []string{"a", "b"} {
		if err := e.store.Add(ctx, e.cred(id)); err != nil {
			t.Fatal(err)
		}
	}
	list, _ := e.store.List(ctx, e.user)
	adminRule := func(keys, identities int) bool { return keys > 1 }
	results := make(chan error, len(list))
	for _, c := range list {
		go func() { results <- e.store.Delete(ctx, e.user, c.ID, adminRule) }()
	}
	var ok, refused int
	for range list {
		switch err := <-results; {
		case err == nil:
			ok++
		case errors.Is(err, passkeys.ErrLastMethod):
			refused++
		default:
			t.Fatalf("erro inesperado: %v", err)
		}
	}
	if left, _ := e.store.List(ctx, e.user); ok != 1 || refused != 1 || len(left) != 1 {
		t.Fatalf("exatamente um apagar deve funcionar: ok=%d recusados=%d restam=%d", ok, refused, len(left))
	}
}

func TestPasskeyLimitIsEnforcedInTheDatabaseLayer(t *testing.T) {
	e := newPkEnv(t)
	ctx := context.Background()
	for i := range passkeys.MaxPerAccount {
		if err := e.store.Add(ctx, e.cred(string(rune('a'+i)))); err != nil {
			t.Fatal(err)
		}
	}
	if err := e.store.Add(ctx, e.cred("extra")); !errors.Is(err, passkeys.ErrTooMany) {
		t.Fatalf("passou do limite: %v", err)
	}
}

func TestAddWithInviteConsumesTheInviteAtomically(t *testing.T) {
	e := newPkEnv(t)
	ctx := context.Background()
	hash := e.invite(t, "tok", e.at.Add(15*time.Minute))
	if got, err := e.store.PeekInvite(ctx, hash, e.at); err != nil || got != e.user {
		t.Fatalf("PeekInvite: %v %v", got, err)
	}
	if err := e.store.AddWithInvite(ctx, e.cred("a"), hash, e.at); err != nil {
		t.Fatal(err)
	}
	if err := e.store.AddWithInvite(ctx, e.cred("b"), hash, e.at); !errors.Is(err, passkeys.ErrUnauthorized) {
		t.Fatalf("convite usado nao vale de novo: %v", err)
	}
	if _, err := e.store.PeekInvite(ctx, hash, e.at); !errors.Is(err, passkeys.ErrUnauthorized) {
		t.Fatalf("PeekInvite de convite usado: %v", err)
	}
	if list, _ := e.store.List(ctx, e.user); len(list) != 1 {
		t.Fatalf("so a primeira passkey fica: %d", len(list))
	}
}

func TestAddWithInviteDoesNotBurnTheInviteWhenTheInsertFails(t *testing.T) {
	e := newPkEnv(t)
	ctx := context.Background()
	if err := e.store.Add(ctx, e.cred("a")); err != nil {
		t.Fatal(err)
	}
	hash := e.invite(t, "tok", e.at.Add(15*time.Minute))
	if err := e.store.AddWithInvite(ctx, e.cred("a"), hash, e.at); !errors.Is(err, passkeys.ErrDuplicate) {
		t.Fatalf("credencial repetida: %v", err)
	}
	if _, err := e.store.PeekInvite(ctx, hash, e.at); err != nil {
		t.Fatalf("a transacao deve desfazer o consumo do convite: %v", err)
	}
}

func TestInviteRejectsExpiredAndForeignUsers(t *testing.T) {
	e := newPkEnv(t)
	ctx := context.Background()
	expired := e.invite(t, "velho", e.at.Add(-time.Second))
	if _, err := e.store.PeekInvite(ctx, expired, e.at); !errors.Is(err, passkeys.ErrUnauthorized) {
		t.Fatalf("convite vencido: %v", err)
	}
	if err := e.store.AddWithInvite(ctx, e.cred("a"), expired, e.at); !errors.Is(err, passkeys.ErrUnauthorized) {
		t.Fatalf("convite vencido no registro: %v", err)
	}
	other, _ := postgres.NewUsers(e.pool).Upsert(ctx, domain.Identity{Issuer: "i", Subject: "outra"})
	valid := e.invite(t, "valido", e.at.Add(time.Minute))
	foreign := e.cred("b")
	foreign.UserID = other.ID
	if err := e.store.AddWithInvite(ctx, foreign, valid, e.at); !errors.Is(err, passkeys.ErrUnauthorized) {
		t.Fatalf("convite de outra conta: %v", err)
	}
}
