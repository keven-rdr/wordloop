package passkeys_test

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/keven-rdr/wordloop/api/internal/identity/domain"
	"github.com/keven-rdr/wordloop/api/internal/identity/identitytest"
	"github.com/keven-rdr/wordloop/api/internal/identity/passkeys"
)

func TestLoginOpensASessionWithMFAAndRecordsTheUse(t *testing.T) {
	e := newEnv(t, domain.RoleAdmin)
	a := identitytest.NewSoftAuthenticator(origin, rpID)
	e.registerWithInvite(t, a)
	before := e.sessions.Len()
	st, err := e.loginBegin().Run(bg)
	if err != nil {
		t.Fatal(err)
	}
	resp, _ := a.Get(st.Options, e.uid())
	opened, err := e.loginFinish().Run(bg, passkeys.FinishLoginInput{ChallengeID: st.ChallengeID, Credential: resp, UserAgent: "ua"})
	if err != nil || opened.SessionID == "" {
		t.Fatalf("login: %+v %v", opened, err)
	}
	if e.sessions.Len() != before+1 {
		t.Fatal("uma sessao nova deve existir")
	}
	if got := e.store.Credentials()[0]; got.SignCount != 1 || got.LastUsedAt.IsZero() {
		t.Fatalf("o uso deve ser registrado: %+v", got)
	}
}

func TestLoginRejectsEverythingThatIsNotAFreshValidAssertion(t *testing.T) {
	e := newEnv(t, domain.RoleUser)
	a := identitytest.NewSoftAuthenticator(origin, rpID)
	e.registerWithInvite(t, a)
	st, _ := e.loginBegin().Run(bg)
	resp, _ := a.Get(st.Options, e.uid())
	in := passkeys.FinishLoginInput{ChallengeID: st.ChallengeID, Credential: resp}
	if _, err := e.loginFinish().Run(bg, in); err != nil {
		t.Fatal(err)
	}
	if _, err := e.loginFinish().Run(bg, in); !errors.Is(err, passkeys.ErrUnauthorized) {
		t.Fatalf("repetir a mesma resposta: %v", err)
	}

	st2, _ := e.loginBegin().Run(bg)
	forged, _ := identitytest.NewSoftAuthenticator(origin, rpID).Get(st2.Options, e.uid())
	if _, err := e.loginFinish().Run(bg, passkeys.FinishLoginInput{ChallengeID: st2.ChallengeID, Credential: forged}); !errors.Is(err, passkeys.ErrUnauthorized) {
		t.Fatalf("passkey desconhecida: %v", err)
	}

	st3, _ := e.loginBegin().Run(bg)
	wrongOwner, _ := a.Get(st3.Options, []byte("outra-pessoa-123"))
	if _, err := e.loginFinish().Run(bg, passkeys.FinishLoginInput{ChallengeID: st3.ChallengeID, Credential: wrongOwner}); !errors.Is(err, passkeys.ErrUnauthorized) {
		t.Fatalf("userHandle de outra pessoa: %v", err)
	}
	if _, err := e.loginFinish().Run(bg, passkeys.FinishLoginInput{ChallengeID: "inventado", Credential: resp}); !errors.Is(err, passkeys.ErrUnauthorized) {
		t.Fatalf("desafio inventado: %v", err)
	}
}

func TestLoginRefusesARegressiveCounter(t *testing.T) {
	e := newEnv(t, domain.RoleUser)
	a := identitytest.NewSoftAuthenticator(origin, rpID)
	e.registerWithInvite(t, a)
	st, _ := e.loginBegin().Run(bg)
	resp, _ := a.Get(st.Options, e.uid())
	a.SetCounter(0)
	stored := e.store.Credentials()[0]
	_ = e.store.RecordLogin(bg, passkeys.Credential{CredentialID: stored.CredentialID, SignCount: 50}, now)
	if _, err := e.loginFinish().Run(bg, passkeys.FinishLoginInput{ChallengeID: st.ChallengeID, Credential: resp}); !errors.Is(err, passkeys.ErrUnauthorized) {
		t.Fatalf("contador regressivo (clone) deve recusar o login: %v", err)
	}
}

func TestListAndDeleteRules(t *testing.T) {
	e := newEnv(t, domain.RoleAdmin)
	e.registerWithInvite(t, identitytest.NewSoftAuthenticator(origin, rpID))
	e.registerSecond(t)
	views, err := passkeys.List{Store: e.store}.Run(bg, e.owner.ID)
	if err != nil || len(views) != 2 {
		t.Fatalf("lista: %v %v", views, err)
	}
	admin := domain.Principal{User: domain.User{ID: e.owner.ID, Role: domain.RoleAdmin}}
	del := passkeys.Delete{Store: e.store}
	if err := del.Run(bg, admin, uuid.New()); !errors.Is(err, passkeys.ErrNotFound) {
		t.Fatalf("id inexistente: %v", err)
	}
	if err := del.Run(bg, admin, views[0].ID); err != nil {
		t.Fatalf("com 2 passkeys pode apagar uma: %v", err)
	}
	if err := del.Run(bg, admin, views[1].ID); !errors.Is(err, passkeys.ErrLastMethod) {
		t.Fatalf("o admin nunca fica sem passkey: %v", err)
	}
	e.store.SetIdentities(e.owner.ID, 1)
	if err := del.Run(bg, admin, views[1].ID); !errors.Is(err, passkeys.ErrLastMethod) {
		t.Fatalf("nem com Google vinculado: %v", err)
	}
	user := domain.Principal{User: domain.User{ID: e.owner.ID, Role: domain.RoleUser}}
	if err := del.Run(bg, user, views[1].ID); err != nil {
		t.Fatalf("aluno com outra forma de entrar pode apagar a ultima passkey: %v", err)
	}
}

func TestDeleteCannotTouchSomeoneElsesPasskey(t *testing.T) {
	e := newEnv(t, domain.RoleUser)
	e.registerWithInvite(t, identitytest.NewSoftAuthenticator(origin, rpID))
	e.registerSecond(t)
	mine := e.store.Credentials()[0]
	stranger := domain.Principal{User: domain.User{ID: uuid.New(), Role: domain.RoleUser}}
	if err := (passkeys.Delete{Store: e.store}).Run(bg, stranger, mine.ID); err == nil {
		t.Fatal("outra pessoa nao pode apagar a minha passkey")
	}
	if len(e.store.Credentials()) != 2 {
		t.Fatal("nada deve ser apagado")
	}
}
