package httpapi_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"testing"

	"github.com/keven-rdr/wordloop/api/internal/api"
	httpapi "github.com/keven-rdr/wordloop/api/internal/identity/adapters/http"
)

func providersOf(t *testing.T, h http.Handler) []string {
	t.Helper()
	res := do(h, http.MethodGet, "/api/v1/auth/providers", nil)
	var body struct {
		Providers []string `json:"providers"`
	}
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil || res.StatusCode != http.StatusOK || body.Providers == nil {
		t.Fatalf("providers: %d %v %v", res.StatusCode, body, err)
	}
	return body.Providers
}

func TestTheScreenLearnsWhichSocialProvidersAreOn(t *testing.T) {
	if got := providersOf(t, newHandler(t, true)); len(got) != 1 || got[0] != "oidc" {
		t.Fatalf("com o provedor padrao: %v", got)
	}
	if got := providersOf(t, newHandler(t, false)); len(got) != 0 {
		t.Fatalf("login social desligado: %v", got)
	}
	if got := providersOf(t, api.Handler(api.Server{})); len(got) != 0 {
		t.Fatalf("sem o modulo de identidade: %v", got)
	}
	if got := providersOf(t, api.Handler(api.Server{Identity: httpapi.Handler{Secure: true}})); len(got) != 0 {
		t.Fatalf("modulo sem provedores: %v", got)
	}
}

func TestStartingLoginWithAnUnknownProviderIsABadRequest(t *testing.T) {
	h := newHandler(t, true)
	if res := do(h, http.MethodGet, "/api/v1/auth/login?provider=oidc", nil); res.StatusCode != http.StatusFound {
		t.Fatalf("provedor ligado: %d", res.StatusCode)
	}
	if res := do(h, http.MethodGet, "/api/v1/auth/login?provider=google", nil); res.StatusCode != http.StatusBadRequest {
		t.Fatalf("provedor desligado: esperava 400, veio %d", res.StatusCode)
	}
	if res := do(h, http.MethodGet, "/api/v1/auth/login?provider=facebook", nil); res.StatusCode != http.StatusBadRequest {
		t.Fatalf("provedor fora da lista: esperava 400, veio %d", res.StatusCode)
	}
}

func TestACallbackWithoutAVerifiedEmailGoesBackToTheLoginAndOpensNoSession(t *testing.T) {
	fx := newFixture(t, true)
	fx.users.RequireVerifiedEmail = true
	res := do(fx.h, http.MethodGet, "/api/v1/auth/login", nil)
	flow := cookie(t, res, httpapi.FlowCookieSecure)
	location := res.Header.Get("Location")
	state := location[strings.Index(location, "state=")+len("state="):]
	cb := do(fx.h, http.MethodGet, "/api/v1/auth/callback?code=code-ok&state="+state, nil, flow)
	if cb.StatusCode != http.StatusFound || cb.Header.Get("Location") != "/login?error=email" {
		t.Fatalf("esperava voltar ao login com error=email, veio %d %s", cb.StatusCode, cb.Header.Get("Location"))
	}
	for _, c := range cb.Cookies() {
		if c.Name == httpapi.SessionCookieSecure && c.MaxAge >= 0 {
			t.Fatal("sem e-mail confirmado nao ha sessao")
		}
	}
	if fx.sessions.Len() != 0 {
		t.Fatal("nenhuma sessao no banco")
	}
}

func TestAnUnexpectedErrorInTheCallbackAlsoGoesBackToTheLogin(t *testing.T) {
	fx := newFixture(t, true)
	fx.users.FailUpsert = errors.New("banco fora do ar")
	res := do(fx.h, http.MethodGet, "/api/v1/auth/login", nil)
	flow := cookie(t, res, httpapi.FlowCookieSecure)
	location := res.Header.Get("Location")
	state := location[strings.Index(location, "state=")+len("state="):]
	cb := do(fx.h, http.MethodGet, "/api/v1/auth/callback?code=code-ok&state="+state, nil, flow)
	if cb.StatusCode != http.StatusFound || cb.Header.Get("Location") != "/login?error=failed" {
		t.Fatalf("esperava voltar ao login, veio %d %s", cb.StatusCode, cb.Header.Get("Location"))
	}
}
