package app_test

import (
	"errors"
	"slices"
	"testing"

	"github.com/keven-rdr/wordloop/api/internal/identity/app"
	"github.com/keven-rdr/wordloop/api/internal/identity/domain"
	"github.com/keven-rdr/wordloop/api/internal/identity/identitytest"
)

func social(e env, who string) *identitytest.Provider {
	return &identitytest.Provider{Identity: domain.Identity{Issuer: who, Subject: who + "-1", DisplayName: who, Email: who + "@exemplo.com", EmailVerified: true}}
}

func twoProviders(e env) (app.StartLogin, app.FinishLogin, *identitytest.Provider, *identitytest.Provider) {
	google, github := social(e, "google"), social(e, "github")
	providers := app.Providers{"google": google, "github": github}
	start := app.StartLogin{Providers: providers, Sealer: e.sealer}
	finish := app.FinishLogin{Providers: providers, Users: e.users, Sessions: e.sessions, Sealer: e.sealer, Now: identitytest.Clock(now)}
	return start, finish, google, github
}

func TestTheChosenProviderTravelsSealedInTheFlowAndHandlesTheCallback(t *testing.T) {
	e := newEnv(t)
	start, finish, google, github := twoProviders(e)
	started, err := start.RunProvider(bg, "github", "/estudar")
	if err != nil {
		t.Fatal(err)
	}
	flow := e.flow(t, started.FlowCookie)
	if flow.Provider != "github" || github.LastState != flow.State || google.LastState != "" {
		t.Fatalf("fluxo: %+v google=%q github=%q", flow, google.LastState, github.LastState)
	}
	done, err := finish.Run(bg, app.FinishInput{Code: identitytest.GoodCode, State: flow.State, FlowCookie: started.FlowCookie})
	if err != nil || done.ReturnTo != "/estudar" {
		t.Fatalf("callback: %+v %v", done, err)
	}
	stored, _ := e.sessions.Last()
	user, _ := e.users.Get(bg, stored.UserID)
	if user.DisplayName != "github" {
		t.Fatalf("a conta veio do provedor errado: %+v", user)
	}
}

func TestUnknownProvidersAreRefusedAndTheListIsSorted(t *testing.T) {
	e := newEnv(t)
	start, _, _, _ := twoProviders(e)
	for _, name := range []string{"facebook", "oidc", "GITHUB"} {
		if _, err := start.RunProvider(bg, name, "/"); !errors.Is(err, app.ErrUnknownProvider) {
			t.Errorf("%q: esperava ErrUnknownProvider, veio %v", name, err)
		}
	}
	if names := start.Names(); !slices.Equal(names, []string{"github", "google"}) {
		t.Fatalf("lista: %v", names)
	}
}

func TestTheSingleDefaultProviderStillWorksAndShowsUpAsOidc(t *testing.T) {
	e := newEnv(t)
	if names := e.start().Names(); !slices.Equal(names, []string{"oidc"}) {
		t.Fatalf("lista: %v", names)
	}
	if _, err := e.start().RunProvider(bg, "oidc", "/"); err != nil {
		t.Fatalf("o nome padrao funciona: %v", err)
	}
	if _, err := e.start().RunProvider(bg, "google", "/"); !errors.Is(err, app.ErrUnknownProvider) {
		t.Fatalf("outro nome sem registro: %v", err)
	}
}

func TestACookieNamingAProviderThatIsNoLongerConfiguredIsInvalid(t *testing.T) {
	e := newEnv(t)
	start, _, _, _ := twoProviders(e)
	started, _ := start.RunProvider(bg, "google", "/")
	onlyGithub := app.FinishLogin{Providers: app.Providers{"github": social(e, "github")}, Users: e.users, Sessions: e.sessions, Sealer: e.sealer}
	in := app.FinishInput{Code: identitytest.GoodCode, State: e.flow(t, started.FlowCookie).State, FlowCookie: started.FlowCookie}
	if _, err := onlyGithub.Run(bg, in); !errors.Is(err, app.ErrInvalidLogin) {
		t.Fatalf("esperava ErrInvalidLogin, veio %v", err)
	}
}

func TestAProviderWithoutAVerifiedEmailCannotSignIn(t *testing.T) {
	e := newEnv(t)
	e.users.RequireVerifiedEmail = true
	start, finish, _, github := twoProviders(e)
	github.Identity.EmailVerified = false
	started, _ := start.RunProvider(bg, "github", "/")
	in := app.FinishInput{Code: identitytest.GoodCode, State: e.flow(t, started.FlowCookie).State, FlowCookie: started.FlowCookie}
	if _, err := finish.Run(bg, in); !errors.Is(err, app.ErrEmailNotVerified) {
		t.Fatalf("esperava ErrEmailNotVerified, veio %v", err)
	}
	if e.sessions.Len() != 0 {
		t.Fatal("nenhuma sessao pode nascer sem e-mail confirmado")
	}
}
