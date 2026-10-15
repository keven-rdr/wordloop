package passwords_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/keven-rdr/wordloop/api/internal/identity/app"
	"github.com/keven-rdr/wordloop/api/internal/identity/domain"
	"github.com/keven-rdr/wordloop/api/internal/identity/passwords"
)

// Falhas de infraestrutura (banco, hasher, sessao) nao podem virar "senha errada": voltam como erro proprio e,
// quando o caso pede, sem derrubar o que ja foi gravado. Cada decorador abaixo faz UMA operacao falhar.

var errBoom = errors.New("falha injetada")

type faultyStore struct {
	passwords.Store
	find, begin, success, peek, set error
}

func (s faultyStore) FindByEmail(ctx context.Context, email string) (passwords.Account, error) {
	if s.find != nil {
		return passwords.Account{}, s.find
	}
	return s.Store.FindByEmail(ctx, email)
}

func (s faultyStore) BeginAttempt(ctx context.Context, id uuid.UUID, at time.Time) (bool, error) {
	if s.begin != nil {
		return false, s.begin
	}
	return s.Store.BeginAttempt(ctx, id, at)
}

func (s faultyStore) RegisterSuccess(ctx context.Context, id uuid.UUID) error {
	if s.success != nil {
		return s.success
	}
	return s.Store.RegisterSuccess(ctx, id)
}

func (s faultyStore) PeekInvite(ctx context.Context, h []byte, now time.Time) (passwords.Owner, error) {
	if s.peek != nil {
		return passwords.Owner{}, s.peek
	}
	return s.Store.PeekInvite(ctx, h, now)
}

func (s faultyStore) SetWithInvite(ctx context.Context, h []byte, id uuid.UUID, hash string, now time.Time) error {
	if s.set != nil {
		return s.set
	}
	return s.Store.SetWithInvite(ctx, h, id, hash, now)
}

type faultyHasher struct {
	passwords.Hasher
	hash, verify error
}

func (h faultyHasher) Hash(ctx context.Context, plain string) (string, error) {
	if h.hash != nil {
		return "", h.hash
	}
	return h.Hasher.Hash(ctx, plain)
}

func (h faultyHasher) Verify(ctx context.Context, plain, encoded string) (bool, error) {
	if h.verify != nil {
		return false, h.verify
	}
	return h.Hasher.Verify(ctx, plain, encoded)
}

type faultyUsers struct {
	app.Users
	get error
}

func (u faultyUsers) Get(ctx context.Context, id uuid.UUID) (domain.User, error) {
	if u.get != nil {
		return domain.User{}, u.get
	}
	return u.Users.Get(ctx, id)
}

type faultySessions struct {
	app.Sessions
	create error
}

func (s faultySessions) Create(ctx context.Context, n app.NewSession) error {
	if s.create != nil {
		return s.create
	}
	return s.Sessions.Create(ctx, n)
}

func TestLoginReportsInfrastructureFailuresAsErrorsNotAsWrongPassword(t *testing.T) {
	e := newEnv(t, domain.RoleUser, false).withPassword()
	cases := map[string]func(l *passwords.Login){
		"busca da conta":    func(l *passwords.Login) { l.Store = faultyStore{Store: e.store, find: errBoom} },
		"contagem":          func(l *passwords.Login) { l.Store = faultyStore{Store: e.store, begin: errBoom} },
		"conferir a senha":  func(l *passwords.Login) { l.Hasher = faultyHasher{Hasher: e.hasher, verify: errBoom} },
		"zerar tentativas":  func(l *passwords.Login) { l.Store = faultyStore{Store: e.store, success: errBoom} },
		"leitura do perfil": func(l *passwords.Login) { l.Users = faultyUsers{Users: e.users, get: errBoom} },
		"abrir a sessao":    func(l *passwords.Login) { l.Opener.Sessions = faultySessions{Sessions: e.sessions, create: errBoom} },
	}
	for name, inject := range cases {
		l := e.login
		inject(&l)
		_, err := l.Run(bg, passwords.LoginInput{Email: email, Password: secret})
		if !errors.Is(err, errBoom) || errors.Is(err, passwords.ErrInvalidCredentials) {
			t.Errorf("%s: esperava o erro da infraestrutura, veio %v", name, err)
		}
	}
	if e.sessions.Len() != 0 {
		t.Fatal("nenhuma sessao pode nascer quando algo falha")
	}
}

func TestLoginWithAProfileThatNoLongerExistsLooksLikeWrongCredentials(t *testing.T) {
	e := newEnv(t, domain.RoleUser, false).withPassword()
	l := e.login
	l.Users = faultyUsers{Users: e.users, get: app.ErrNotFound}
	if _, err := l.Run(bg, passwords.LoginInput{Email: email, Password: secret}); !errors.Is(err, passwords.ErrInvalidCredentials) {
		t.Fatalf("perfil removido: esperava ErrInvalidCredentials, veio %v", err)
	}
}

func TestSetPasswordReportsInfrastructureFailures(t *testing.T) {
	e := newEnv(t, domain.RoleAdmin, false)
	cases := map[string]func(s *passwords.SetPassword){
		"ler o convite":  func(s *passwords.SetPassword) { s.Store = faultyStore{Store: e.store, peek: errBoom} },
		"gerar o hash":   func(s *passwords.SetPassword) { s.Hasher = faultyHasher{Hasher: e.hasher, hash: errBoom} },
		"gravar a senha": func(s *passwords.SetPassword) { s.Store = faultyStore{Store: e.store, set: errBoom} },
	}
	for name, inject := range cases {
		hash := invite(e, now.Add(10*time.Minute))
		s := e.set
		inject(&s)
		opened, err := s.Run(bg, passwords.SetInput{Invite: "segredo-do-convite", Password: secret})
		if !errors.Is(err, errBoom) || opened != nil || errors.Is(err, passwords.ErrUnauthorized) {
			t.Errorf("%s: esperava o erro da infraestrutura, veio %v (%+v)", name, err, opened)
		}
		if e.store.InviteUsed(hash) {
			t.Errorf("%s: o convite nao pode ser gasto se a senha nao foi gravada", name)
		}
	}
}

// Depois que a senha foi gravada, falhar ao abrir a sessao nao e erro: a pessoa so precisa entrar de novo.
func TestPasswordIsKeptWhenTheSessionCannotBeOpenedAfterwards(t *testing.T) {
	e := newEnv(t, domain.RoleAdmin, false)
	for name, inject := range map[string]func(s *passwords.SetPassword){
		"perfil nao lido": func(s *passwords.SetPassword) { s.Users = faultyUsers{Users: e.users, get: errBoom} },
		"sessao nao abriu": func(s *passwords.SetPassword) {
			s.Opener.Sessions = faultySessions{Sessions: e.sessions, create: errBoom}
		},
	} {
		invite(e, now.Add(10*time.Minute))
		s := e.set
		inject(&s)
		opened, err := s.Run(bg, passwords.SetInput{Invite: "segredo-do-convite", Password: secret})
		if err != nil || opened != nil {
			t.Errorf("%s: esperava sucesso sem sessao, veio %+v %v", name, opened, err)
		}
		if stored, _ := e.store.Row(e.id); stored != "h:"+secret {
			t.Errorf("%s: a senha deveria ter sido gravada", name)
		}
		e.store.Add(e.id, email, "") // limpa a senha para a proxima rodada
	}
}
