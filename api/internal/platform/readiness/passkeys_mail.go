package readiness

import (
	"net/mail"
	"net/url"
	"strings"
)

func (e *env) passkeys() {
	rpID, origin := e.val("WEBAUTHN_RP_ID"), e.val("WEBAUTHN_ORIGIN")
	switch {
	case rpID == "" && origin == "":
		if e.strict {
			e.add(LevelWarn, "passkey", "WEBAUTHN_RP_ID e WEBAUTHN_ORIGIN ausentes: login por passkey desligado")
		}
	case rpID == "" || origin == "":
		e.add(LevelError, "passkey", "defina WEBAUTHN_RP_ID e WEBAUTHN_ORIGIN juntos")
	case !e.passkeyProblem(rpID, origin):
		e.add(LevelOK, "passkey", "passkeys ligadas em %s", origin)
	}
}

// passkeyProblem registra o que estiver errado em RP ID e origem e devolve true se achou algo.
func (e *env) passkeyProblem(rpID, origin string) bool {
	u, err := url.Parse(origin)
	if err != nil || !plainOrigin(u) {
		e.add(LevelError, "passkey", "WEBAUTHN_ORIGIN invalida: use http(s)://host[:porta], sem caminho")
		return true
	}
	if u.Hostname() != rpID && !strings.HasSuffix(u.Hostname(), "."+rpID) {
		e.add(LevelError, "passkey", "WEBAUTHN_RP_ID (%s) precisa ser o host de WEBAUTHN_ORIGIN (%s) ou um dominio acima dele", rpID, u.Hostname())
		return true
	}
	if e.public != nil && u.Scheme+"://"+u.Host != e.public.Scheme+"://"+e.public.Host {
		e.add(LevelError, "passkey", "WEBAUTHN_ORIGIN (%s) difere de APP_PUBLIC_URL (%s): a passkey so vale na origem exata", origin, e.public)
		return true
	}
	return false
}

func plainOrigin(u *url.URL) bool {
	return (u.Scheme == "http" || u.Scheme == "https") && u.Host != "" && u.Path == ""
}

func (e *env) mail() {
	if e.val("MAIL_DRIVER") == "log" {
		if e.strict {
			e.add(LevelError, "e-mail", "MAIL_DRIVER=log escreve os links no log: so vale em desenvolvimento")
			return
		}
		e.add(LevelWarn, "e-mail", "MAIL_DRIVER=log: os e-mails (com os links) vao para o log")
		return
	}
	if e.val("BREVO_API_KEY") == "" {
		e.add(LevelWarn, "e-mail", "BREVO_API_KEY ausente: cadastro e recuperacao de senha por e-mail respondem 503")
		return
	}
	from, err := mail.ParseAddress(e.val("MAIL_FROM"))
	if err != nil {
		e.add(LevelError, "e-mail", "com BREVO_API_KEY, defina MAIL_FROM com um endereco valido (remetente verificado no Brevo)")
		return
	}
	if e.public == nil {
		e.add(LevelError, "e-mail", "o e-mail precisa de APP_PUBLIC_URL para montar os links")
		return
	}
	e.add(LevelOK, "e-mail", "e-mail por Brevo ligado, remetente %s", from.Address)
}
