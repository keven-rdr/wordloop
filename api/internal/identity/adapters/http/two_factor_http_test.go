package httpapi_test

import (
	"net/http"
	"testing"

	"github.com/keven-rdr/wordloop/api/internal/identity/domain"
)

func putTwoFactor(fx fixture, body string, headers map[string]string, cs ...*http.Cookie) *http.Response {
	return sendJSON(fx.h, http.MethodPut, "/api/v1/me/two-factor", body, headers, cs...)
}

func TestTwoFactorNeedsASessionAndTheAntiCSRFHeader(t *testing.T) {
	fx := newFixture(t, true)
	if res := putTwoFactor(fx, `{"enabled":true}`, csrf); res.StatusCode != http.StatusUnauthorized {
		t.Fatalf("sem sessao: esperava 401, veio %d", res.StatusCode)
	}
	session := loginFlow(t, fx.h)
	if res := putTwoFactor(fx, `{"enabled":true}`, nil, session); res.StatusCode != http.StatusBadRequest {
		t.Fatalf("sem o cabecalho anti-CSRF: esperava 400, veio %d", res.StatusCode)
	}
	if res := putTwoFactor(fx, `{"enabled":"talvez"}`, csrf, session); res.StatusCode != http.StatusBadRequest {
		t.Fatalf("corpo invalido: esperava 400, veio %d", res.StatusCode)
	}
}

func TestOrdinaryUsersCannotTurnTwoFactorOn(t *testing.T) {
	fx := newFixture(t, true)
	session := loginFlow(t, fx.h)
	if res := putTwoFactor(fx, `{"enabled":true}`, csrf, session); res.StatusCode != http.StatusForbidden {
		t.Fatalf("aluno: esperava 403, veio %d", res.StatusCode)
	}
	if _, me := getMe(t, fx.h, session); me.TwoFactorEnabled {
		t.Fatal("o pedido recusado nao pode ter ligado nada")
	}
}

func TestAdminTurnsTwoFactorOnOnlyFromASessionThatProvedIt(t *testing.T) {
	fx := newFixture(t, true)
	weak := loginFlow(t, fx.h)
	_, me := getMe(t, fx.h, weak)
	fx.users.SetRole(uuidOf(t, me.ID), domain.RoleAdmin)
	if res := putTwoFactor(fx, `{"enabled":true}`, csrf, weak); res.StatusCode != http.StatusConflict {
		t.Fatalf("sessao sem 2 fatores: esperava 409, veio %d", res.StatusCode)
	}

	fx.provider.Identity.MFA = true
	strong := loginFlow(t, fx.h)
	if res := putTwoFactor(fx, `{"enabled":true}`, csrf, strong); res.StatusCode != http.StatusNoContent {
		t.Fatalf("admin com prova: esperava 204, veio %d", res.StatusCode)
	}
	if _, after := getMe(t, fx.h, strong); !after.TwoFactorEnabled {
		t.Fatal("GET /me deve mostrar two_factor_enabled=true")
	}
	if res := putTwoFactor(fx, `{"enabled":false}`, csrf, weak); res.StatusCode != http.StatusNoContent {
		t.Fatalf("desligar de qualquer sessao: esperava 204, veio %d", res.StatusCode)
	}
	if _, after := getMe(t, fx.h, strong); after.TwoFactorEnabled {
		t.Fatal("o 2FA deveria ter sido desligado")
	}
}
