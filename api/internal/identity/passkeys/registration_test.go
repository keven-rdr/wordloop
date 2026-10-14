package passkeys_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/keven-rdr/wordloop/api/internal/identity/app"
	"github.com/keven-rdr/wordloop/api/internal/identity/domain"
	"github.com/keven-rdr/wordloop/api/internal/identity/identitytest"
	"github.com/keven-rdr/wordloop/api/internal/identity/passkeys"
)

func TestRegisterWithInviteStoresPasskeyConsumesInviteAndOpensSession(t *testing.T) {
	e := newEnv(t, domain.RoleAdmin)
	got := e.registerWithInvite(t, identitytest.NewSoftAuthenticator(origin, rpID))
	if got.Session == nil || got.Credential.Name != "Meu celular" || !got.Credential.UserVerified {
		t.Fatalf("resultado inesperado: %+v", got)
	}
	if len(e.store.Credentials()) != 1 || !e.store.InviteUsed(invite) {
		t.Fatal("a passkey deve ser gravada e o convite consumido")
	}
	stored, ok := e.sessions.Last()
	if !ok || !stored.MFA || stored.UserID != e.owner.ID {
		t.Fatalf("a sessao do admin deve nascer com MFA verdadeiro (passkey com UV): %+v", stored)
	}
	if _, err := e.begin().Run(bg, passkeys.Subject{InviteToken: invite}); !errors.Is(err, passkeys.ErrUnauthorized) {
		t.Fatalf("convite usado nao pode iniciar outro registro: %v", err)
	}
}

func TestBeginRegistrationNeedsAnInviteOrASession(t *testing.T) {
	e := newEnv(t, domain.RoleUser)
	for name, s := range map[string]passkeys.Subject{
		"sem nada":          {},
		"convite inventado": {InviteToken: "inventado"},
	} {
		if _, err := e.begin().Run(bg, s); !errors.Is(err, passkeys.ErrUnauthorized) {
			t.Errorf("%s: esperava ErrUnauthorized, veio %v", name, err)
		}
	}
	e.store.AddInvite("vencido", e.owner.ID, now.Add(-time.Second))
	if _, err := e.begin().Run(bg, passkeys.Subject{InviteToken: "vencido"}); !errors.Is(err, passkeys.ErrUnauthorized) {
		t.Fatalf("convite vencido: %v", err)
	}
	if e.challenges.Len() != 0 {
		t.Fatal("nenhum desafio pode ser guardado sem autorizacao")
	}
}

func TestChallengeIsSingleUseAndBoundToItsPurpose(t *testing.T) {
	e := newEnv(t, domain.RoleUser)
	st, _ := e.begin().Run(bg, passkeys.Subject{InviteToken: invite})
	resp, _ := identitytest.NewSoftAuthenticator(origin, rpID).Create(st.Options)
	in := passkeys.FinishRegistrationInput{ChallengeID: st.ChallengeID, Credential: resp}
	if _, err := e.finish().Run(bg, in); err != nil {
		t.Fatal(err)
	}
	if _, err := e.finish().Run(bg, in); !errors.Is(err, passkeys.ErrInvalid) {
		t.Fatalf("desafio reutilizado: esperava ErrInvalid, veio %v", err)
	}
	if _, err := e.finish().Run(bg, passkeys.FinishRegistrationInput{ChallengeID: "inventado", Credential: resp}); !errors.Is(err, passkeys.ErrInvalid) {
		t.Fatalf("desafio inventado: %v", err)
	}
	login, _ := e.loginBegin().Run(bg)
	if _, err := e.finish().Run(bg, passkeys.FinishRegistrationInput{ChallengeID: login.ChallengeID, Credential: resp}); !errors.Is(err, passkeys.ErrInvalid) {
		t.Fatalf("desafio de login nao pode concluir um registro: %v", err)
	}
}

func TestExpiredChallengeIsRejected(t *testing.T) {
	e := newEnv(t, domain.RoleUser)
	st, _ := e.begin().Run(bg, passkeys.Subject{InviteToken: invite})
	resp, _ := identitytest.NewSoftAuthenticator(origin, rpID).Create(st.Options)
	late := e.finish()
	late.Now = func() time.Time { return now.Add(passkeys.ChallengeTTL + time.Second) }
	if _, err := late.Run(bg, passkeys.FinishRegistrationInput{ChallengeID: st.ChallengeID, Credential: resp}); !errors.Is(err, passkeys.ErrInvalid) {
		t.Fatalf("desafio vencido: %v", err)
	}
}

func TestTwoRegistrationsRacingForTheSameInviteOnlyOneWins(t *testing.T) {
	e := newEnv(t, domain.RoleUser)
	first, _ := e.begin().Run(bg, passkeys.Subject{InviteToken: invite})
	second, _ := e.begin().Run(bg, passkeys.Subject{InviteToken: invite})
	r1, _ := identitytest.NewSoftAuthenticator(origin, rpID).Create(first.Options)
	r2, _ := identitytest.NewSoftAuthenticator(origin, rpID).Create(second.Options)
	if _, err := e.finish().Run(bg, passkeys.FinishRegistrationInput{ChallengeID: first.ChallengeID, Credential: r1}); err != nil {
		t.Fatal(err)
	}
	if _, err := e.finish().Run(bg, passkeys.FinishRegistrationInput{ChallengeID: second.ChallengeID, Credential: r2}); !errors.Is(err, passkeys.ErrUnauthorized) {
		t.Fatalf("o segundo registro com o mesmo convite deve falhar: %v", err)
	}
	if len(e.store.Credentials()) != 1 {
		t.Fatal("so uma passkey pode existir")
	}
}

func TestRegistrationRejectsAuthenticatorsWithoutUserVerification(t *testing.T) {
	e := newEnv(t, domain.RoleAdmin)
	a := identitytest.NewSoftAuthenticator(origin, rpID)
	a.NoUV = true
	st, _ := e.begin().Run(bg, passkeys.Subject{InviteToken: invite})
	resp, _ := a.Create(st.Options)
	if _, err := e.finish().Run(bg, passkeys.FinishRegistrationInput{ChallengeID: st.ChallengeID, Credential: resp}); !errors.Is(err, passkeys.ErrInvalid) {
		t.Fatalf("sem UV: %v", err)
	}
	if len(e.store.Credentials()) != 0 || e.store.InviteUsed(invite) {
		t.Fatal("nada pode ser gravado nem consumido")
	}
}

func TestSecondPasskeyViaSessionIsExcludedFromReuseAndOpensNoSession(t *testing.T) {
	e := newEnv(t, domain.RoleUser)
	e.registerWithInvite(t, identitytest.NewSoftAuthenticator(origin, rpID))
	before := e.sessions.Len()
	st, err := e.begin().Run(bg, passkeys.Subject{SessionUserID: e.owner.ID})
	if err != nil || !strings.Contains(string(st.Options), `"excludeCredentials":[{`) {
		t.Fatalf("a passkey existente deve estar na lista de exclusao: %v %s", err, st.Options)
	}
	resp, _ := identitytest.NewSoftAuthenticator(origin, rpID).Create(st.Options)
	got, err := e.finish().Run(bg, passkeys.FinishRegistrationInput{ChallengeID: st.ChallengeID, Credential: resp, Name: "  "})
	if err != nil || got.Session != nil || got.Credential.Name != passkeys.DefaultName {
		t.Fatalf("segunda passkey: %+v %v", got, err)
	}
	if e.sessions.Len() != before || len(e.store.Credentials()) != 2 {
		t.Fatal("nao abre sessao nova e a lista passa a ter 2")
	}
}

func TestPasskeyLimitPerAccount(t *testing.T) {
	e := newEnv(t, domain.RoleUser)
	for range passkeys.MaxPerAccount {
		e.registerSecond(t)
	}
	if _, err := e.begin().Run(bg, passkeys.Subject{SessionUserID: e.owner.ID}); !errors.Is(err, passkeys.ErrTooMany) {
		t.Fatalf("passou do limite: %v", err)
	}
}

func TestNameIsCleanedAndLimited(t *testing.T) {
	e := newEnv(t, domain.RoleUser)
	long := strings.Repeat("ã", 200)
	st, _ := e.begin().Run(bg, passkeys.Subject{InviteToken: invite})
	resp, _ := identitytest.NewSoftAuthenticator(origin, rpID).Create(st.Options)
	got, err := e.finish().Run(bg, passkeys.FinishRegistrationInput{ChallengeID: st.ChallengeID, Credential: resp, Name: "\x00\n" + long})
	if err != nil {
		t.Fatal(err)
	}
	if n := len([]rune(got.Credential.Name)); n != passkeys.MaxNameLength || strings.ContainsAny(got.Credential.Name, "\x00\n") {
		t.Fatalf("nome deve perder controles e ter no maximo %d caracteres, tem %d", passkeys.MaxNameLength, n)
	}
}

type brokenSessions struct{ *identitytest.Sessions }

func (brokenSessions) Create(context.Context, app.NewSession) error {
	return errors.New("neon fora do ar")
}

func TestRegistrationStillSucceedsWhenOnlyOpeningTheSessionFails(t *testing.T) {
	e := newEnv(t, domain.RoleUser)
	st, _ := e.begin().Run(bg, passkeys.Subject{InviteToken: invite})
	resp, _ := identitytest.NewSoftAuthenticator(origin, rpID).Create(st.Options)
	uc := e.finish()
	uc.Opener.Sessions = brokenSessions{e.sessions}
	got, err := uc.Run(bg, passkeys.FinishRegistrationInput{ChallengeID: st.ChallengeID, Credential: resp})
	if err != nil || got.Session != nil {
		t.Fatalf("a passkey ja foi gravada: deve voltar sucesso sem sessao, veio %+v %v", got, err)
	}
	if len(e.store.Credentials()) != 1 || !e.store.InviteUsed(invite) {
		t.Fatal("passkey gravada e convite consumido")
	}
}

func TestVeryLongNamesAreTruncatedQuickly(t *testing.T) {
	e := newEnv(t, domain.RoleUser)
	st, _ := e.begin().Run(bg, passkeys.Subject{InviteToken: invite})
	resp, _ := identitytest.NewSoftAuthenticator(origin, rpID).Create(st.Options)
	start := time.Now()
	got, err := e.finish().Run(bg, passkeys.FinishRegistrationInput{ChallengeID: st.ChallengeID, Credential: resp, Name: strings.Repeat("é", 60000)})
	if err != nil || len([]rune(got.Credential.Name)) != passkeys.MaxNameLength {
		t.Fatalf("nome: %v %d", err, len([]rune(got.Credential.Name)))
	}
	if time.Since(start) > 2*time.Second {
		t.Fatalf("truncar um nome de 60 mil caracteres deve ser instantaneo, levou %v", time.Since(start))
	}
}
