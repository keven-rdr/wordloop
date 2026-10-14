// Package httpapi liga os casos de uso do modulo identity a interface HTTP gerada do contrato (cookies e redirecionamentos).
package httpapi

import (
	"net/http"
	"time"
)

const (
	// SessionCookie guarda o id da sessao do BFF (httpOnly: o JavaScript nunca o ve).
	SessionCookie = "wl_session"
	// FlowCookie guarda o estado do login em andamento, selado, por poucos minutos.
	FlowCookie = "wl_oidc"

	// Com HTTPS os nomes ganham prefixo: "__Host-" obriga Secure, Path=/ e proibe Domain (nenhum subdominio o sobrescreve);
	// o cookie de fluxo usa "__Secure-" porque seu Path e /api/v1/auth (o "__Host-" exige Path=/).
	SessionCookieSecure = "__Host-" + SessionCookie
	FlowCookieSecure    = "__Secure-" + FlowCookie

	flowPath = "/api/v1/auth"
	flowTTL  = 10 * time.Minute
)

// cookies monta os cookies com os mesmos atributos de seguranca; Secure vale quando a URL publica e https.
type cookies struct{ secure bool }

// sessionName e flowName escolhem o nome com ou sem prefixo; sem HTTPS (dev local) o navegador recusaria o prefixo.
func (c cookies) sessionName() string {
	if c.secure {
		return SessionCookieSecure
	}
	return SessionCookie
}

func (c cookies) flowName() string {
	if c.secure {
		return FlowCookieSecure
	}
	return FlowCookie
}

func (c cookies) base(name, value, path string) *http.Cookie {
	return &http.Cookie{Name: name, Value: value, Path: path, HttpOnly: true, Secure: c.secure, SameSite: http.SameSiteLaxMode}
}

func (c cookies) session(id string, expires time.Time) *http.Cookie {
	ck := c.base(c.sessionName(), id, "/")
	ck.Expires = expires
	ck.MaxAge = int(time.Until(expires).Seconds())
	return ck
}

func (c cookies) flow(sealed string) *http.Cookie {
	ck := c.base(c.flowName(), sealed, flowPath)
	ck.MaxAge = int(flowTTL.Seconds())
	return ck
}

func (c cookies) clearSession() *http.Cookie {
	ck := c.base(c.sessionName(), "", "/")
	ck.MaxAge = -1
	return ck
}

func (c cookies) clearFlow() *http.Cookie {
	ck := c.base(c.flowName(), "", flowPath)
	ck.MaxAge = -1
	return ck
}
