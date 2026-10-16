//go:build integration

package postgres_test

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/keven-rdr/wordloop/api/internal/identity/adapters/postgres"
	"github.com/keven-rdr/wordloop/api/internal/identity/app"
	"github.com/keven-rdr/wordloop/api/internal/identity/domain"
)

func socialIdentity(provider, subject, email string, verified bool) domain.Identity {
	return domain.Identity{Issuer: provider, Subject: subject, DisplayName: "Ana", Email: email, EmailVerified: verified}
}

func count(t *testing.T, pool *pgxpool.Pool, sql string, args ...any) int {
	t.Helper()
	var n int
	if err := pool.QueryRow(context.Background(), sql, args...).Scan(&n); err != nil {
		t.Fatal(err)
	}
	return n
}

func TestFirstSocialLoginWithAVerifiedEmailCreatesAnAccountAndKeepsItOnTheNextOnes(t *testing.T) {
	pool := setupPool(t)
	ctx := context.Background()
	users := postgres.NewUsers(pool)
	first, err := users.Upsert(ctx, socialIdentity("google", "g-1", "Ana@Exemplo.com", true))
	if err != nil || first.Role != domain.RoleUser {
		t.Fatalf("primeiro login: %+v %v", first, err)
	}
	again, err := users.Upsert(ctx, socialIdentity("google", "g-1", "ana@exemplo.com", true))
	if err != nil || again.ID != first.ID {
		t.Fatalf("segundo login na mesma conta: %+v %v", again, err)
	}
	if n := count(t, pool, `SELECT count(*) FROM app_user WHERE email = 'ana@exemplo.com' AND email_verified_at IS NOT NULL`); n != 1 {
		t.Fatalf("e-mail em minusculas e confirmado: %d", n)
	}
	if n := count(t, pool, `SELECT count(*) FROM auth_identity WHERE provider = 'google' AND subject = 'g-1'`); n != 1 {
		t.Fatalf("vinculo: %d", n)
	}
}

func TestAVerifiedEmailLinksToTheExistingAccountInsteadOfCreatingAnother(t *testing.T) {
	pool := setupPool(t)
	ctx := context.Background()
	if _, err := pool.Exec(ctx, `INSERT INTO app_user (display_name, email, email_verified_at, role) VALUES ('Ana', 'ana@exemplo.com', now(), 'user')`); err != nil {
		t.Fatal(err)
	}
	var existing string
	_ = pool.QueryRow(ctx, `SELECT id::text FROM app_user WHERE email = 'ana@exemplo.com'`).Scan(&existing)
	got, err := postgres.NewUsers(pool).Upsert(ctx, socialIdentity("github", "gh-9", "ANA@exemplo.com", true))
	if err != nil || got.ID.String() != existing {
		t.Fatalf("deveria entrar na conta existente: %+v %v", got, err)
	}
	if n := count(t, pool, `SELECT count(*) FROM app_user`); n != 1 {
		t.Fatalf("nenhuma conta nova: %d", n)
	}
	if n := count(t, pool, `SELECT count(*) FROM auth_identity WHERE user_id = $1::uuid`, existing); n != 1 {
		t.Fatalf("vinculo criado: %d", n)
	}
}

func TestTwoProvidersWithTheSameEmailShareOneAccount(t *testing.T) {
	pool := setupPool(t)
	ctx := context.Background()
	users := postgres.NewUsers(pool)
	a, err1 := users.Upsert(ctx, socialIdentity("google", "g-1", "ana@exemplo.com", true))
	b, err2 := users.Upsert(ctx, socialIdentity("github", "gh-1", "ana@exemplo.com", true))
	if err1 != nil || err2 != nil || a.ID != b.ID {
		t.Fatalf("mesma conta: %+v %+v %v %v", a, b, err1, err2)
	}
	if n := count(t, pool, `SELECT count(*) FROM auth_identity WHERE user_id = $1`, a.ID); n != 2 {
		t.Fatalf("dois vinculos: %d", n)
	}
}

func TestLinkingConfirmsAPendingSignUpSoSquattingTheEmailGainsNothing(t *testing.T) {
	pool := setupPool(t)
	ctx := context.Background()
	if _, err := pool.Exec(ctx, `INSERT INTO app_user (display_name, email, role) VALUES ('ana', 'ana@exemplo.com', 'user')`); err != nil {
		t.Fatal(err)
	}
	if _, err := postgres.NewUsers(pool).Upsert(ctx, socialIdentity("google", "g-1", "ana@exemplo.com", true)); err != nil {
		t.Fatal(err)
	}
	if n := count(t, pool, `SELECT count(*) FROM app_user WHERE email = 'ana@exemplo.com' AND email_verified_at IS NOT NULL AND password_hash IS NULL`); n != 1 {
		t.Fatalf("a conta pendente foi confirmada pelo provedor e continua sem senha: %d", n)
	}
}

func TestAnAdministratorIsReachedByTheirVerifiedEmailAndKeepsTheRole(t *testing.T) {
	pool := setupPool(t)
	ctx := context.Background()
	id := newAdmin(t, pool, "dono@exemplo.com")
	got, err := postgres.NewUsers(pool).Upsert(ctx, socialIdentity("google", "g-dono", "dono@exemplo.com", true))
	if err != nil || got.ID != id || got.Role != domain.RoleAdmin {
		t.Fatalf("admin: %+v %v", got, err)
	}
}

func TestWithoutAVerifiedEmailNothingIsCreatedOrLinked(t *testing.T) {
	pool := setupPool(t)
	ctx := context.Background()
	if _, err := pool.Exec(ctx, `INSERT INTO app_user (display_name, email, email_verified_at, role) VALUES ('Ana', 'ana@exemplo.com', now(), 'user')`); err != nil {
		t.Fatal(err)
	}
	users := postgres.NewUsers(pool)
	for name, id := range map[string]domain.Identity{
		"e-mail nao confirmado": socialIdentity("github", "gh-1", "ana@exemplo.com", false),
		"sem e-mail":            socialIdentity("github", "gh-2", "", true),
		"so espacos":            socialIdentity("github", "gh-3", "   ", true),
	} {
		if _, err := users.Upsert(ctx, id); !errors.Is(err, app.ErrEmailNotVerified) {
			t.Errorf("%s: esperava ErrEmailNotVerified, veio %v", name, err)
		}
	}
	if n := count(t, pool, `SELECT count(*) FROM app_user`); n != 1 {
		t.Fatalf("nenhuma conta criada: %d", n)
	}
	if n := count(t, pool, `SELECT count(*) FROM auth_identity`); n != 0 {
		t.Fatalf("nenhum vinculo criado: %d", n)
	}
}

func TestAnEmailHeldByADeletedAccountFailsCleanlyWithoutLooping(t *testing.T) {
	pool := setupPool(t)
	ctx := context.Background()
	if _, err := pool.Exec(ctx, `INSERT INTO app_user (display_name, email, role, deleted_at) VALUES ('x', 'apagada@exemplo.com', 'user', now())`); err != nil {
		t.Fatal(err)
	}
	if _, err := postgres.NewUsers(pool).Upsert(ctx, socialIdentity("google", "g-1", "apagada@exemplo.com", true)); err == nil {
		t.Fatal("conta apagada segura o e-mail: o login social nao pode passar")
	}
}

func TestThePictureIsSavedAndNotErasedByALoginWithoutOne(t *testing.T) {
	pool := setupPool(t)
	ctx := context.Background()
	users := postgres.NewUsers(pool)
	id := socialIdentity("google", "g-7", "ana@exemplo.com", true)
	id.PictureURL = "https://lh3.googleusercontent.com/a/x"
	first, err := users.Upsert(ctx, id)
	if err != nil || first.PictureURL != id.PictureURL {
		t.Fatalf("a foto fica no perfil: %+v %v", first, err)
	}
	id.PictureURL = ""
	if again, err := users.Upsert(ctx, id); err != nil || again.PictureURL != "https://lh3.googleusercontent.com/a/x" {
		t.Fatalf("login sem foto nao apaga a anterior: %+v %v", again, err)
	}
	id.PictureURL = "http://inseguro.test/a.png"
	if again, err := users.Upsert(ctx, id); err != nil || again.PictureURL != "https://lh3.googleusercontent.com/a/x" {
		t.Fatalf("so https e aceito: %+v %v", again, err)
	}
	if got, err := users.Get(ctx, first.ID); err != nil || got.PictureURL != first.PictureURL {
		t.Fatalf("Get devolve a foto: %+v %v", got, err)
	}
}
