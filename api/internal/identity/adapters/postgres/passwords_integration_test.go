//go:build integration

package postgres_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/keven-rdr/wordloop/api/internal/identity/accounts"
	"github.com/keven-rdr/wordloop/api/internal/identity/adapters/postgres"
	"github.com/keven-rdr/wordloop/api/internal/identity/passwords"
	"github.com/keven-rdr/wordloop/api/internal/platform/token"
)

func newAdmin(t *testing.T, pool *pgxpool.Pool, email string) uuid.UUID {
	t.Helper()
	id, err := postgres.NewAccounts(pool).CreateAdmin(context.Background(), email, "Dono", time.Now())
	if err != nil {
		t.Fatal(err)
	}
	return id
}

func attempt(t *testing.T, store *postgres.Passwords, id uuid.UUID, at time.Time) bool {
	t.Helper()
	allowed, err := store.BeginAttempt(context.Background(), id, at)
	if err != nil {
		t.Fatal(err)
	}
	return allowed
}

func lockOf(t *testing.T, store *postgres.Passwords) time.Time {
	t.Helper()
	acc, err := store.FindByEmail(context.Background(), "dono@exemplo.com")
	if err != nil {
		t.Fatal(err)
	}
	return acc.LockedUntil
}

func TestTheFifthAttemptStillChecksAndLocksTheNextOnesWithGrowingTime(t *testing.T) {
	pool := setupPool(t)
	ctx := context.Background()
	id := newAdmin(t, pool, "dono@exemplo.com")
	store := postgres.NewPasswords(pool)
	at := time.Now()
	for i := 1; i <= 4; i++ {
		if !attempt(t, store, id, at) || !lockOf(t, store).IsZero() {
			t.Fatalf("a tentativa %d confere e ainda nao bloqueia", i)
		}
	}
	if !attempt(t, store, id, at) {
		t.Fatal("a 5a tentativa ainda confere a senha")
	}
	first := lockOf(t, store)
	if d := first.Sub(at); d < 25*time.Second || d > 35*time.Second {
		t.Fatalf("a 5a tentativa errada bloqueia as proximas por ~30s, veio %v", d)
	}
	if attempt(t, store, id, at) || !lockOf(t, store).Equal(first) {
		t.Fatal("a 6a tentativa e recusada sem contar e sem mudar o bloqueio")
	}
	at = first.Add(time.Second)
	if !attempt(t, store, id, at) {
		t.Fatal("passado o bloqueio a tentativa confere a senha (a senha certa entra)")
	}
	if d := lockOf(t, store).Sub(at); d < 55*time.Second || d > 65*time.Second {
		t.Fatalf("o erro seguinte dobra para ~60s, veio %v", d)
	}
	for range 30 {
		at = lockOf(t, store).Add(time.Second)
		attempt(t, store, id, at)
	}
	if d := lockOf(t, store).Sub(at); d > 15*time.Minute+time.Second {
		t.Fatalf("o bloqueio nunca passa de 15 minutos: %v", d)
	}
	if err := store.RegisterSuccess(ctx, id); err != nil {
		t.Fatal(err)
	}
	if !lockOf(t, store).IsZero() {
		t.Fatal("o acerto libera a conta")
	}
}

func TestParallelAttemptsNeverGetMoreThanFiveChecks(t *testing.T) {
	pool := setupPool(t)
	id := newAdmin(t, pool, "dono@exemplo.com")
	store := postgres.NewPasswords(pool)
	at := time.Now()
	var wg sync.WaitGroup
	var mu sync.Mutex
	allowed := 0
	for range 20 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if attempt(t, store, id, at) {
				mu.Lock()
				allowed++
				mu.Unlock()
			}
		}()
	}
	wg.Wait()
	if allowed != 5 {
		t.Fatalf("de 20 tentativas simultaneas so 5 podem conferir a senha, vieram %d", allowed)
	}
}

func TestOldAttemptsDecayAfterTheWindow(t *testing.T) {
	pool := setupPool(t)
	ctx := context.Background()
	id := newAdmin(t, pool, "dono@exemplo.com")
	store := postgres.NewPasswords(pool)
	old := time.Now().Add(-time.Hour)
	for range 4 {
		attempt(t, store, id, old)
	}
	attempt(t, store, id, time.Now())
	var count int
	if err := pool.QueryRow(ctx, `SELECT failed_logins FROM app_user WHERE id = $1`, id).Scan(&count); err != nil || count != 1 {
		t.Fatalf("tentativas de mais de 15 minutos atras nao somam: %d %v", count, err)
	}
}

func TestUnknownEmailIsNotFoundAndAccountWithoutPasswordHasNoHash(t *testing.T) {
	pool := setupPool(t)
	ctx := context.Background()
	newAdmin(t, pool, "dono@exemplo.com")
	store := postgres.NewPasswords(pool)
	if _, err := store.FindByEmail(ctx, "ninguem@exemplo.com"); !errors.Is(err, passwords.ErrNotFound) {
		t.Fatalf("esperava ErrNotFound, veio %v", err)
	}
	if acc, err := store.FindByEmail(ctx, "dono@exemplo.com"); err != nil || acc.Hash != "" {
		t.Fatalf("sem senha o hash e vazio: %+v %v", acc, err)
	}
}

func putInvite(t *testing.T, pool *pgxpool.Pool, id uuid.UUID, secret string, expires time.Time) []byte {
	t.Helper()
	hash := token.Hash(secret)
	err := postgres.NewAccounts(pool).ReplaceInvite(context.Background(), id, hash, time.Now(), expires)
	if err != nil {
		t.Fatal(err)
	}
	return hash
}

func TestInviteSetsThePasswordInOneTransactionAndOnlyOnce(t *testing.T) {
	pool := setupPool(t)
	ctx := context.Background()
	id := newAdmin(t, pool, "dono@exemplo.com")
	hash := putInvite(t, pool, id, "convite", time.Now().Add(10*time.Minute))
	store := postgres.NewPasswords(pool)
	owner, err := store.PeekInvite(ctx, hash, time.Now())
	if err != nil || owner.ID != id || owner.Email != "dono@exemplo.com" {
		t.Fatalf("PeekInvite: %+v %v", owner, err)
	}
	if err := store.SetWithInvite(ctx, hash, id, "$argon2id$teste", time.Now()); err != nil {
		t.Fatal(err)
	}
	if acc, _ := store.FindByEmail(ctx, "dono@exemplo.com"); acc.Hash != "$argon2id$teste" {
		t.Fatalf("hash gravado: %+v", acc)
	}
	if err := store.SetWithInvite(ctx, hash, id, "$argon2id$outro", time.Now()); !errors.Is(err, passwords.ErrUnauthorized) {
		t.Fatalf("convite usado: esperava ErrUnauthorized, veio %v", err)
	}
	if _, err := store.PeekInvite(ctx, hash, time.Now()); !errors.Is(err, passwords.ErrNotFound) {
		t.Fatalf("convite usado nao aparece mais: %v", err)
	}
	if acc, _ := store.FindByEmail(ctx, "dono@exemplo.com"); acc.Hash != "$argon2id$teste" {
		t.Fatal("a segunda tentativa nao pode trocar a senha")
	}
}

func TestExpiredInviteAndSomeoneElsesInviteChangeNothing(t *testing.T) {
	pool := setupPool(t)
	ctx := context.Background()
	id := newAdmin(t, pool, "dono@exemplo.com")
	other := newAdmin(t, pool, "outro@exemplo.com")
	store := postgres.NewPasswords(pool)
	expired := putInvite(t, pool, id, "vencido", time.Now().Add(-time.Minute))
	if _, err := store.PeekInvite(ctx, expired, time.Now()); !errors.Is(err, passwords.ErrNotFound) {
		t.Fatalf("convite vencido: %v", err)
	}
	valid := putInvite(t, pool, other, "valido", time.Now().Add(time.Minute))
	if err := store.SetWithInvite(ctx, valid, id, "$argon2id$x", time.Now()); !errors.Is(err, passwords.ErrUnauthorized) {
		t.Fatalf("convite de outra pessoa: %v", err)
	}
	if acc, _ := store.FindByEmail(ctx, "dono@exemplo.com"); acc.Hash != "" {
		t.Fatal("a senha nao pode ser gravada")
	}
	if acc, _ := store.FindByEmail(ctx, "outro@exemplo.com"); acc.Hash != "" {
		t.Fatal("o convite de outra pessoa nao pode ter sido consumido por engano")
	}
	if _, err := store.PeekInvite(ctx, valid, time.Now()); err != nil {
		t.Fatalf("o convite valido segue valendo (a transacao desfez tudo): %v", err)
	}
}

func TestSimultaneousInviteUsesSetThePasswordExactlyOnce(t *testing.T) {
	pool := setupPool(t)
	ctx := context.Background()
	id := newAdmin(t, pool, "dono@exemplo.com")
	hash := putInvite(t, pool, id, "convite", time.Now().Add(10*time.Minute))
	store := postgres.NewPasswords(pool)
	var wg sync.WaitGroup
	results := make(chan error, 6)
	for i := range 6 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			results <- store.SetWithInvite(ctx, hash, id, "$argon2id$"+string(rune('a'+i)), time.Now())
		}()
	}
	wg.Wait()
	close(results)
	won := 0
	for err := range results {
		if err == nil {
			won++
		}
	}
	if won != 1 {
		t.Fatalf("exatamente um uso do convite deve vencer, vieram %d", won)
	}
}

func TestAccountWithAPasswordIsNoLongerAPendingAdmin(t *testing.T) {
	pool := setupPool(t)
	ctx := context.Background()
	id := newAdmin(t, pool, "dono@exemplo.com")
	accs := postgres.NewAccounts(pool)
	if acc, _ := accs.FindByEmail(ctx, "dono@exemplo.com"); acc.HasCredentials {
		t.Fatal("recem-criado, ainda sem como entrar")
	}
	hash := putInvite(t, pool, id, "convite", time.Now().Add(time.Minute))
	if err := postgres.NewPasswords(pool).SetWithInvite(ctx, hash, id, "$argon2id$x", time.Now()); err != nil {
		t.Fatal(err)
	}
	if acc, _ := accs.FindByEmail(ctx, "dono@exemplo.com"); !acc.HasCredentials {
		t.Fatal("com senha a conta ja tem como entrar (create-admin nao reemite convite)")
	}
	var _ accounts.Store = accs
}

func TestSettingThePasswordWithAnInviteEndsEveryOldSession(t *testing.T) {
	pool := setupPool(t)
	ctx := context.Background()
	id := newAdmin(t, pool, "dono@exemplo.com")
	_, err := pool.Exec(ctx, `INSERT INTO auth_session (id_hash, user_id, refresh_token_enc, created_at, last_used_at, expires_at)
		VALUES ('\x01', $1, '', now(), now(), now() + interval '1 hour'), ('\x02', $1, '', now(), now(), now() + interval '1 hour')`, id)
	if err != nil {
		t.Fatal(err)
	}
	hash := putInvite(t, pool, id, "convite", time.Now().Add(time.Minute))
	if err := postgres.NewPasswords(pool).SetWithInvite(ctx, hash, id, "$argon2id$x", time.Now()); err != nil {
		t.Fatal(err)
	}
	var sessions int
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM auth_session WHERE user_id = $1`, id).Scan(&sessions); err != nil || sessions != 0 {
		t.Fatalf("trocar a senha encerra as sessoes antigas: %d %v", sessions, err)
	}
}
