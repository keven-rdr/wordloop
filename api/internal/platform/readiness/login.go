package readiness

import (
	"net/url"
	"strings"

	"github.com/keven-rdr/wordloop/api/internal/platform/seal"
)

const callbackPath = "/api/v1/auth/callback"

type provider struct {
	name string
	keys []string
}

var providers = []provider{
	{"google", []string{"GOOGLE_CLIENT_ID", "GOOGLE_CLIENT_SECRET"}},
	{"github", []string{"GITHUB_CLIENT_ID", "GITHUB_CLIENT_SECRET"}},
	{"oidc", []string{"OIDC_ISSUER", "OIDC_CLIENT_ID", "OIDC_CLIENT_SECRET"}},
}

// activeProviders devolve os provedores completos e registra os pela metade (tudo ou nada).
func (e *env) activeProviders() []string {
	var on []string
	for _, p := range providers {
		set := 0
		for _, k := range p.keys {
			if e.val(k) != "" {
				set++
			}
		}
		switch {
		case set == len(p.keys):
			on = append(on, p.name)
		case set > 0:
			e.add(LevelError, "login", "%s pela metade: defina todas ou nenhuma de %s", p.name, strings.Join(p.keys, ", "))
		}
	}
	return on
}

func (e *env) login() {
	on := e.activeProviders()
	if len(on) == 0 {
		if e.strict {
			e.add(LevelWarn, "login", "nenhum provedor social ligado: so e-mail, senha e passkey")
		}
		return
	}
	e.sessionKey()
	if callback := e.callback(); callback != "" {
		e.add(LevelOK, "login", "login social ligado (%s); registre este callback em cada provedor: %s", strings.Join(on, ", "), callback)
	}
}

func (e *env) sessionKey() {
	key := e.val("SESSION_KEY")
	if key == "" {
		e.add(LevelError, "login", "com login social, defina SESSION_KEY (openssl rand -base64 32): sela o fluxo de entrada")
		return
	}
	if _, err := seal.ParseKey(key); err != nil {
		e.add(LevelError, "login", "SESSION_KEY invalida: esperava 32 bytes em base64 (openssl rand -base64 32)")
	}
}

// callback devolve a URL de retorno dos provedores ("" se nao ha como montar ou se esta errada).
func (e *env) callback() string {
	raw := e.val("OIDC_REDIRECT_URL")
	if raw == "" && e.public != nil {
		return e.public.String() + callbackPath
	}
	if raw == "" {
		e.add(LevelError, "login", "defina OIDC_REDIRECT_URL (ou APP_PUBLIC_URL): e para la que o provedor devolve o navegador")
		return ""
	}
	u, err := url.Parse(raw)
	if err != nil || u.Host == "" {
		e.add(LevelError, "login", "OIDC_REDIRECT_URL invalida: %s", raw)
		return ""
	}
	if !e.redirectOK(u) {
		return ""
	}
	return raw
}

func (e *env) redirectOK(u *url.URL) bool {
	ok := true
	if u.Path != callbackPath {
		e.add(LevelError, "login", "OIDC_REDIRECT_URL deve terminar em %s (veio %s)", callbackPath, u.Path)
		ok = false
	}
	if e.strict && u.Scheme != "https" {
		e.add(LevelError, "login", "OIDC_REDIRECT_URL de %s precisa ser https: o cookie do fluxo e Secure", e.name)
		ok = false
	}
	if e.public != nil && u.Host != e.public.Host {
		e.add(LevelError, "login", "OIDC_REDIRECT_URL (%s) e APP_PUBLIC_URL (%s) em dominios diferentes: o cookie nao volta", u.Host, e.public.Host)
		ok = false
	}
	return ok
}
