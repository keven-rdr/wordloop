package passwords_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/keven-rdr/wordloop/api/internal/identity/app"
	"github.com/keven-rdr/wordloop/api/internal/identity/domain"
	"github.com/keven-rdr/wordloop/api/internal/identity/identitytest"
	"github.com/keven-rdr/wordloop/api/internal/identity/passwords"
	"github.com/keven-rdr/wordloop/api/internal/platform/password"
	"github.com/keven-rdr/wordloop/api/internal/platform/token"
)

var (
	bg  = context.Background()
	now = time.Date(2026, 10, 16, 12, 0, 0, 0, time.UTC)
)

const (
	email    = "dono@exemplo.com"
	secret   = "uma senha bem comprida"
	dummyPwd = "h:isca-que-nunca-confere"
)

type env struct {
	store    *identitytest.PasswordStore
	users    *identitytest.Users
	sessions *identitytest.Sessions
	hasher   *identitytest.PlainHasher
	login    passwords.Login
	set      passwords.SetPassword
	id       uuid.UUID
}

func newEnv(t *testing.T, role domain.Role, mfaRequired bool) env {
	t.Helper()
	e := env{store: &identitytest.PasswordStore{}, users: &identitytest.Users{}, sessions: &identitytest.Sessions{}, hasher: &identitytest.PlainHasher{}}
	user, err := e.users.Upsert(bg, domain.Identity{Issuer: "i", Subject: "s", DisplayName: "Dono"})
	if err != nil {
		t.Fatal(err)
	}
	e.id = user.ID
	e.users.SetRole(e.id, role)
	_ = e.users.SetMFARequired(bg, e.id, mfaRequired)
	clock := identitytest.Clock(now)
	opener := app.SessionOpener{Sessions: e.sessions, Now: clock}
	e.login = passwords.Login{Store: e.store, Hasher: e.hasher, Users: e.users, Opener: opener, DummyHash: dummyPwd, Now: clock}
	e.set = passwords.SetPassword{Store: e.store, Hasher: e.hasher, Users: e.users, Opener: opener, Now: clock}
	return e
}

func (e env) withPassword() env {
	e.store.Add(e.id, email, "h:"+secret)
	return e
}

func (e env) signIn(emailIn, pwd string) error {
	_, err := e.login.Run(bg, passwords.LoginInput{Email: emailIn, Password: pwd, UserAgent: "ua"})
	return err
}

func TestRightPasswordOpensASessionWithoutProofOfTwoFactors(t *testing.T) {
	e := newEnv(t, domain.RoleAdmin, false).withPassword()
	opened, err := e.login.Run(bg, passwords.LoginInput{Email: "  Dono@Exemplo.com ", Password: secret, UserAgent: "ua"})
	if err != nil || opened.SessionID == "" {
		t.Fatalf("login: %+v %v", opened, err)
	}
	stored, _ := e.sessions.Last()
	if stored.MFA || stored.UserID != e.id || stored.UserAgent != "ua" {
		t.Fatalf("a sessao por senha nao comprova 2 fatores: %+v", stored)
	}
	if want := now.Add(domain.DefaultSessionPolicy().Admin.Absolute); !opened.ExpiresAt.Equal(want) {
		t.Fatalf("o admin tem sessao curta: %v, esperava %v", opened.ExpiresAt, want)
	}
}

func TestEveryRefusalLooksTheSameAndStillSpendsAHashCheck(t *testing.T) {
	e := newEnv(t, domain.RoleUser, false).withPassword()
	for name, in := range map[string][2]string{
		"senha errada": {email, "outra senha comprida"}, "e-mail desconhecido": {"ninguem@exemplo.com", secret}, "sem senha": {"", ""},
	} {
		if err := e.signIn(in[0], in[1]); !errors.Is(err, passwords.ErrInvalidCredentials) {
			t.Errorf("%s: esperava ErrInvalidCredentials, veio %v", name, err)
		}
	}
	decoys := 0
	for _, c := range e.hasher.Calls {
		if c == dummyPwd {
			decoys++
		}
	}
	if len(e.hasher.Calls) != 3 || decoys != 2 {
		t.Fatalf("sem conta tambem se confere um hash (o tempo nao revela nada): %v", e.hasher.Calls)
	}
	if e.sessions.Len() != 0 {
		t.Fatal("nenhuma sessao pode nascer")
	}
}

func TestAccountWithoutAPasswordCannotSignInByPassword(t *testing.T) {
	e := newEnv(t, domain.RoleAdmin, false)
	e.store.Add(e.id, email, "")
	if err := e.signIn(email, ""); !errors.Is(err, passwords.ErrInvalidCredentials) {
		t.Fatalf("conta sem senha: %v", err)
	}
}

func TestFiveWrongPasswordsLockTheAccountEvenForTheRightOne(t *testing.T) {
	e := newEnv(t, domain.RoleUser, false).withPassword()
	for range 5 {
		_ = e.signIn(email, "errada errada errada")
	}
	if _, attempts := e.store.Row(e.id); attempts != 5 {
		t.Fatalf("esperava 5 tentativas contadas, veio %d", attempts)
	}
	if err := e.signIn(email, secret); !errors.Is(err, passwords.ErrInvalidCredentials) {
		t.Fatalf("a 6a tentativa encontra a conta bloqueada, ate com a senha certa: %v", err)
	}
	if err := e.signIn(email, secret); !errors.Is(err, passwords.ErrInvalidCredentials) {
		t.Fatalf("conta bloqueada recusa ate a senha certa: %v", err)
	}
	if _, attempts := e.store.Row(e.id); attempts != 5 {
		t.Fatalf("tentativa em conta bloqueada nao conta de novo: %d", attempts)
	}
	if e.sessions.Len() != 0 {
		t.Fatal("nenhuma sessao enquanto bloqueada")
	}
	later := e.login
	later.Now = identitytest.Clock(now.Add(2 * time.Minute))
	if _, err := later.Run(bg, passwords.LoginInput{Email: email, Password: secret}); err != nil {
		t.Fatalf("passado o bloqueio a senha certa entra: %v", err)
	}
	if _, attempts := e.store.Row(e.id); attempts != 0 {
		t.Fatalf("o acerto zera as tentativas: %d", attempts)
	}
}

func TestAccountThatTurnedTwoFactorOnCannotSignInWithAPasswordAlone(t *testing.T) {
	e := newEnv(t, domain.RoleAdmin, true).withPassword()
	if err := e.signIn(email, secret); !errors.Is(err, app.ErrMFARequired) {
		t.Fatalf("esperava ErrMFARequired, veio %v", err)
	}
	if e.sessions.Len() != 0 {
		t.Fatal("nenhuma sessao pode nascer")
	}
}

func TestAbsurdlyLongPasswordsAreRefusedWithoutHashing(t *testing.T) {
	e := newEnv(t, domain.RoleUser, false).withPassword()
	long := make([]byte, 5000)
	for i := range long {
		long[i] = 'a'
	}
	if err := e.signIn(email, string(long)); !errors.Is(err, passwords.ErrInvalidCredentials) || len(e.hasher.Calls) != 0 {
		t.Fatalf("esperava recusa sem gastar hash: %v %v", err, e.hasher.Calls)
	}
}

func TestPolicy(t *testing.T) {
	cases := map[string]struct {
		pwd, email string
		weak       bool
	}{
		"curta":         {"curta", "a@b.co", true},
		"onze":          {"onze letras", "a@b.co", true},
		"doze":          {"doze letras!", "a@b.co", false},
		"igual ao mail": {"dono@exemplo.com", "dono@exemplo.com", true},
		"igual ao nome": {"donodonodono", "donodonodono@exemplo.com", true},
		"repetida":      {"aaaaaaaaaaaaaaaa", "a@b.co", true},
		"muito longa":   {string(make([]rune, 129)), "a@b.co", true},
		"com acentos":   {"ação de graças", "a@b.co", false},
	}
	for name, c := range cases {
		if err := passwords.CheckPolicy(c.pwd, c.email); (err != nil) != c.weak {
			t.Errorf("%s: fraca=%v, erro=%v", name, c.weak, err)
		}
	}
}

func invite(e env, until time.Time) []byte {
	hash := token.Hash("segredo-do-convite")
	e.store.AddInvite(hash, e.id, email, until)
	return hash
}

func TestInviteSetsThePasswordConsumesItAndOpensTheSession(t *testing.T) {
	e := newEnv(t, domain.RoleAdmin, false)
	hash := invite(e, now.Add(10*time.Minute))
	opened, err := e.set.Run(bg, passwords.SetInput{Token: "segredo-do-convite", Password: secret, UserAgent: "ua"})
	if err != nil || opened == nil {
		t.Fatalf("set: %+v %v", opened, err)
	}
	if stored, _ := e.store.Row(e.id); stored != "h:"+secret || !e.store.InviteUsed(hash) {
		t.Fatalf("senha gravada e convite consumido: %q %v", stored, e.store.InviteUsed(hash))
	}
	if _, err := e.set.Run(bg, passwords.SetInput{Token: "segredo-do-convite", Password: secret}); !errors.Is(err, passwords.ErrUnauthorized) {
		t.Fatalf("o convite vale uma vez so: %v", err)
	}
}

func TestInviteThatIsUnknownOrExpiredChangesNothing(t *testing.T) {
	e := newEnv(t, domain.RoleAdmin, false)
	invite(e, now.Add(-time.Minute))
	for _, inv := range []string{"segredo-do-convite", "inventado", ""} {
		if _, err := e.set.Run(bg, passwords.SetInput{Token: inv, Password: secret}); !errors.Is(err, passwords.ErrUnauthorized) {
			t.Errorf("convite %q: esperava ErrUnauthorized, veio %v", inv, err)
		}
	}
	if stored, _ := e.store.Row(e.id); stored != "" || e.sessions.Len() != 0 {
		t.Fatal("nada pode mudar com convite invalido")
	}
}

func TestWeakPasswordKeepsTheInviteForAnotherTry(t *testing.T) {
	e := newEnv(t, domain.RoleAdmin, false)
	hash := invite(e, now.Add(10*time.Minute))
	if _, err := e.set.Run(bg, passwords.SetInput{Token: "segredo-do-convite", Password: "curta"}); !errors.Is(err, passwords.ErrWeak) {
		t.Fatalf("esperava ErrWeak, veio %v", err)
	}
	if e.store.InviteUsed(hash) {
		t.Fatal("senha fraca nao pode gastar o convite")
	}
}

func TestInviteForATwoFactorAccountSetsThePasswordButOpensNoSession(t *testing.T) {
	e := newEnv(t, domain.RoleAdmin, true)
	invite(e, now.Add(10*time.Minute))
	opened, err := e.set.Run(bg, passwords.SetInput{Token: "segredo-do-convite", Password: secret})
	if err != nil || opened != nil || e.sessions.Len() != 0 {
		t.Fatalf("esperava senha gravada e nenhuma sessao: %+v %v", opened, err)
	}
	if stored, _ := e.store.Row(e.id); stored != "h:"+secret {
		t.Fatal("a senha deveria ter sido gravada")
	}
}

func TestSetThenLoginWithTheRealArgon2idHasher(t *testing.T) {
	e := newEnv(t, domain.RoleAdmin, false)
	argon := password.NewHasher(1)
	e.set.Hasher, e.login.Hasher = argon, argon
	e.login.DummyHash, _ = argon.Hash(bg, "isca")
	invite(e, now.Add(10*time.Minute))
	if _, err := e.set.Run(bg, passwords.SetInput{Token: "segredo-do-convite", Password: secret}); err != nil {
		t.Fatal(err)
	}
	stored, _ := e.store.Row(e.id)
	if !strings.HasPrefix(stored, "$argon2id$") || strings.Contains(stored, secret) {
		t.Fatalf("o banco guarda so o hash Argon2id: %q", stored)
	}
	if err := e.signIn(email, secret); err != nil {
		t.Fatalf("senha certa: %v", err)
	}
	if err := e.signIn(email, secret+"x"); !errors.Is(err, passwords.ErrInvalidCredentials) {
		t.Fatalf("senha errada: %v", err)
	}
}
