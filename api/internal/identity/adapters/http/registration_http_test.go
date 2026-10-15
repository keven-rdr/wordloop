package httpapi_test

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/keven-rdr/wordloop/api/internal/api"
	httpapi "github.com/keven-rdr/wordloop/api/internal/identity/adapters/http"
	"github.com/keven-rdr/wordloop/api/internal/identity/domain"
)

func registerEmail(fx fixture, path, email string, headers map[string]string) *http.Response {
	return postJSON(fx.h, path, `{"email":"`+email+`","locale":"pt-BR"}`, headers)
}

func TestRegisterAndForgotAnswerTheSameWhetherOrNotTheEmailExists(t *testing.T) {
	fx := newFixture(t, true)
	fx.reg.Seed("aluna@exemplo.com", domain.RoleUser, true)
	fx.reg.Seed("dono@exemplo.com", domain.RoleAdmin, true)
	for _, path := range []string{"/api/v1/auth/register", "/api/v1/auth/password/forgot"} {
		var first string
		for i, addr := range []string{"novo@exemplo.com", "aluna@exemplo.com", "dono@exemplo.com", "ninguem@exemplo.com"} {
			res := registerEmail(fx, path, addr, csrf)
			body, _ := io.ReadAll(res.Body)
			if res.StatusCode != http.StatusAccepted || len(res.Cookies()) != 0 || res.Header.Get("Cache-Control") != "no-store" {
				t.Fatalf("%s %s: %d cookies=%v", path, addr, res.StatusCode, res.Cookies())
			}
			if i == 0 {
				first = string(body)
			} else if string(body) != first {
				t.Fatalf("%s: o corpo muda conforme o e-mail existe: %q x %q", path, body, first)
			}
		}
	}
}

func TestRegisterSendsTheLinkOnlyToNewAddressesAndNeverReturnsIt(t *testing.T) {
	fx := newFixture(t, true)
	res := registerEmail(fx, "/api/v1/auth/register", "novo@exemplo.com", csrf)
	body, _ := io.ReadAll(res.Body)
	if fx.mailbox.Count() != 1 || fx.mailbox.LastSecret() == "" || strings.Contains(string(body), fx.mailbox.LastSecret()) {
		t.Fatalf("e-mails=%d corpo=%q", fx.mailbox.Count(), body)
	}
	if !strings.Contains(fx.mailbox.Last().Text, "https://app.test/verify#") {
		t.Fatalf("link: %q", fx.mailbox.Last().Text)
	}
}

func TestRegistrationRoutesNeedTheHeaderAndAWellFormedEmail(t *testing.T) {
	fx := newFixture(t, true)
	for _, path := range []string{"/api/v1/auth/register", "/api/v1/auth/password/forgot"} {
		if res := registerEmail(fx, path, "novo@exemplo.com", nil); res.StatusCode != http.StatusBadRequest {
			t.Errorf("%s sem o cabecalho: %d", path, res.StatusCode)
		}
		for _, bad := range []string{"", "sem-arroba", "a@b"} {
			if res := registerEmail(fx, path, bad, csrf); res.StatusCode != http.StatusBadRequest {
				t.Errorf("%s com %q: %d", path, bad, res.StatusCode)
			}
		}
		for _, body := range []string{``, `{`, `{"email":1}`} {
			if res := postJSON(fx.h, path, body, csrf); res.StatusCode != http.StatusBadRequest {
				t.Errorf("%s com o corpo %q: %d", path, body, res.StatusCode)
			}
		}
	}
	if fx.mailbox.Count() != 0 || fx.reg.Accounts() != 0 {
		t.Fatal("pedidos invalidos nao criam nem enviam nada")
	}
}

func TestAnUnknownLanguageFallsBackToPortuguese(t *testing.T) {
	fx := newFixture(t, true)
	res := postJSON(fx.h, "/api/v1/auth/register", `{"email":"novo@exemplo.com","locale":"fr-FR"}`, csrf)
	if res.StatusCode != http.StatusAccepted || !strings.Contains(fx.mailbox.Last().Subject, "Confirme o seu e-mail") {
		t.Fatalf("%d %q", res.StatusCode, fx.mailbox.Last().Subject)
	}
}

func TestRegistrationIsUnavailableWithoutAMailService(t *testing.T) {
	bare := api.Handler(api.Server{Identity: httpapi.Handler{Secure: true}})
	none := api.Handler(api.Server{})
	for _, h := range []http.Handler{bare, none} {
		for _, path := range []string{"/api/v1/auth/register", "/api/v1/auth/password/forgot"} {
			if res := postJSON(h, path, `{"email":"novo@exemplo.com"}`, csrf); res.StatusCode != http.StatusServiceUnavailable {
				t.Errorf("%s: esperava 503, veio %d", path, res.StatusCode)
			}
		}
	}
}
