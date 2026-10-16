package httpapi_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/keven-rdr/wordloop/api/internal/api"
	"github.com/keven-rdr/wordloop/api/internal/identity/accounts"
	httpapi "github.com/keven-rdr/wordloop/api/internal/identity/adapters/http"
	"github.com/keven-rdr/wordloop/api/internal/identity/app"
	"github.com/keven-rdr/wordloop/api/internal/identity/domain"
	"github.com/keven-rdr/wordloop/api/internal/identity/identitytest"
	"github.com/keven-rdr/wordloop/api/internal/identity/passwords"
	"github.com/keven-rdr/wordloop/api/internal/identity/registration"
	"github.com/keven-rdr/wordloop/api/internal/platform/seal"
)

type fixture struct {
	h        http.Handler
	provider *identitytest.Provider
	users    *identitytest.Users
	sessions *identitytest.Sessions
	pw       *identitytest.PasswordStore
	reg      *identitytest.RegistrationStore
	mailbox  *identitytest.MailBox
}

func newFixture(t *testing.T, enabled bool) fixture {
	t.Helper()
	sealer, err := seal.New(bytes.Repeat([]byte{5}, seal.KeySize))
	if err != nil {
		t.Fatal(err)
	}
	provider := &identitytest.Provider{Identity: domain.Identity{Issuer: "https://idp.test", Subject: "abc", DisplayName: "Aluno"}, Refresh: "r"}
	users, sessions := &identitytest.Users{}, &identitytest.Sessions{}
	auth := app.Authenticate{Sessions: sessions, Users: users}
	h := httpapi.Handler{
		Auth: auth, Out: app.Logout{Sessions: sessions}, Secure: true,
		List:   app.ListSessions{Auth: auth, Sessions: sessions},
		Revoke: app.RevokeSession{Auth: auth, Sessions: sessions}, RevokeAll: app.RevokeAllSessions{Auth: auth, Sessions: sessions},
		TwoFactor: accounts.TwoFactor{Store: users},
	}
	pw, hasher, opener := &identitytest.PasswordStore{}, &identitytest.PlainHasher{}, app.SessionOpener{Sessions: sessions}
	h.Passwords = &httpapi.Passwords{
		Login: passwords.Login{Store: pw, Hasher: hasher, Users: users, Opener: opener, DummyHash: "h:isca"},
		Set:   passwords.SetPassword{Store: pw, Hasher: hasher, Users: users, Opener: opener},
	}
	reg, box := &identitytest.RegistrationStore{}, &identitytest.MailBox{}
	issuer := registration.Issuer{Store: reg, Mail: box, PublicURL: "https://app.test"}
	h.Registration = &httpapi.Registration{Register: registration.Register{Issuer: issuer}, Forgot: registration.Forgot{Issuer: issuer}}
	if enabled {
		h.Start = &app.StartLogin{Provider: provider, Sealer: sealer}
		h.Finish = &app.FinishLogin{Provider: provider, Users: users, Sessions: sessions, Sealer: sealer}
	}
	return fixture{h: api.Handler(api.Server{Identity: h}), provider: provider, users: users, sessions: sessions, pw: pw, reg: reg, mailbox: box}
}

func newHandler(t *testing.T, enabled bool) http.Handler { return newFixture(t, enabled).h }

func do(h http.Handler, method, target string, headers map[string]string, cookies ...*http.Cookie) *http.Response {
	req := httptest.NewRequest(method, target, nil)
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	for _, c := range cookies {
		req.AddCookie(c)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec.Result()
}

func cookie(t *testing.T, res *http.Response, name string) *http.Cookie {
	t.Helper()
	for _, c := range res.Cookies() {
		if c.Name == name {
			return c
		}
	}
	t.Fatalf("cookie %s ausente; veio %v", name, res.Cookies())
	return nil
}

// loginFlow faz login ate o callback e devolve o cookie da sessao.
func loginFlow(t *testing.T, h http.Handler) *http.Cookie {
	t.Helper()
	res := do(h, http.MethodGet, "/api/v1/auth/login?return_to=/estudo", nil)
	if res.StatusCode != http.StatusFound {
		t.Fatalf("login: esperava 302, veio %d", res.StatusCode)
	}
	loc, err := url.Parse(res.Header.Get("Location"))
	if err != nil || loc.Host != "idp.test" {
		t.Fatalf("Location inesperado: %q", res.Header.Get("Location"))
	}
	flow := cookie(t, res, httpapi.FlowCookieSecure)
	if !flow.HttpOnly || !flow.Secure || flow.SameSite != http.SameSiteLaxMode || flow.Path != "/api/v1/auth" || flow.MaxAge <= 0 {
		t.Fatalf("cookie de estado sem os atributos de seguranca: %+v", flow)
	}
	cb := do(h, http.MethodGet, "/api/v1/auth/callback?code="+identitytest.GoodCode+"&state="+loc.Query().Get("state"), nil, flow)
	if cb.StatusCode != http.StatusFound || cb.Header.Get("Location") != "/estudo" {
		t.Fatalf("callback: %d Location=%q", cb.StatusCode, cb.Header.Get("Location"))
	}
	if cleared := cookie(t, cb, httpapi.FlowCookieSecure); cleared.MaxAge >= 0 {
		t.Fatalf("o cookie de estado deveria ser apagado: %+v", cleared)
	}
	return cookie(t, cb, httpapi.SessionCookieSecure)
}

func TestFullLoginMeLogoutFlow(t *testing.T) {
	h := newHandler(t, true)
	session := loginFlow(t, h)
	if !session.HttpOnly || !session.Secure || session.SameSite != http.SameSiteLaxMode || session.Path != "/" || session.MaxAge <= 0 {
		t.Fatalf("cookie da sessao sem os atributos de seguranca: %+v", session)
	}

	me := do(h, http.MethodGet, "/api/v1/me", nil, session)
	var body map[string]any
	if err := json.NewDecoder(me.Body).Decode(&body); err != nil || me.StatusCode != http.StatusOK || body["display_name"] != "Aluno" || body["id"] == "" || body["role"] != "user" {
		t.Fatalf("/me: %d %v %v", me.StatusCode, body, err)
	}

	if res := do(h, http.MethodPost, "/api/v1/auth/logout", nil, session); res.StatusCode != http.StatusBadRequest {
		t.Fatalf("logout sem o cabecalho anti-CSRF: esperava 400, veio %d", res.StatusCode)
	}
	out := do(h, http.MethodPost, "/api/v1/auth/logout", map[string]string{"X-Requested-With": "XMLHttpRequest"}, session)
	if out.StatusCode != http.StatusNoContent || cookie(t, out, httpapi.SessionCookieSecure).MaxAge >= 0 {
		t.Fatalf("logout: %d", out.StatusCode)
	}
	if res := do(h, http.MethodGet, "/api/v1/me", nil, session); res.StatusCode != http.StatusUnauthorized {
		t.Fatalf("/me depois do logout: esperava 401, veio %d", res.StatusCode)
	}
}

func TestMeWithoutOrWithBogusSessionIs401(t *testing.T) {
	h := newHandler(t, true)
	for name, cs := range map[string][]*http.Cookie{"sem cookie": nil, "cookie falso": {{Name: httpapi.SessionCookieSecure, Value: "falso"}}} {
		if res := do(h, http.MethodGet, "/api/v1/me", nil, cs...); res.StatusCode != http.StatusUnauthorized {
			t.Errorf("%s: esperava 401, veio %d", name, res.StatusCode)
		}
	}
}

func TestCallbackRejectsInvalidLoginAndClearsStateCookie(t *testing.T) {
	h := newHandler(t, true)
	res := do(h, http.MethodGet, "/api/v1/auth/callback?code=x&state=y", nil, &http.Cookie{Name: httpapi.FlowCookieSecure, Value: "lixo"})
	if res.StatusCode != http.StatusFound || res.Header.Get("Location") != "/login?error=failed" || cookie(t, res, httpapi.FlowCookieSecure).MaxAge >= 0 {
		t.Fatalf("callback invalido: %d %s", res.StatusCode, res.Header.Get("Location"))
	}
	if res := do(h, http.MethodGet, "/api/v1/auth/callback?error=access_denied", nil); res.StatusCode != http.StatusFound || res.Header.Get("Location") != "/login?error=failed" {
		t.Fatalf("callback com erro do provedor: esperava o redirecionamento ao login, veio %d %s", res.StatusCode, res.Header.Get("Location"))
	}
}

func TestLoginIsUnavailableWhenNotConfigured(t *testing.T) {
	h := newHandler(t, false)
	for _, path := range []string{"/api/v1/auth/login", "/api/v1/auth/callback?code=x&state=y"} {
		if res := do(h, http.MethodGet, path, nil); res.StatusCode != http.StatusServiceUnavailable {
			t.Errorf("%s: esperava 503, veio %d", path, res.StatusCode)
		}
	}
	if res := do(h, http.MethodGet, "/api/v1/me", nil); res.StatusCode != http.StatusUnauthorized {
		t.Errorf("/me: esperava 401, veio %d", res.StatusCode)
	}
}

func TestSessionCookieIsNotSecureOverPlainHTTP(t *testing.T) {
	sealer, _ := seal.New(bytes.Repeat([]byte{5}, seal.KeySize))
	provider := &identitytest.Provider{}
	h := api.Handler(api.Server{Identity: httpapi.Handler{
		Start:  &app.StartLogin{Provider: provider, Sealer: sealer},
		Finish: &app.FinishLogin{Provider: provider, Sealer: sealer},
		Secure: false,
	}})
	res := do(h, http.MethodGet, "/api/v1/auth/login", nil)
	if c := cookie(t, res, httpapi.FlowCookie); c.Secure || !c.HttpOnly {
		t.Fatalf("em http local o cookie nao pode ser Secure (o navegador o descartaria): %+v", c)
	}
	if !strings.HasPrefix(res.Header.Get("Location"), "https://idp.test/") {
		t.Fatalf("Location = %q", res.Header.Get("Location"))
	}
}

func TestMeReturnsThePictureOnlyWhenThereIsOne(t *testing.T) {
	f := newFixture(t, true)
	session := loginFlow(t, f.h)
	var body map[string]any
	_ = json.NewDecoder(do(f.h, http.MethodGet, "/api/v1/me", nil, session).Body).Decode(&body)
	if _, has := body["picture_url"]; has {
		t.Fatalf("sem foto o campo nao aparece: %v", body)
	}
	f.provider.Identity.PictureURL = "https://lh3.googleusercontent.com/a/x"
	session = loginFlow(t, f.h)
	body = nil
	_ = json.NewDecoder(do(f.h, http.MethodGet, "/api/v1/me", nil, session).Body).Decode(&body)
	if body["picture_url"] != "https://lh3.googleusercontent.com/a/x" {
		t.Fatalf("com foto: %v", body)
	}
}
