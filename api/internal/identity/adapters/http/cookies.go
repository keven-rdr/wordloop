package httpapi

import (
	"net/http"
	"time"
)

const (
	SessionCookie = "wl_session"
	FlowCookie    = "wl_oidc"

	SessionCookieSecure = "__Host-" + SessionCookie
	FlowCookieSecure    = "__Secure-" + FlowCookie

	ChallengeCookie       = "wl_pk"
	ChallengeCookieSecure = "__Secure-wl_pk"

	flowPath      = "/api/v1/auth"
	challengePath = "/api/v1/auth/passkeys"
	challengeTTL  = 5 * time.Minute
	flowTTL       = 10 * time.Minute
)

type cookies struct{ secure bool }

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

func (c cookies) challengeName() string {
	if c.secure {
		return ChallengeCookieSecure
	}
	return ChallengeCookie
}

func (c cookies) challenge(id string) *http.Cookie {
	ck := c.base(c.challengeName(), id, challengePath)
	ck.MaxAge = int(challengeTTL.Seconds())
	return ck
}

func (c cookies) clearChallenge() *http.Cookie {
	ck := c.base(c.challengeName(), "", challengePath)
	ck.MaxAge = -1
	return ck
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
