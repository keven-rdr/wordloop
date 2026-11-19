// Package readiness confere a configuracao de um ambiente (variaveis de ambiente) ANTES de subir a API: o que esta
// ligado, o que esta desligado e o que esta errado. Existe porque o erro mais comum de uma estreia e de configuracao
// (callback do Google que nao bate com o dominio, chave de sessao ausente) e ele so apareceria com a API em
// crash-loop. Nunca devolve o valor de um segredo: so os nomes das variaveis e o que e publico (URLs, remetente).
package readiness

import (
	"fmt"
	"io"
	"net/url"
	"strings"
)

type Level string

const (
	LevelOK    Level = "ok"
	LevelWarn  Level = "aviso"
	LevelError Level = "erro"
)

type Finding struct {
	Level   Level
	Area    string
	Message string
}

type Report struct{ Findings []Finding }

// Failed diz se ha algum erro (avisos nao reprovam).
func (r Report) Failed() bool {
	for _, f := range r.Findings {
		if f.Level == LevelError {
			return true
		}
	}
	return false
}

// Write imprime uma linha por achado: "erro   login      mensagem".
func (r Report) Write(w io.Writer) {
	for _, f := range r.Findings {
		_, _ = fmt.Fprintf(w, "%-5s  %-9s  %s\n", f.Level, f.Area, f.Message)
	}
}

// env e a leitura das variaveis mais os achados que se acumulam. "Estrito" e tst e prd: la o que falta ou destoa
// reprova; em dev vira so aviso.
type env struct {
	get      func(string) string
	name     string
	strict   bool
	public   *url.URL
	findings []Finding
}

func (e *env) val(key string) string { return strings.TrimSpace(e.get(key)) }

func (e *env) add(level Level, area, format string, args ...any) {
	e.findings = append(e.findings, Finding{Level: level, Area: area, Message: fmt.Sprintf(format, args...)})
}

// missing registra o que falta: erro nos ambientes reais, aviso em dev.
func (e *env) missing(area, format string, args ...any) {
	if e.strict {
		e.add(LevelError, area, format, args...)
		return
	}
	e.add(LevelWarn, area, format, args...)
}

// Check confere a configuracao lida por get (em producao, os.Getenv).
func Check(get func(string) string) Report {
	e := &env{get: get}
	e.name = e.val("APP_ENV")
	if e.name == "" {
		e.name = "dev"
	}
	e.strict = e.name == "tst" || e.name == "prd"
	e.database()
	e.publicURL()
	e.login()
	e.passkeys()
	e.mail()
	e.hardening()
	return Report{Findings: e.findings}
}

func (e *env) database() {
	dsn := e.val("DATABASE_URL")
	switch {
	case dsn == "":
		e.add(LevelError, "banco", "DATABASE_URL ausente")
	case strings.Contains(dsn, "sslmode=disable") && e.strict:
		e.add(LevelWarn, "banco", "DATABASE_URL com sslmode=disable: so aceitavel se o banco estiver na mesma rede privada")
	default:
		e.add(LevelOK, "banco", "DATABASE_URL definida")
	}
}

func (e *env) publicURL() {
	raw := strings.TrimRight(e.val("APP_PUBLIC_URL"), "/")
	if raw == "" {
		e.missing("endereco", "APP_PUBLIC_URL ausente: sem ela nao ha link de convite, confirmacao nem login social")
		return
	}
	u, err := url.Parse(raw)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" || u.Path != "" {
		e.add(LevelError, "endereco", "APP_PUBLIC_URL invalida: use https://host, sem caminho")
		return
	}
	e.public = u
	if e.strict && (u.Scheme != "https" || isLocal(u.Hostname())) {
		e.add(LevelError, "endereco", "APP_PUBLIC_URL de %s precisa ser https e publica (veio %s)", e.name, raw)
		return
	}
	e.add(LevelOK, "endereco", "app publico em %s", raw)
}

func isLocal(host string) bool {
	return host == "localhost" || host == "127.0.0.1" || host == "::1" || strings.HasSuffix(host, ".localhost")
}

func (e *env) hardening() {
	if !e.strict {
		return
	}
	if e.val("API_TRUST_PROXY") != "true" {
		e.add(LevelWarn, "protecao", "API_TRUST_PROXY nao e true: atras do Caddy, o limite de tentativas enxerga um unico IP (o do proxy)")
	}
	if e.name == "prd" && e.val("API_DOCS_ENABLED") != "false" {
		e.add(LevelWarn, "protecao", "API_DOCS_ENABLED nao e false: a documentacao interativa fica ligada na API em producao")
	}
}
