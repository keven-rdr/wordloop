package readiness_test

import (
	"bytes"
	"encoding/base64"
	"strings"
	"testing"

	"github.com/keven-rdr/wordloop/api/internal/platform/readiness"
)

const (
	canaryDB     = "postgres://canary-user:canary-db-value@db.example/app?sslmode=require"
	canaryOAuth  = "canary-oauth-value"
	canaryBrevo  = "canary-mail-value"
	publicDomain = "https://app.exemplo.com.br"
)

func lookup(m map[string]string) func(string) string { return func(k string) string { return m[k] } }

// prd devolve uma configuracao de producao completa e correta; cada teste estraga uma coisa.
func prd() map[string]string {
	return map[string]string{
		"APP_ENV": "prd", "DATABASE_URL": canaryDB, "APP_PUBLIC_URL": publicDomain,
		"OIDC_REDIRECT_URL": publicDomain + "/api/v1/auth/callback",
		"GOOGLE_CLIENT_ID":  "canary-client-id", "GOOGLE_CLIENT_SECRET": canaryOAuth,
		"SESSION_KEY":    base64.StdEncoding.EncodeToString(make([]byte, 32)),
		"WEBAUTHN_RP_ID": "app.exemplo.com.br", "WEBAUTHN_ORIGIN": publicDomain,
		"BREVO_API_KEY": canaryBrevo, "MAIL_FROM": "nao-responda@exemplo.com.br",
		"API_TRUST_PROXY": "true", "API_DOCS_ENABLED": "false",
	}
}

func has(r readiness.Report, level readiness.Level, part string) bool {
	for _, f := range r.Findings {
		if f.Level == level && strings.Contains(f.Message, part) {
			return true
		}
	}
	return false
}

func TestAFullProductionConfigurationHasNoErrorsNorWarnings(t *testing.T) {
	r := readiness.Check(lookup(prd()))
	for _, f := range r.Findings {
		if f.Level != readiness.LevelOK {
			t.Errorf("achado inesperado: %+v", f)
		}
	}
	if r.Failed() {
		t.Fatal("configuracao correta nao pode reprovar")
	}
	if !has(r, readiness.LevelOK, "https://app.exemplo.com.br/api/v1/auth/callback") {
		t.Error("o relatorio deve dizer qual callback registrar nos provedores")
	}
}

func TestEachMistakeInProductionIsAnError(t *testing.T) {
	cases := []struct {
		name   string
		mutate func(map[string]string)
		want   string
	}{
		{"sem banco", func(m map[string]string) { delete(m, "DATABASE_URL") }, "DATABASE_URL ausente"},
		{"sem endereco publico", func(m map[string]string) { delete(m, "APP_PUBLIC_URL") }, "APP_PUBLIC_URL ausente"},
		{"endereco em http", func(m map[string]string) { m["APP_PUBLIC_URL"] = "http://app.exemplo.com.br" }, "precisa ser https"},
		{"endereco local", func(m map[string]string) { m["APP_PUBLIC_URL"] = "https://localhost" }, "precisa ser https"},
		{"endereco com caminho", func(m map[string]string) { m["APP_PUBLIC_URL"] = publicDomain + "/app" }, "invalida"},
		{"google pela metade", func(m map[string]string) { delete(m, "GOOGLE_CLIENT_SECRET") }, "google pela metade"},
		{"sem chave de sessao", func(m map[string]string) { delete(m, "SESSION_KEY") }, "SESSION_KEY"},
		{"chave de sessao curta", func(m map[string]string) { m["SESSION_KEY"] = "curta" }, "SESSION_KEY invalida"},
		{"callback em outro dominio", func(m map[string]string) { m["OIDC_REDIRECT_URL"] = "https://outro.com/api/v1/auth/callback" }, "dominios diferentes"},
		{"callback com caminho errado", func(m map[string]string) { m["OIDC_REDIRECT_URL"] = publicDomain + "/callback" }, "deve terminar em"},
		{"callback em http", func(m map[string]string) { m["OIDC_REDIRECT_URL"] = "http://app.exemplo.com.br/api/v1/auth/callback" }, "precisa ser https"},
		{"passkey pela metade", func(m map[string]string) { delete(m, "WEBAUTHN_RP_ID") }, "juntos"},
		{"passkey em outra origem", func(m map[string]string) { m["WEBAUTHN_ORIGIN"] = "https://app.exemplo.com.br:8443" }, "difere de APP_PUBLIC_URL"},
		{"passkey com rp errado", func(m map[string]string) { m["WEBAUTHN_RP_ID"] = "outro.com" }, "precisa ser o host"},
		{"passkey com origem torta", func(m map[string]string) { m["WEBAUTHN_ORIGIN"] = "https://app.exemplo.com.br/x" }, "invalida"},
		{"e-mail no log", func(m map[string]string) { m["MAIL_DRIVER"] = "log" }, "so vale em desenvolvimento"},
		{"brevo sem remetente", func(m map[string]string) { delete(m, "MAIL_FROM") }, "MAIL_FROM"},
		{"brevo com remetente torto", func(m map[string]string) { m["MAIL_FROM"] = "isso-nao-e-email" }, "MAIL_FROM"},
		{"brevo sem endereco publico", func(m map[string]string) { delete(m, "APP_PUBLIC_URL"); delete(m, "OIDC_REDIRECT_URL") }, "APP_PUBLIC_URL"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			m := prd()
			c.mutate(m)
			r := readiness.Check(lookup(m))
			if !r.Failed() || !has(r, readiness.LevelError, c.want) {
				t.Fatalf("esperava erro com %q, veio %+v", c.want, r.Findings)
			}
		})
	}
}

func TestOnlyWarningsDoNotFail(t *testing.T) {
	m := prd()
	m["API_TRUST_PROXY"], m["API_DOCS_ENABLED"] = "", ""
	for _, k := range []string{"GOOGLE_CLIENT_ID", "GOOGLE_CLIENT_SECRET", "SESSION_KEY", "WEBAUTHN_RP_ID", "WEBAUTHN_ORIGIN", "BREVO_API_KEY", "OIDC_REDIRECT_URL"} {
		delete(m, k)
	}
	m["DATABASE_URL"] = canaryDB + "&sslmode=disable"
	r := readiness.Check(lookup(m))
	if r.Failed() {
		t.Fatalf("so avisos: nao pode reprovar: %+v", r.Findings)
	}
	for _, w := range []string{"API_TRUST_PROXY", "API_DOCS_ENABLED", "nenhum provedor social", "passkey desligado", "BREVO_API_KEY ausente", "sslmode=disable"} {
		if !has(r, readiness.LevelWarn, w) {
			t.Errorf("faltou o aviso sobre %q: %+v", w, r.Findings)
		}
	}
}

func TestDevelopmentIsForgiving(t *testing.T) {
	r := readiness.Check(lookup(map[string]string{"DATABASE_URL": canaryDB}))
	if r.Failed() {
		t.Fatalf("dev so com banco deve passar: %+v", r.Findings)
	}
	dev := readiness.Check(lookup(map[string]string{"DATABASE_URL": canaryDB, "APP_PUBLIC_URL": "http://localhost:5173", "MAIL_DRIVER": "log"}))
	if dev.Failed() || !has(dev, readiness.LevelWarn, "MAIL_DRIVER=log") {
		t.Fatalf("dev com http e e-mail no log: so aviso: %+v", dev.Findings)
	}
}

func TestTheCallbackIsDerivedFromThePublicAddressWhenNotSet(t *testing.T) {
	m := prd()
	delete(m, "OIDC_REDIRECT_URL")
	r := readiness.Check(lookup(m))
	if r.Failed() || !has(r, readiness.LevelOK, publicDomain+"/api/v1/auth/callback") {
		t.Fatalf("callback deveria vir de APP_PUBLIC_URL: %+v", r.Findings)
	}
}

func TestTheReportNeverPrintsASecret(t *testing.T) {
	m := prd()
	m["SESSION_KEY"] = "chave-invalida-canary"
	m["MAIL_FROM"] = "invalido"
	var out bytes.Buffer
	readiness.Check(lookup(m)).Write(&out)
	for _, secret := range []string{canaryDB, "canary-db-value", canaryOAuth, canaryBrevo, "chave-invalida-canary"} {
		if strings.Contains(out.String(), secret) {
			t.Fatalf("o relatorio vazou %q:\n%s", secret, out.String())
		}
	}
	if !strings.Contains(out.String(), "erro") {
		t.Fatalf("o relatorio deveria listar os erros:\n%s", out.String())
	}
}
