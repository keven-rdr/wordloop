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
	"github.com/keven-rdr/wordloop/api/internal/identity/identitytest"
	"github.com/keven-rdr/wordloop/api/internal/identity/passwords"
	"github.com/keven-rdr/wordloop/api/internal/identity/registration"
	"github.com/keven-rdr/wordloop/api/internal/platform/password"
	"github.com/keven-rdr/wordloop/api/internal/platform/token"
)

type flow struct {
	pool     *pgxpool.Pool
	mailbox  *identitytest.MailBox
	register registration.Register
	forgot   registration.Forgot
	set      passwords.SetPassword
	login    passwords.Login
}

func newFlow(t *testing.T) flow {
	t.Helper()
	pool := setupPool(t)
	box := &identitytest.MailBox{}
	issuer := registration.Issuer{
		Store: postgres.NewRegistration(pool), Mail: box, PublicURL: "https://tst.wordloop.com.br",
	}
	hasher := password.NewHasher(2)
	dummy, _ := hasher.Hash(context.Background(), "isca")
	store, users := postgres.NewPasswords(pool), postgres.NewUsers(pool)
	opener := app.SessionOpener{Sessions: postgres.NewSessions(pool)}
	return flow{
		pool: pool, mailbox: box,
		register: registration.Register{Issuer: issuer}, forgot: registration.Forgot{Issuer: issuer},
		set:   passwords.SetPassword{Store: store, Hasher: hasher, Users: users, Opener: opener},
		login: passwords.Login{Store: store, Hasher: hasher, Users: users, Opener: opener, DummyHash: dummy},
	}
}

func (f flow) verified(t *testing.T, email string) bool {
	t.Helper()
	acc, err := postgres.NewRegistration(f.pool).FindByEmail(context.Background(), email)
	if err != nil {
		t.Fatal(err)
	}
	return acc.Verified
}

func TestSignUpConfirmEmailSetPasswordAndSignInOnRealPostgres(t *testing.T) {
	f := newFlow(t)
	ctx := context.Background()
	if err := f.register.Run(ctx, registration.Input{Email: "Ana@Exemplo.com", Locale: "pt-BR"}); err != nil {
		t.Fatal(err)
	}
	if f.verified(t, "ana@exemplo.com") {
		t.Fatal("antes de confirmar o e-mail nao esta confirmado")
	}
	secret := f.mailbox.LastSecret()
	if _, err := f.login.Run(ctx, passwords.LoginInput{Email: "ana@exemplo.com", Password: "qualquer senha comprida"}); !errors.Is(err, passwords.ErrInvalidCredentials) {
		t.Fatalf("conta pendente nao entra: %v", err)
	}
	opened, err := f.set.Run(ctx, passwords.SetInput{Token: secret, Password: "senha bem comprida da ana"})
	if err != nil || opened == nil {
		t.Fatalf("definir a senha: %+v %v", opened, err)
	}
	if !f.verified(t, "ana@exemplo.com") {
		t.Fatal("o link confirma o e-mail")
	}
	if _, err := f.set.Run(ctx, passwords.SetInput{Token: secret, Password: "outra senha bem comprida"}); !errors.Is(err, passwords.ErrUnauthorized) {
		t.Fatalf("o link vale uma vez so: %v", err)
	}
	if _, err := f.login.Run(ctx, passwords.LoginInput{Email: "ana@exemplo.com", Password: "senha bem comprida da ana"}); err != nil {
		t.Fatalf("entrar com a senha nova: %v", err)
	}
}

func TestResendReplacesTheOpenTokenAndTheOldLinkStopsWorking(t *testing.T) {
	f := newFlow(t)
	ctx := context.Background()
	_ = f.register.Run(ctx, registration.Input{Email: "ana@exemplo.com"})
	first := f.mailbox.LastSecret()
	_, err := f.pool.Exec(ctx, `UPDATE auth_token SET created_at = created_at - interval '2 minutes'`)
	if err != nil {
		t.Fatal(err)
	}
	_ = f.register.Run(ctx, registration.Input{Email: "ana@exemplo.com"})
	second := f.mailbox.LastSecret()
	var open int
	if err := f.pool.QueryRow(ctx, `SELECT count(*) FROM auth_token WHERE purpose = 'verify_email' AND used_at IS NULL AND expires_at > $1`, time.Now()).Scan(&open); err != nil || open != 1 {
		t.Fatalf("so um token aberto: %d %v", open, err)
	}
	if _, err := f.set.Run(ctx, passwords.SetInput{Token: first, Password: "senha bem comprida da ana"}); !errors.Is(err, passwords.ErrUnauthorized) {
		t.Fatalf("o link antigo morre: %v", err)
	}
	if _, err := f.set.Run(ctx, passwords.SetInput{Token: second, Password: "senha bem comprida da ana"}); err != nil {
		t.Fatalf("o link novo vale: %v", err)
	}
}

func TestRecoveryChangesThePasswordEndsSessionsAndIsRefusedForAdmins(t *testing.T) {
	f := newFlow(t)
	ctx := context.Background()
	_ = f.register.Run(ctx, registration.Input{Email: "ana@exemplo.com"})
	if _, err := f.set.Run(ctx, passwords.SetInput{Token: f.mailbox.LastSecret(), Password: "senha bem comprida da ana"}); err != nil {
		t.Fatal(err)
	}
	var sessions int
	_ = f.pool.QueryRow(ctx, `SELECT count(*) FROM auth_session`).Scan(&sessions)
	if sessions != 1 {
		t.Fatalf("a confirmacao abriu uma sessao: %d", sessions)
	}
	_, _ = f.pool.Exec(ctx, `UPDATE auth_token SET created_at = created_at - interval '2 minutes'`)
	if err := f.forgot.Run(ctx, registration.Input{Email: "ana@exemplo.com", Locale: "en-US"}); err != nil {
		t.Fatal(err)
	}
	if _, err := f.set.Run(ctx, passwords.SetInput{Token: f.mailbox.LastSecret(), Password: "nova senha bem comprida"}); err != nil {
		t.Fatal(err)
	}
	if err := f.pool.QueryRow(ctx, `SELECT count(*) FROM auth_session`).Scan(&sessions); err != nil || sessions != 1 {
		t.Fatalf("a troca encerra as sessoes antigas e abre so a nova: %d %v", sessions, err)
	}
	if _, err := f.login.Run(ctx, passwords.LoginInput{Email: "ana@exemplo.com", Password: "senha bem comprida da ana"}); !errors.Is(err, passwords.ErrInvalidCredentials) {
		t.Fatalf("a senha antiga deixa de valer: %v", err)
	}

	admin := newAdmin(t, f.pool, "dono@exemplo.com")
	before := f.mailbox.Count()
	_ = f.forgot.Run(ctx, registration.Input{Email: "dono@exemplo.com"})
	_ = f.register.Run(ctx, registration.Input{Email: "dono@exemplo.com"})
	if f.mailbox.Count() != before {
		t.Fatal("admin nunca recebe e-mail de recuperacao nem de cadastro")
	}
	hash := putRecovery(t, f.pool, admin)
	if _, err := f.set.Run(ctx, passwords.SetInput{Token: hash, Password: "senha bem comprida do dono"}); !errors.Is(err, passwords.ErrUnauthorized) {
		t.Fatalf("token de recuperacao de admin nao vale: %v", err)
	}
}

func putRecovery(t *testing.T, pool *pgxpool.Pool, userID uuid.UUID) string {
	t.Helper()
	secret := "recuperacao-forjada"
	err := postgres.NewRegistration(pool).ReplaceToken(context.Background(), userID, domain.TokenRecovery, token.Hash(secret), time.Now(), time.Now().Add(time.Minute))
	if err != nil {
		t.Fatal(err)
	}
	return secret
}

func TestTheHourlyLimitCountsExpiredTokensToo(t *testing.T) {
	f := newFlow(t)
	ctx := context.Background()
	for range 7 {
		_ = f.register.Run(ctx, registration.Input{Email: "spam@exemplo.com"})
		_, err := f.pool.Exec(ctx, `UPDATE auth_token SET created_at = created_at - interval '2 minutes'`)
		if err != nil {
			t.Fatal(err)
		}
	}
	if f.mailbox.Count() != 5 {
		t.Fatalf("no maximo 5 e-mails por hora para o mesmo endereco, vieram %d", f.mailbox.Count())
	}
}

func TestAConfirmationLinkStopsWorkingOnceTheEmailIsConfirmedByAnotherPath(t *testing.T) {
	f := newFlow(t)
	ctx := context.Background()
	_ = f.register.Run(ctx, registration.Input{Email: "ana@exemplo.com"})
	secret := f.mailbox.LastSecret()
	if _, err := f.pool.Exec(ctx, `UPDATE app_user SET email_verified_at = now() WHERE lower(email) = 'ana@exemplo.com'`); err != nil {
		t.Fatal(err)
	}
	if _, err := f.set.Run(ctx, passwords.SetInput{Token: secret, Password: "senha bem comprida da ana"}); !errors.Is(err, passwords.ErrUnauthorized) {
		t.Fatalf("o link de confirmacao so vale enquanto o e-mail nao esta confirmado: %v", err)
	}
}

func TestEmailsOfDeletedAccountsAreIgnoredWithoutLooping(t *testing.T) {
	f := newFlow(t)
	ctx := context.Background()
	if _, err := f.pool.Exec(ctx, `INSERT INTO app_user (display_name, email, role, deleted_at) VALUES ('x', 'apagada@exemplo.com', 'user', now())`); err != nil {
		t.Fatal(err)
	}
	if err := f.register.Run(ctx, registration.Input{Email: "apagada@exemplo.com"}); err != nil || f.mailbox.Count() != 0 {
		t.Fatalf("conta apagada: %v %d", err, f.mailbox.Count())
	}
}
