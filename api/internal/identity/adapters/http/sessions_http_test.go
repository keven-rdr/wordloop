package httpapi_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"

	httpapi "github.com/keven-rdr/wordloop/api/internal/identity/adapters/http"
	"github.com/keven-rdr/wordloop/api/internal/identity/domain"
)

var csrf = map[string]string{"X-Requested-With": "XMLHttpRequest"}

type meBody struct {
	ID          string `json:"id"`
	DisplayName string `json:"display_name"`
	Role        string `json:"role"`

	TwoFactorEnabled bool `json:"two_factor_enabled"`
}

func getMe(t *testing.T, h http.Handler, cs ...*http.Cookie) (int, meBody) {
	t.Helper()
	res := do(h, http.MethodGet, "/api/v1/me", nil, cs...)
	var b meBody
	_ = json.NewDecoder(res.Body).Decode(&b)
	return res.StatusCode, b
}

type sessionBody struct {
	ID      string `json:"id"`
	Current bool   `json:"current"`
}

func listSessions(t *testing.T, h http.Handler, c *http.Cookie) []sessionBody {
	t.Helper()
	res := do(h, http.MethodGet, "/api/v1/me/sessions", nil, c)
	var out []sessionBody
	if err := json.NewDecoder(res.Body).Decode(&out); err != nil || res.StatusCode != http.StatusOK {
		t.Fatalf("GET /me/sessions: %d %v", res.StatusCode, err)
	}
	return out
}

func loginAs(t *testing.T, fx fixture, subject string) *http.Cookie {
	t.Helper()
	fx.provider.Identity = domain.Identity{Issuer: "https://idp.test", Subject: subject, DisplayName: subject}
	return loginFlow(t, fx.h)
}

func uuidOf(t *testing.T, s string) uuid.UUID {
	t.Helper()
	id, err := uuid.Parse(s)
	if err != nil {
		t.Fatal(err)
	}
	return id
}

func TestSessionCookieUsesHostPrefixOverHTTPS(t *testing.T) {
	c := loginFlow(t, newHandler(t, true))
	if c.Name != "__Host-wl_session" || !c.Secure || !c.HttpOnly || c.Path != "/" || c.Domain != "" {
		t.Fatalf("o cookie da sessao deve cumprir as regras do prefixo __Host-: %+v", c)
	}
}

func TestUnprefixedCookieIsIgnoredOverHTTPS(t *testing.T) {
	fx := newFixture(t, true)
	session := loginFlow(t, fx.h)
	tossed := &http.Cookie{Name: httpapi.SessionCookie, Value: session.Value}
	if code, _ := getMe(t, fx.h, tossed); code != http.StatusUnauthorized {
		t.Fatalf("cookie sem prefixo deveria ser ignorado, veio %d", code)
	}
}

func TestForgedSessionCookieIsRejectedEverywhere(t *testing.T) {
	h := newHandler(t, true)
	forged := &http.Cookie{Name: httpapi.SessionCookieSecure, Value: "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"}
	const id = "/api/v1/me/sessions/3f0c4f4e-8d1e-4a3b-9a55-1d2c3b4a5e6f"
	cases := []struct {
		method, path string
		headers      map[string]string
	}{
		{http.MethodGet, "/api/v1/me", nil},
		{http.MethodGet, "/api/v1/me/sessions", nil},
		{http.MethodDelete, id, csrf},
		{http.MethodDelete, "/api/v1/me/sessions", csrf},
	}
	for _, c := range cases {
		if res := do(h, c.method, c.path, c.headers, forged); res.StatusCode != http.StatusUnauthorized {
			t.Errorf("%s %s com cookie inventado: esperava 401, veio %d", c.method, c.path, res.StatusCode)
		}
	}
}

func TestNothingInTheRequestChangesWhoYouAreOrYourRole(t *testing.T) {
	fx := newFixture(t, true)
	session := loginFlow(t, fx.h)
	_, before := getMe(t, fx.h, session)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/me?role=admin&user_id=00000000-0000-0000-0000-000000000001&admin=true", strings.NewReader(`{"role":"admin","id":"00000000-0000-0000-0000-000000000001"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-Id", "00000000-0000-0000-0000-000000000001")
	req.Header.Set("X-Role", "admin")
	req.AddCookie(session)
	rec := httptest.NewRecorder()
	fx.h.ServeHTTP(rec, req)
	var after meBody
	if err := json.NewDecoder(rec.Result().Body).Decode(&after); err != nil || after != before || after.Role != "user" {
		t.Fatalf("o pedido forjado mudou a identidade: antes=%+v depois=%+v (%v)", before, after, err)
	}
}

func TestRoleOnlyChangesInTheDatabaseAndShowsInMe(t *testing.T) {
	fx := newFixture(t, true)
	session := loginFlow(t, fx.h)
	_, me := getMe(t, fx.h, session)
	fx.users.SetRole(uuidOf(t, me.ID), domain.RoleAdmin)
	if _, after := getMe(t, fx.h, session); after.Role != "admin" {
		t.Fatalf("o papel vem do banco a cada pedido; esperava admin, veio %q", after.Role)
	}
}

func TestAdminWithTwoFactorOffEntersWithoutProofOfIt(t *testing.T) {
	fx := newFixture(t, true)
	first := loginFlow(t, fx.h)
	_, me := getMe(t, fx.h, first)
	fx.users.SetRole(uuidOf(t, me.ID), domain.RoleAdmin)
	do(fx.h, http.MethodDelete, "/api/v1/me/sessions", csrf, first)
	second := loginFlow(t, fx.h)
	if code, admin := getMe(t, fx.h, second); code != http.StatusOK || admin.Role != "admin" || admin.TwoFactorEnabled {
		t.Fatalf("o 2FA vem desligado: admin entra sem provar, veio %d %+v", code, admin)
	}
}

func TestAccountThatTurnedTwoFactorOnIsRefusedWith403AndNoSessionWithoutProof(t *testing.T) {
	fx := newFixture(t, true)
	first := loginFlow(t, fx.h)
	_, me := getMe(t, fx.h, first)
	fx.users.SetRole(uuidOf(t, me.ID), domain.RoleAdmin)
	_ = fx.users.SetMFARequired(context.Background(), uuidOf(t, me.ID), true)
	do(fx.h, http.MethodDelete, "/api/v1/me/sessions", csrf, first)

	res := do(fx.h, http.MethodGet, "/api/v1/auth/login", nil)
	flow := cookie(t, res, httpapi.FlowCookieSecure)
	state := res.Header.Get("Location")[strings.Index(res.Header.Get("Location"), "state=")+len("state="):]
	cb := do(fx.h, http.MethodGet, "/api/v1/auth/callback?code=code-ok&state="+state, nil, flow)
	if cb.StatusCode != http.StatusForbidden {
		t.Fatalf("admin sem 2 fatores: esperava 403, veio %d", cb.StatusCode)
	}
	for _, c := range cb.Cookies() {
		if c.Name == httpapi.SessionCookieSecure && c.MaxAge >= 0 {
			t.Fatal("nenhum cookie de sessao pode ser entregue quando o admin nao tem 2 fatores")
		}
	}

	fx.provider.Identity.MFA = true
	session := loginFlow(t, fx.h)
	if _, admin := getMe(t, fx.h, session); admin.Role != "admin" {
		t.Fatalf("admin com 2 fatores deveria entrar com papel admin, veio %q", admin.Role)
	}
}

func TestRevokeRequiresTheAntiCSRFHeader(t *testing.T) {
	fx := newFixture(t, true)
	session := loginFlow(t, fx.h)
	other := listSessions(t, fx.h, session)[0].ID
	for _, path := range []string{"/api/v1/me/sessions", "/api/v1/me/sessions/" + other} {
		if res := do(fx.h, http.MethodDelete, path, nil, session); res.StatusCode != http.StatusBadRequest {
			t.Errorf("DELETE %s sem o cabecalho: esperava 400, veio %d", path, res.StatusCode)
		}
	}
	if code, _ := getMe(t, fx.h, session); code != http.StatusOK {
		t.Fatal("a sessao nao pode ser derrubada por pedido sem o cabecalho anti-CSRF")
	}
}

func TestCannotRevokeSomeoneElsesSession(t *testing.T) {
	fx := newFixture(t, true)
	alice := loginAs(t, fx, "alice")
	bob := loginAs(t, fx, "bob")
	aliceSession := listSessions(t, fx.h, alice)[0].ID

	res := do(fx.h, http.MethodDelete, "/api/v1/me/sessions/"+aliceSession, csrf, bob)
	if res.StatusCode != http.StatusNotFound {
		t.Fatalf("revogar sessao alheia: esperava 404 (igual a inexistente), veio %d", res.StatusCode)
	}
	if code, _ := getMe(t, fx.h, alice); code != http.StatusOK {
		t.Fatal("a sessao da Alice deve continuar valida")
	}
	if got := listSessions(t, fx.h, bob); len(got) != 1 {
		t.Fatalf("a lista do Bob so pode ter a sessao dele, veio %d", len(got))
	}
}

func TestRevokeOneAndRevokeAllEndSessionsAndClearTheCookie(t *testing.T) {
	fx := newFixture(t, true)
	phone := loginAs(t, fx, "alice")
	laptop := loginAs(t, fx, "alice")
	var phoneID, laptopID string
	for _, s := range listSessions(t, fx.h, laptop) {
		if s.Current {
			laptopID = s.ID
		} else {
			phoneID = s.ID
		}
	}
	if res := do(fx.h, http.MethodDelete, "/api/v1/me/sessions/"+phoneID, csrf, laptop); res.StatusCode != http.StatusNoContent {
		t.Fatalf("derrubar o celular: %d", res.StatusCode)
	}
	if code, _ := getMe(t, fx.h, phone); code != http.StatusUnauthorized {
		t.Fatalf("o celular derrubado deve receber 401, veio %d", code)
	}
	self := do(fx.h, http.MethodDelete, "/api/v1/me/sessions/"+laptopID, csrf, laptop)
	if self.StatusCode != http.StatusNoContent || cookie(t, self, httpapi.SessionCookieSecure).MaxAge >= 0 {
		t.Fatalf("encerrar a propria sessao deve limpar o cookie: %d", self.StatusCode)
	}

	again := loginAs(t, fx, "alice")
	all := do(fx.h, http.MethodDelete, "/api/v1/me/sessions", csrf, again)
	if all.StatusCode != http.StatusNoContent || cookie(t, all, httpapi.SessionCookieSecure).MaxAge >= 0 {
		t.Fatalf("sair de todos os aparelhos: %d", all.StatusCode)
	}
	if code, _ := getMe(t, fx.h, again); code != http.StatusUnauthorized {
		t.Fatalf("depois de sair de todos, esperava 401, veio %d", code)
	}
}
