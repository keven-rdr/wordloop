package app_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/keven-rdr/wordloop/api/internal/identity/app"
	"github.com/keven-rdr/wordloop/api/internal/identity/domain"
	"github.com/keven-rdr/wordloop/api/internal/identity/identitytest"
	"github.com/keven-rdr/wordloop/api/internal/platform/seal"
)

var now = time.Date(2026, 10, 13, 12, 0, 0, 0, time.UTC)

type env struct {
	provider *identitytest.Provider
	users    *identitytest.Users
	sessions *identitytest.Sessions
	sealer   *seal.Sealer
}

func newEnv(t *testing.T) env {
	t.Helper()
	sealer, err := seal.New(bytes.Repeat([]byte{3}, seal.KeySize))
	if err != nil {
		t.Fatal(err)
	}
	return env{
		provider: &identitytest.Provider{Identity: domain.Identity{Issuer: "https://idp.test", Subject: "abc", DisplayName: "Aluno"}, Refresh: "refresh-1"},
		users:    &identitytest.Users{}, sessions: &identitytest.Sessions{}, sealer: sealer,
	}
}

func (e env) start() app.StartLogin { return app.StartLogin{Provider: e.provider, Sealer: e.sealer} }

func (e env) finish() app.FinishLogin {
	return app.FinishLogin{Provider: e.provider, Users: e.users, Sessions: e.sessions, Sealer: e.sealer, Now: identitytest.Clock(now)}
}

func (e env) flow(t *testing.T, cookie string) app.Flow {
	t.Helper()
	raw, err := e.sealer.Open(cookie)
	if err != nil {
		t.Fatal(err)
	}
	var f app.Flow
	if err := json.Unmarshal(raw, &f); err != nil {
		t.Fatal(err)
	}
	return f
}

func TestStartLoginSealsStateAndKeepsOnlyRelativeReturnTo(t *testing.T) {
	cases := map[string]string{
		"/estudo?x=1":      "/estudo?x=1",
		"":                 "/",
		"https://evil.com": "/",
		"//evil.com":       "/",
		"/\\evil.com":      "/",
		"/a\r\nSet-Cookie": "/",
	}
	for in, want := range cases {
		e := newEnv(t)
		got, err := e.start().Run(context.Background(), in)
		if err != nil {
			t.Fatal(err)
		}
		f := e.flow(t, got.FlowCookie)
		if f.ReturnTo != want || f.State != e.provider.LastState || f.Nonce != e.provider.LastNonce || f.Verifier != e.provider.LastVerifier {
			t.Errorf("return_to %q: flow = %+v, esperava return_to %q", in, f, want)
		}
		if f.State == "" || f.State == f.Nonce || f.Nonce == f.Verifier {
			t.Errorf("state, nonce e verifier precisam ser distintos e nao vazios: %+v", f)
		}
	}
}

func login(t *testing.T, e env, returnTo string) (app.LoggedIn, string) {
	t.Helper()
	started, err := e.start().Run(context.Background(), returnTo)
	if err != nil {
		t.Fatal(err)
	}
	f := e.flow(t, started.FlowCookie)
	done, err := e.finish().Run(context.Background(), app.FinishInput{
		Code: identitytest.GoodCode, State: f.State, FlowCookie: started.FlowCookie, UserAgent: "teste",
	})
	if err != nil {
		t.Fatal(err)
	}
	return done, started.FlowCookie
}

func TestFinishLoginOpensSessionAndStoresOnlyHashAndSealedRefresh(t *testing.T) {
	e := newEnv(t)
	done, _ := login(t, e, "/estudo")
	if done.ReturnTo != "/estudo" || !done.ExpiresAt.Equal(now.Add(app.DefaultSessionTTL)) {
		t.Fatalf("resultado inesperado: %+v", done)
	}
	stored, ok := e.sessions.Last()
	if !ok || e.sessions.Len() != 1 {
		t.Fatal("esperava exatamente uma sessao gravada")
	}
	if bytes.Contains(stored.IDHash, []byte(done.SessionID)) || len(stored.IDHash) != 32 {
		t.Fatal("o banco deve guardar o hash SHA-256 e nunca o id da sessao")
	}
	if stored.RefreshTokenEnc == "refresh-1" {
		t.Fatal("o refresh token nao pode ir em claro")
	}
	if plain, err := e.sealer.Open(stored.RefreshTokenEnc); err != nil || string(plain) != "refresh-1" {
		t.Fatalf("refresh selado nao abre: %q %v", plain, err)
	}
	if stored.UserAgent != "teste" {
		t.Fatalf("user agent = %q", stored.UserAgent)
	}
}

func TestFinishLoginRejectsInvalidRequests(t *testing.T) {
	e := newEnv(t)
	started, _ := e.start().Run(context.Background(), "/")
	f := e.flow(t, started.FlowCookie)
	other, _ := newEnv(t).start().Run(context.Background(), "/")
	cases := map[string]app.FinishInput{
		"erro do provedor":   {Code: identitytest.GoodCode, State: f.State, FlowCookie: started.FlowCookie, ProviderError: "access_denied"},
		"sem code":           {State: f.State, FlowCookie: started.FlowCookie},
		"sem state":          {Code: identitytest.GoodCode, FlowCookie: started.FlowCookie},
		"sem cookie":         {Code: identitytest.GoodCode, State: f.State},
		"state diferente":    {Code: identitytest.GoodCode, State: "outro", FlowCookie: started.FlowCookie},
		"cookie adulterado":  {Code: identitytest.GoodCode, State: f.State, FlowCookie: started.FlowCookie + "x"},
		"cookie de outro":    {Code: identitytest.GoodCode, State: f.State, FlowCookie: other.FlowCookie},
		"provedor recusa":    {Code: "code-ruim", State: f.State, FlowCookie: started.FlowCookie},
		"cookie nao e valor": {Code: identitytest.GoodCode, State: f.State, FlowCookie: "***"},
	}
	for name, in := range cases {
		if _, err := e.finish().Run(context.Background(), in); !errors.Is(err, app.ErrInvalidLogin) {
			t.Errorf("%s: esperava ErrInvalidLogin, veio %v", name, err)
		}
	}
	if e.sessions.Len() != 0 {
		t.Fatal("nenhuma sessao pode ser criada em login invalido")
	}
}

func TestGetMeReturnsProfileForValidSessionAndNoSessionOtherwise(t *testing.T) {
	e := newEnv(t)
	done, _ := login(t, e, "/")
	me := app.GetMe{Sessions: e.sessions, Users: e.users, Now: identitytest.Clock(now.Add(time.Hour))}
	user, err := me.Run(context.Background(), done.SessionID)
	if err != nil || user.DisplayName != "Aluno" {
		t.Fatalf("GetMe = %+v, %v", user, err)
	}
	if e.sessions.Touched != 1 {
		t.Fatalf("esperava 1 atualizacao do ultimo uso, veio %d", e.sessions.Touched)
	}
	for name, id := range map[string]string{"vazio": "", "desconhecido": "naoexiste"} {
		if _, err := me.Run(context.Background(), id); !errors.Is(err, app.ErrNoSession) {
			t.Errorf("%s: esperava ErrNoSession, veio %v", name, err)
		}
	}
}

func TestGetMeExpiredSessionIsDeleted(t *testing.T) {
	e := newEnv(t)
	done, _ := login(t, e, "/")
	late := app.GetMe{Sessions: e.sessions, Users: e.users, Now: identitytest.Clock(now.Add(app.DefaultSessionTTL + time.Second))}
	if _, err := late.Run(context.Background(), done.SessionID); !errors.Is(err, app.ErrNoSession) {
		t.Fatalf("esperava ErrNoSession, veio %v", err)
	}
	if e.sessions.Len() != 0 {
		t.Fatal("a sessao expirada deve ser apagada")
	}
}

func TestGetMeSessionOfRemovedUserIsNoSession(t *testing.T) {
	e := newEnv(t)
	done, _ := login(t, e, "/")
	stored, _ := e.sessions.Last()
	e.users.Remove(stored.UserID)
	me := app.GetMe{Sessions: e.sessions, Users: e.users, Now: identitytest.Clock(now)}
	if _, err := me.Run(context.Background(), done.SessionID); !errors.Is(err, app.ErrNoSession) {
		t.Fatalf("esperava ErrNoSession, veio %v", err)
	}
}

func TestLogoutDeletesSessionAndIsIdempotent(t *testing.T) {
	e := newEnv(t)
	done, _ := login(t, e, "/")
	out := app.Logout{Sessions: e.sessions}
	for i := 0; i < 2; i++ {
		if err := out.Run(context.Background(), done.SessionID); err != nil {
			t.Fatal(err)
		}
	}
	if err := out.Run(context.Background(), ""); err != nil {
		t.Fatal(err)
	}
	if e.sessions.Len() != 0 {
		t.Fatal("a sessao deveria ter sido apagada")
	}
}
