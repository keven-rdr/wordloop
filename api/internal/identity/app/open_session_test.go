package app_test

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/keven-rdr/wordloop/api/internal/identity/app"
	"github.com/keven-rdr/wordloop/api/internal/identity/domain"
	"github.com/keven-rdr/wordloop/api/internal/identity/identitytest"
)

func TestSessionOpenerWithoutRefreshTokenNeedsNoSealerAndStoresNothingSecret(t *testing.T) {
	sessions := &identitytest.Sessions{}
	opener := app.SessionOpener{Sessions: sessions, Now: identitytest.Clock(now)}
	user := domain.User{ID: uuid.New(), Role: domain.RoleUser}
	opened, err := opener.Open(bg, app.Authentication{User: user, MFA: true, AuthTime: now.Add(-time.Minute), UserAgent: "ua"})
	if err != nil || opened.SessionID == "" {
		t.Fatalf("Open: %+v %v", opened, err)
	}
	stored, _ := sessions.Last()
	if stored.RefreshTokenEnc != "" || !stored.MFA || !stored.AuthenticatedAt.Equal(now.Add(-time.Minute)) || stored.UserAgent != "ua" {
		t.Fatalf("sessao gravada: %+v", stored)
	}
	if want := now.Add(domain.DefaultSessionPolicy().User.Absolute); !opened.ExpiresAt.Equal(want) {
		t.Fatalf("expira em %v, esperava %v", opened.ExpiresAt, want)
	}
}

func TestSessionOpenerLetsAnAdminInWithoutSecondFactorUnlessTheAccountAsksForIt(t *testing.T) {
	sessions := &identitytest.Sessions{}
	opener := app.SessionOpener{Sessions: sessions, Now: identitytest.Clock(now)}
	admin := domain.User{ID: uuid.New(), Role: domain.RoleAdmin}
	if _, err := opener.Open(bg, app.Authentication{User: admin}); err != nil {
		t.Fatalf("2FA vem desligado: o admin entra sem ele, veio %v", err)
	}
}

func TestSessionOpenerRefusesAnAccountThatTurnedOnTwoFactorWithoutProofAndCreatesNothing(t *testing.T) {
	sessions := &identitytest.Sessions{}
	opener := app.SessionOpener{Sessions: sessions, Now: identitytest.Clock(now)}
	admin := domain.User{ID: uuid.New(), Role: domain.RoleAdmin, MFARequired: true}
	if _, err := opener.Open(bg, app.Authentication{User: admin}); !errors.Is(err, app.ErrMFARequired) {
		t.Fatalf("2FA ligado e sem prova: %v", err)
	}
	if sessions.Len() != 0 {
		t.Fatal("nenhuma sessao pode nascer")
	}
	opened, err := opener.Open(bg, app.Authentication{User: admin, MFA: true})
	if err != nil {
		t.Fatal(err)
	}
	if want := now.Add(domain.DefaultSessionPolicy().Admin.Absolute); !opened.ExpiresAt.Equal(want) {
		t.Fatalf("a sessao do admin e curta: %v, esperava %v", opened.ExpiresAt, want)
	}
}

func TestSessionOpenerClampsAFutureAuthTime(t *testing.T) {
	sessions := &identitytest.Sessions{}
	opener := app.SessionOpener{Sessions: sessions, Now: identitytest.Clock(now)}
	if _, err := opener.Open(bg, app.Authentication{User: domain.User{ID: uuid.New()}, AuthTime: now.Add(time.Hour)}); err != nil {
		t.Fatal(err)
	}
	if stored, _ := sessions.Last(); !stored.AuthenticatedAt.Equal(now) {
		t.Fatalf("auth_time no futuro vira agora: %v", stored.AuthenticatedAt)
	}
}
