//go:build integration

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/cookiejar"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/keven-rdr/wordloop/api/internal/api"
	"github.com/keven-rdr/wordloop/api/internal/identity/accounts"
	"github.com/keven-rdr/wordloop/api/internal/identity/identitytest"
	"github.com/keven-rdr/wordloop/api/internal/platform/db/dbtest"
)

type browserSession struct {
	t      *testing.T
	base   string
	client *http.Client
}

func newBrowserSession(t *testing.T, base string) *browserSession {
	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatal(err)
	}
	return &browserSession{t: t, base: base, client: &http.Client{Jar: jar, Timeout: 10 * time.Second}}
}

func (b *browserSession) call(method, path string, body any) (int, []byte) {
	b.t.Helper()
	var rd io.Reader
	if body != nil {
		raw, _ := json.Marshal(body)
		rd = bytes.NewReader(raw)
	}
	req, err := http.NewRequest(method, b.base+path, rd)
	if err != nil {
		b.t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	res, err := b.client.Do(req)
	if err != nil {
		b.t.Fatal(err)
	}
	defer res.Body.Close()
	raw, _ := io.ReadAll(res.Body)
	return res.StatusCode, raw
}

func optionsOf(t *testing.T, raw []byte) []byte {
	t.Helper()
	var out struct {
		Options json.RawMessage `json:"options"`
	}
	if err := json.Unmarshal(raw, &out); err != nil || len(out.Options) == 0 {
		t.Fatalf("opcoes ausentes: %s (%v)", raw, err)
	}
	return out.Options
}

func TestPasskeyFlowEndToEndOnRealPostgres(t *testing.T) {
	url, _ := dbtest.Postgres(t)
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	origin := fmt.Sprintf("http://localhost:%d", ln.Addr().(*net.TCPAddr).Port)
	t.Setenv("DATABASE_URL", url)
	t.Setenv("APP_PUBLIC_URL", origin)
	t.Setenv("WEBAUTHN_RP_ID", "localhost")
	t.Setenv("WEBAUTHN_ORIGIN", origin)
	t.Setenv("OIDC_ISSUER", "")
	ctx := context.Background()
	if err := execute(ctx, []string{"migrate"}); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	if err := createAdmin(ctx, []string{"--email", "voce@exemplo.com", "--name", "Voce"}, &out); err != nil {
		t.Fatal(err)
	}
	_, after, found := strings.Cut(out.String(), "/invite#")
	if !found {
		t.Fatalf("convite ausente na saida: %q", out.String())
	}
	invite := strings.TrimSpace(after)

	pool, err := openDatabase(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	ident, err := buildIdentity(pool)
	if err != nil {
		t.Fatal(err)
	}
	srv := &http.Server{Handler: api.Handler(api.Server{Identity: ident, DB: pool}), ReadHeaderTimeout: time.Second}
	go func() { _ = srv.Serve(ln) }()
	defer srv.Close()

	soft := identitytest.NewSoftAuthenticator(origin, "localhost")
	first := newBrowserSession(t, "http://"+ln.Addr().String())

	code, raw := first.call(http.MethodPost, "/api/v1/auth/passkeys/register/begin", map[string]string{"invite": invite})
	if code != http.StatusOK {
		t.Fatalf("register/begin: %d %s", code, raw)
	}
	cred, err := soft.Create(optionsOf(t, raw))
	if err != nil {
		t.Fatal(err)
	}
	code, raw = first.call(http.MethodPost, "/api/v1/auth/passkeys/register/finish", map[string]any{"credential": json.RawMessage(cred), "name": "Meu celular"})
	if code != http.StatusCreated {
		t.Fatalf("register/finish: %d %s", code, raw)
	}
	code, raw = first.call(http.MethodGet, "/api/v1/me", nil)
	var me struct {
		ID   uuid.UUID `json:"id"`
		Role string    `json:"role"`
	}
	if err := json.Unmarshal(raw, &me); err != nil || code != http.StatusOK || me.Role != "admin" {
		t.Fatalf("/me depois do convite: %d %s", code, raw)
	}

	if code, _ := newBrowserSession(t, first.base).call(http.MethodPost, "/api/v1/auth/passkeys/register/begin", map[string]string{"invite": invite}); code != http.StatusUnauthorized {
		t.Fatalf("o convite usado nao vale de novo: %d", code)
	}
	if err := createAdmin(ctx, []string{"--email", "voce@exemplo.com"}, &bytes.Buffer{}); !errors.Is(err, accounts.ErrAccountExists) {
		t.Fatalf("create-admin nao pode reemitir convite para quem ja tem passkey: %v", err)
	}

	second := newBrowserSession(t, first.base)
	code, raw = second.call(http.MethodPost, "/api/v1/auth/passkeys/login/begin", nil)
	if code != http.StatusOK {
		t.Fatalf("login/begin: %d %s", code, raw)
	}
	resp, err := soft.Get(optionsOf(t, raw), me.ID[:])
	if err != nil {
		t.Fatal(err)
	}
	if code, raw := second.call(http.MethodPost, "/api/v1/auth/passkeys/login/finish", map[string]any{"credential": json.RawMessage(resp)}); code != http.StatusNoContent {
		t.Fatalf("login/finish: %d %s", code, raw)
	}
	if code, _ := second.call(http.MethodGet, "/api/v1/me", nil); code != http.StatusOK {
		t.Fatalf("/me depois do login por passkey: %d", code)
	}

	code, raw = second.call(http.MethodGet, "/api/v1/me/passkeys", nil)
	var list []struct {
		ID uuid.UUID `json:"id"`
	}
	if err := json.Unmarshal(raw, &list); err != nil || code != http.StatusOK || len(list) != 1 {
		t.Fatalf("lista de passkeys: %d %s", code, raw)
	}
	if code, _ := second.call(http.MethodDelete, "/api/v1/me/passkeys/"+list[0].ID.String(), nil); code != http.StatusConflict {
		t.Fatalf("o admin nao pode apagar a unica passkey: %d", code)
	}
}
