package readiness_test

import (
	"bufio"
	"encoding/base64"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/keven-rdr/wordloop/api/internal/platform/readiness"
)

// Contrato: os arquivos de ambiente versionados em deploy/ (os que o CI manda para a VM) passam na conferencia. Se
// alguem mexer num deles e quebrar a coerencia (callback, passkey, remetente), o teste acusa antes do deploy.

func readEnvFile(t *testing.T, rel string, replace map[string]string) map[string]string {
	t.Helper()
	f, err := os.Open(filepath.Join("..", "..", "..", "..", "deploy", rel))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = f.Close() }()
	out := map[string]string{}
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		for from, to := range replace {
			line = strings.ReplaceAll(line, from, to)
		}
		k, v, _ := strings.Cut(line, "=")
		out[k] = v
	}
	if err := sc.Err(); err != nil {
		t.Fatal(err)
	}
	return out
}

// secrets: o que o CI acrescenta a partir dos secrets do GitHub (valores de mentira).
func secrets(m map[string]string) {
	m["DATABASE_URL"] = canaryDB
	m["GOOGLE_CLIENT_ID"], m["GOOGLE_CLIENT_SECRET"] = "canary-client-id", canaryOAuth
	m["SESSION_KEY"] = base64.StdEncoding.EncodeToString(make([]byte, 32))
	m["BREVO_API_KEY"] = canaryBrevo
}

func TestTheTstEnvironmentFileIsCoherent(t *testing.T) {
	m := readEnvFile(t, "tst/api.env", nil)
	m["APP_ENV"] = "tst" // vem do compose
	secrets(m)
	r := readiness.Check(lookup(m))
	if r.Failed() {
		t.Fatalf("deploy/tst/api.env reprova: %+v", r.Findings)
	}
}

func TestTheProductionTemplateIsCoherentOnceFilled(t *testing.T) {
	m := readEnvFile(t, "prd/api.env.tmpl", map[string]string{"{{DOMAIN}}": "app.exemplo.com.br", "{{MAIL_FROM}}": "nao-responda@exemplo.com.br"})
	m["APP_ENV"] = "prd" // vem do compose
	secrets(m)
	r := readiness.Check(lookup(m))
	for _, f := range r.Findings {
		if f.Level != readiness.LevelOK {
			t.Errorf("deploy/prd/api.env.tmpl deixa um achado: %+v", f)
		}
	}
}
