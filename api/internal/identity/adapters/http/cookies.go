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

	flowPath = "/api/v1/auth"
	flowTTL  = 10 * time.Minute
)

// cookies monta os cookies com os mesmos atributos de seguranca; Secure vale quando a URL publica e https.
type cookies struct{ secure bool }

func (c cookies) base(name, value, path string) *http.Cookie {
	return &http.Cookie{Name: name, Value: value, Path: path, HttpOnly: true, Secure: c.secure, SameSite: http.SameSiteLaxMode}
}

func (c cookies) session(id string, expires time.Time) *http.Cookie {
	ck := c.base(SessionCookie, id, "/")
	ck.Expires = expires
	ck.MaxAge = int(time.Until(expires).Seconds())
	return ck
}

func (c cookies) flow(sealed string) *http.Cookie {
	ck := c.base(FlowCookie, sealed, flowPath)
	ck.MaxAge = int(flowTTL.Seconds())
	return ck
}

func (c cookies) clearSession() *http.Cookie {
	ck := c.base(SessionCookie, "", "/")
	ck.MaxAge = -1
	return ck
}

func (c cookies) clearFlow() *http.Cookie {
	ck := c.base(FlowCookie, "", flowPath)
	ck.MaxAge = -1
	return ck
}
