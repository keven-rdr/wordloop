package httpapi_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	httpapi "github.com/keven-rdr/wordloop/api/internal/identity/adapters/http"
	"github.com/keven-rdr/wordloop/api/internal/identity/domain"
	"github.com/keven-rdr/wordloop/api/internal/platform/token"
)

const (
	pwEmail = "dono@exemplo.com"
	pwGood  = "uma senha bem comprida"
)

func sendJSON(h http.Handler, method, path, body string, headers map[string]string, cs ...*http.Cookie) *http.Response {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	for _, c := range cs {
		req.AddCookie(c)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec.Result()
}

func postJSON(h http.Handler, path, body string, headers map[string]string, cs ...*http.Cookie) *http.Response {
	return sendJSON(h, http.MethodPost, path, body, headers, cs...)
}

func passwordLogin(fx fixture, email, password string) *http.Response {
	return postJSON(fx.h, "/api/v1/auth/password/login", `{"email":"`+email+`","password":"`+password+`"}`, csrf)
}

func sessionCookie(res *http.Response) *http.Cookie {
	for _, c := range res.Cookies() {
		if c.Name == httpapi.SessionCookieSecure && c.MaxAge >= 0 {
			return c
		}
	}
	return nil
}

func accountWithPassword(t *testing.T, fx fixture, role domain.Role) uuid.UUID {
	t.Helper()
	user, err := fx.users.Upsert(context.Background(), domain.Identity{Issuer: "i", Subject: "dono", DisplayName: "Dono"})
	if err != nil {
		t.Fatal(err)
	}
	fx.users.SetRole(user.ID, role)
	fx.pw.Add(user.ID, pwEmail, "h:"+pwGood)
	return user.ID
}

func TestPasswordLoginOpensTheSessionCookieAndMeWorks(t *testing.T) {
	fx := newFixture(t, true)
	accountWithPassword(t, fx, domain.RoleAdmin)
	res := passwordLogin(fx, "Dono@Exemplo.com", pwGood)
	c := sessionCookie(res)
	if res.StatusCode != http.StatusNoContent || c == nil || !c.HttpOnly || !c.Secure || c.Name != "__Host-wl_session" {
		t.Fatalf("login: %d %+v", res.StatusCode, c)
	}
	if code, me := getMe(t, fx.h, c); code != http.StatusOK || me.Role != "admin" || me.DisplayName != "Dono" {
		t.Fatalf("/me depois do login: %d %+v", code, me)
	}
}

func TestPasswordLoginNeverRevealsWhichEmailsExist(t *testing.T) {
	fx := newFixture(t, true)
	accountWithPassword(t, fx, domain.RoleUser)
	wrong, unknown := passwordLogin(fx, pwEmail, "senha errada mesmo"), passwordLogin(fx, "ninguem@exemplo.com", pwGood)
	a, _ := io.ReadAll(wrong.Body)
	b, _ := io.ReadAll(unknown.Body)
	if wrong.StatusCode != http.StatusUnauthorized || unknown.StatusCode != http.StatusUnauthorized || string(a) != string(b) {
		t.Fatalf("recusas diferentes: %d %q x %d %q", wrong.StatusCode, a, unknown.StatusCode, b)
	}
	if sessionCookie(wrong) != nil || sessionCookie(unknown) != nil || len(wrong.Cookies()) != 0 {
		t.Fatal("recusa nao entrega cookie")
	}
}

func TestPasswordLoginNeedsTheAntiCSRFHeaderAndAWellFormedBody(t *testing.T) {
	fx := newFixture(t, true)
	accountWithPassword(t, fx, domain.RoleUser)
	body := `{"email":"` + pwEmail + `","password":"` + pwGood + `"}`
	if res := postJSON(fx.h, "/api/v1/auth/password/login", body, nil); res.StatusCode != http.StatusBadRequest {
		t.Fatalf("sem o cabecalho: esperava 400, veio %d", res.StatusCode)
	}
	for _, bad := range []string{``, `{`, `{"email":1,"password":2}`} {
		if res := postJSON(fx.h, "/api/v1/auth/password/login", bad, csrf); res.StatusCode != http.StatusBadRequest {
			t.Errorf("corpo %q: esperava 400, veio %d", bad, res.StatusCode)
		}
	}
}

func TestPasswordLoginWithAMissingFieldIsJustAnotherWrongLogin(t *testing.T) {
	fx := newFixture(t, true)
	accountWithPassword(t, fx, domain.RoleUser)
	if res := postJSON(fx.h, "/api/v1/auth/password/login", `{"email":"`+pwEmail+`"}`, csrf); res.StatusCode != http.StatusUnauthorized {
		t.Fatalf("sem senha: esperava 401, veio %d", res.StatusCode)
	}
}

func TestFiveWrongPasswordsLockTheAccountOverHTTP(t *testing.T) {
	fx := newFixture(t, true)
	accountWithPassword(t, fx, domain.RoleUser)
	for range 5 {
		if res := passwordLogin(fx, pwEmail, "errada errada errada"); res.StatusCode != http.StatusUnauthorized {
			t.Fatalf("esperava 401, veio %d", res.StatusCode)
		}
	}
	if res := passwordLogin(fx, pwEmail, pwGood); res.StatusCode != http.StatusUnauthorized || sessionCookie(res) != nil {
		t.Fatalf("conta bloqueada recusa ate a senha certa: %d", res.StatusCode)
	}
}

func TestAccountWithTwoFactorOnGets403AndNoCookieFromPasswordAlone(t *testing.T) {
	fx := newFixture(t, true)
	id := accountWithPassword(t, fx, domain.RoleAdmin)
	_ = fx.users.SetMFARequired(context.Background(), id, true)
	res := passwordLogin(fx, pwEmail, pwGood)
	if res.StatusCode != http.StatusForbidden || sessionCookie(res) != nil {
		t.Fatalf("esperava 403 sem cookie, veio %d", res.StatusCode)
	}
	if wrong := passwordLogin(fx, pwEmail, "senha errada mesmo"); wrong.StatusCode != http.StatusUnauthorized {
		t.Fatalf("senha errada continua 401 (nao revela o 2FA): %d", wrong.StatusCode)
	}
}

func setPassword(fx fixture, invite, password string) *http.Response {
	return postJSON(fx.h, "/api/v1/auth/password/set", `{"invite":"`+invite+`","password":"`+password+`"}`, csrf)
}

func pendingAdmin(t *testing.T, fx fixture) {
	t.Helper()
	user, _ := fx.users.Upsert(context.Background(), domain.Identity{Issuer: "i", Subject: "novo", DisplayName: "Novo"})
	fx.users.SetRole(user.ID, domain.RoleAdmin)
	fx.pw.AddInvite(token.Hash("convite-secreto"), user.ID, pwEmail, time.Now().Add(10*time.Minute))
}

func TestInviteSetsThePasswordOnceAndLogsIn(t *testing.T) {
	fx := newFixture(t, true)
	pendingAdmin(t, fx)
	res := setPassword(fx, "convite-secreto", pwGood)
	c := sessionCookie(res)
	if res.StatusCode != http.StatusNoContent || c == nil {
		t.Fatalf("esperava 204 com cookie, veio %d", res.StatusCode)
	}
	if code, me := getMe(t, fx.h, c); code != http.StatusOK || me.Role != "admin" {
		t.Fatalf("/me: %d %+v", code, me)
	}
	if again := setPassword(fx, "convite-secreto", pwGood); again.StatusCode != http.StatusUnauthorized {
		t.Fatalf("convite usado nao vale de novo: %d", again.StatusCode)
	}
	if res := passwordLogin(fx, pwEmail, pwGood); res.StatusCode != http.StatusNoContent {
		t.Fatalf("depois de definir, entra com e-mail e senha: %d", res.StatusCode)
	}
}

func TestInviteRefusesWeakPasswordsAndUnknownInvitesWithoutSpendingIt(t *testing.T) {
	fx := newFixture(t, true)
	pendingAdmin(t, fx)
	for _, weak := range []string{"curta", pwEmail, "aaaaaaaaaaaaaaaa"} {
		if res := setPassword(fx, "convite-secreto", weak); res.StatusCode != http.StatusUnprocessableEntity || sessionCookie(res) != nil {
			t.Errorf("senha %q: esperava 422, veio %d", weak, res.StatusCode)
		}
	}
	for _, inv := range []string{"inventado", "x"} {
		if res := setPassword(fx, inv, pwGood); res.StatusCode != http.StatusUnauthorized {
			t.Errorf("convite %q: esperava 401, veio %d", inv, res.StatusCode)
		}
	}
	if res := setPassword(fx, "convite-secreto", pwGood); res.StatusCode != http.StatusNoContent {
		t.Fatalf("o convite continua valendo depois de senhas fracas: %d", res.StatusCode)
	}
}

func TestSetPasswordNeedsTheAntiCSRFHeader(t *testing.T) {
	fx := newFixture(t, true)
	pendingAdmin(t, fx)
	body := `{"invite":"convite-secreto","password":"` + pwGood + `"}`
	if res := postJSON(fx.h, "/api/v1/auth/password/set", body, nil); res.StatusCode != http.StatusBadRequest {
		t.Fatalf("sem o cabecalho: esperava 400, veio %d", res.StatusCode)
	}
}
