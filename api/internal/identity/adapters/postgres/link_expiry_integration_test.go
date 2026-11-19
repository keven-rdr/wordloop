//go:build integration

package postgres_test

import (
	"context"
	"errors"
	"testing"

	"github.com/keven-rdr/wordloop/api/internal/identity/adapters/postgres"
	"github.com/keven-rdr/wordloop/api/internal/identity/passwords"
	"github.com/keven-rdr/wordloop/api/internal/identity/registration"
)

// Criterios do roadmap: I-03 (senha -> Google com o mesmo e-mail entra na mesma conta) e I-04 (o link expira).

func TestTheConfirmationAndRecoveryLinksLastTheirTimeAndThenStopWorking(t *testing.T) {
	f := newFlow(t)
	ctx := context.Background()
	_ = f.register.Run(ctx, registration.Input{Email: "ana@exemplo.com"})
	var secs int
	if err := f.pool.QueryRow(ctx, `SELECT extract(epoch FROM expires_at - created_at)::int FROM auth_token WHERE purpose = 'verify_email'`).Scan(&secs); err != nil || secs != int(registration.VerifyTTL.Seconds()) {
		t.Fatalf("o link de confirmacao vale %v: %d s %v", registration.VerifyTTL, secs, err)
	}
	secret := f.mailbox.LastSecret()
	if _, err := f.pool.Exec(ctx, `UPDATE auth_token SET expires_at = now() - interval '1 second'`); err != nil {
		t.Fatal(err)
	}
	if _, err := f.set.Run(ctx, passwords.SetInput{Token: secret, Password: "senha bem comprida da ana"}); !errors.Is(err, passwords.ErrUnauthorized) {
		t.Fatalf("link de confirmacao vencido: %v", err)
	}
	if f.verified(t, "ana@exemplo.com") {
		t.Fatal("link vencido nao confirma o e-mail")
	}

	// recuperacao: reenvio (envelhece o anterior para passar do limite de 1 por minuto), conta confirmada, depois o link vence
	_, _ = f.pool.Exec(ctx, `UPDATE auth_token SET created_at = created_at - interval '2 minutes'`)
	_ = f.register.Run(ctx, registration.Input{Email: "ana@exemplo.com"})
	if _, err := f.set.Run(ctx, passwords.SetInput{Token: f.mailbox.LastSecret(), Password: "senha bem comprida da ana"}); err != nil {
		t.Fatal(err)
	}
	_, _ = f.pool.Exec(ctx, `UPDATE auth_token SET created_at = created_at - interval '2 minutes'`)
	if err := f.forgot.Run(ctx, registration.Input{Email: "ana@exemplo.com"}); err != nil {
		t.Fatal(err)
	}
	if err := f.pool.QueryRow(ctx, `SELECT extract(epoch FROM expires_at - created_at)::int FROM auth_token WHERE purpose = 'recovery' AND used_at IS NULL`).Scan(&secs); err != nil || secs != int(registration.RecoveryTTL.Seconds()) {
		t.Fatalf("o link de recuperacao vale %v: %d s %v", registration.RecoveryTTL, secs, err)
	}
	recovery := f.mailbox.LastSecret()
	if _, err := f.pool.Exec(ctx, `UPDATE auth_token SET expires_at = now() - interval '1 second' WHERE purpose = 'recovery'`); err != nil {
		t.Fatal(err)
	}
	if _, err := f.set.Run(ctx, passwords.SetInput{Token: recovery, Password: "senha nova que nao deve valer"}); !errors.Is(err, passwords.ErrUnauthorized) {
		t.Fatalf("link de recuperacao vencido: %v", err)
	}
	if _, err := f.login.Run(ctx, passwords.LoginInput{Email: "ana@exemplo.com", Password: "senha bem comprida da ana"}); err != nil {
		t.Fatalf("a senha antiga continua valendo: %v", err)
	}
}

func TestAPasswordAccountThatEntersWithGoogleIsTheSameAccountAndKeepsItsPassword(t *testing.T) {
	f := newFlow(t)
	ctx := context.Background()
	_ = f.register.Run(ctx, registration.Input{Email: "ana@exemplo.com"})
	if _, err := f.set.Run(ctx, passwords.SetInput{Token: f.mailbox.LastSecret(), Password: "senha bem comprida da ana"}); err != nil {
		t.Fatal(err)
	}
	var existing string
	if err := f.pool.QueryRow(ctx, `SELECT id::text FROM app_user WHERE email = 'ana@exemplo.com'`).Scan(&existing); err != nil {
		t.Fatal(err)
	}
	got, err := postgres.NewUsers(f.pool).Upsert(ctx, socialIdentity("google", "g-ana", "ANA@Exemplo.com", true))
	if err != nil || got.ID.String() != existing {
		t.Fatalf("o Google deve entrar na conta da senha: %+v %v", got, err)
	}
	if n := count(t, f.pool, `SELECT count(*) FROM app_user`); n != 1 {
		t.Fatalf("nenhuma conta nova: %d", n)
	}
	if _, err := f.login.Run(ctx, passwords.LoginInput{Email: "ana@exemplo.com", Password: "senha bem comprida da ana"}); err != nil {
		t.Fatalf("a senha continua valendo depois do vinculo: %v", err)
	}
}

func TestAfterTheAccountIsErasedTheSameEmailStartsAFreshAccountWithNothingFromBefore(t *testing.T) {
	pool := setupPool(t)
	ctx := context.Background()
	users := postgres.NewUsers(pool)
	first, err := users.Upsert(ctx, socialIdentity("google", "g-ana", "ana@exemplo.com", true))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := pool.Exec(ctx, `DELETE FROM app_user WHERE id = $1`, first.ID); err != nil { // o que a exclusao da conta faz no fim
		t.Fatal(err)
	}
	second, err := users.Upsert(ctx, socialIdentity("google", "g-ana", "ana@exemplo.com", true))
	if err != nil || second.ID == first.ID {
		t.Fatalf("depois de apagada, a mesma pessoa comeca uma conta nova: %+v %v", second, err)
	}
	if n := count(t, pool, `SELECT count(*) FROM app_user`); n != 1 {
		t.Fatalf("so a conta nova existe: %d", n)
	}
	if n := count(t, pool, `SELECT count(*) FROM auth_identity WHERE user_id = $1`, first.ID); n != 0 {
		t.Fatalf("o vinculo antigo foi embora: %d", n)
	}
}
