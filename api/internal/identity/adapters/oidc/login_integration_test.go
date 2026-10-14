//go:build integration

package oidc_test

import (
	"bytes"
	"context"
	"encoding/json"
	"html"
	"io"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"
	"testing"

	"github.com/keven-rdr/wordloop/api/internal/api"
	httpapi "github.com/keven-rdr/wordloop/api/internal/identity/adapters/http"
	"github.com/keven-rdr/wordloop/api/internal/identity/adapters/oidc"
	"github.com/keven-rdr/wordloop/api/internal/identity/adapters/postgres"
	"github.com/keven-rdr/wordloop/api/internal/identity/app"
	"github.com/keven-rdr/wordloop/api/internal/identity/identitytest"
	"github.com/keven-rdr/wordloop/api/internal/platform/db"
	"github.com/keven-rdr/wordloop/api/internal/platform/db/dbtest"
	"github.com/keven-rdr/wordloop/api/internal/platform/seal"
	"github.com/keven-rdr/wordloop/api/migrations"
)

var formAction = regexp.MustCompile(`action="([^"]+)"`)

// startAPI sobe Keycloak e Postgres REAIS e a API ligada a eles; devolve a URL base e um atalho para consultar o banco.
func startAPI(t *testing.T) (base string, query func(sql string, dest ...any)) {
	t.Helper()
	ctx := context.Background()
	kc := identitytest.StartKeycloak(t)
	dbURL, _ := dbtest.Postgres(t)
	pool, err := db.Open(ctx, dbURL)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	if err := db.Migrate(ctx, pool, migrations.FS); err != nil {
		t.Fatal(err)
	}

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = ln.Close() })
	base = "http://" + ln.Addr().String()
	redirect := base + "/api/v1/auth/callback"
	kc.AllowRedirect(redirect)

	sealer, err := seal.New(bytes.Repeat([]byte{4}, seal.KeySize))
	if err != nil {
		t.Fatal(err)
	}
	provider := oidc.New(oidc.Config{Issuer: kc.Issuer(), ClientID: identitytest.ClientID, ClientSecret: identitytest.ClientSecret, RedirectURL: redirect})
	users, sessions := postgres.NewUsers(pool), postgres.NewSessions(pool)
	auth := app.Authenticate{Sessions: sessions, Users: users}
	h := httpapi.Handler{
		Start:  &app.StartLogin{Provider: provider, Sealer: sealer},
		Finish: &app.FinishLogin{Provider: provider, Users: users, Sessions: sessions, Sealer: sealer},
		Auth:   auth, Out: app.Logout{Sessions: sessions},
		List:   app.ListSessions{Auth: auth, Sessions: sessions},
		Revoke: app.RevokeSession{Auth: auth, Sessions: sessions}, RevokeAll: app.RevokeAllSessions{Auth: auth, Sessions: sessions},
	}
	go func() { _ = http.Serve(ln, api.Handler(api.Server{Identity: h, DB: pool})) }()
	return base, func(sql string, dest ...any) {
		t.Helper()
		if err := pool.QueryRow(ctx, sql).Scan(dest...); err != nil {
			t.Fatalf("%s: %v", sql, err)
		}
	}
}

// browser imita o navegador: guarda cookies e segue redirecionamentos a mao.
type browser struct {
	t      *testing.T
	client *http.Client
}

func newBrowser(t *testing.T) *browser {
	jar, _ := cookiejar.New(nil)
	return &browser{t: t, client: &http.Client{Jar: jar, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}}
}

func (b *browser) do(method, target string, form url.Values, headers map[string]string) *http.Response {
	b.t.Helper()
	var body io.Reader
	if form != nil {
		body = bytes.NewBufferString(form.Encode())
	}
	req, err := http.NewRequest(method, target, body)
	if err != nil {
		b.t.Fatal(err)
	}
	if form != nil {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	res, err := b.client.Do(req)
	if err != nil {
		b.t.Fatal(err)
	}
	return res
}

// followTo segue redirecionamentos ate a resposta do host da API (callback) e a devolve.
func (b *browser) followTo(res *http.Response, apiHost string) *http.Response {
	b.t.Helper()
	for i := 0; i < 10; i++ {
		loc, err := res.Location()
		if err != nil {
			b.t.Fatalf("esperava redirecionamento, veio %d", res.StatusCode)
		}
		_ = res.Body.Close()
		res = b.do(http.MethodGet, loc.String(), nil, nil)
		if loc.Host == apiHost {
			return res
		}
	}
	b.t.Fatal("redirecionamentos demais")
	return nil
}

func (b *browser) signIn(base, user, password string) *http.Response {
	b.t.Helper()
	host := mustHost(b.t, base)
	login := b.do(http.MethodGet, base+"/api/v1/auth/login?return_to=/estudo", nil, nil)
	if login.StatusCode != http.StatusFound {
		b.t.Fatalf("login: esperava 302, veio %d", login.StatusCode)
	}
	page := b.do(http.MethodGet, login.Header.Get("Location"), nil, nil)
	raw, _ := io.ReadAll(page.Body)
	_ = page.Body.Close()
	m := formAction.FindSubmatch(raw)
	if m == nil {
		b.t.Fatalf("a tela de login do Keycloak nao tem formulario (status %d)", page.StatusCode)
	}
	submit := b.do(http.MethodPost, html.UnescapeString(string(m[1])), url.Values{"username": {user}, "password": {password}, "credentialId": {""}}, nil)
	if submit.StatusCode != http.StatusFound { // senha errada: o Keycloak mostra a tela de novo (200)
		_ = submit.Body.Close()
		return submit
	}
	return b.followTo(submit, host)
}

func mustHost(t *testing.T, base string) string {
	u, err := url.Parse(base)
	if err != nil {
		t.Fatal(err)
	}
	return u.Host
}

func TestLoginWithRealKeycloakAndPostgres(t *testing.T) {
	base, query := startAPI(t)
	b := newBrowser(t)

	callback := b.signIn(base, identitytest.UserName, identitytest.UserPassword)
	if callback.StatusCode != http.StatusFound || callback.Header.Get("Location") != "/estudo" {
		t.Fatalf("callback: %d Location=%q", callback.StatusCode, callback.Header.Get("Location"))
	}

	me := b.do(http.MethodGet, base+"/api/v1/me", nil, nil)
	var body map[string]any
	if err := json.NewDecoder(me.Body).Decode(&body); err != nil || me.StatusCode != http.StatusOK {
		t.Fatalf("/me: %d %v", me.StatusCode, err)
	}
	if body["display_name"] != identitytest.UserFullName || body["id"] == "" {
		t.Fatalf("perfil inesperado: %v", body)
	}

	var users, sessions int
	var issuer, subject, refresh string
	query(`SELECT count(*) FROM app_user`, &users)
	query(`SELECT count(*) FROM auth_session`, &sessions)
	query(`SELECT idp_issuer, idp_subject FROM app_user LIMIT 1`, &issuer, &subject)
	query(`SELECT refresh_token_enc FROM auth_session LIMIT 1`, &refresh)
	if users != 1 || sessions != 1 || issuer == "" || subject == "" {
		t.Fatalf("banco: %d perfis, %d sessoes, issuer=%q subject=%q", users, sessions, issuer, subject)
	}
	if refresh == "" || bytes.Contains([]byte(refresh), []byte(".")) { // refresh do Keycloak e um JWT (tem pontos); selado nao
		t.Fatalf("o refresh token deve estar selado, nao em claro: %q", refresh)
	}

	out := b.do(http.MethodPost, base+"/api/v1/auth/logout", nil, map[string]string{"X-Requested-With": "XMLHttpRequest"})
	if out.StatusCode != http.StatusNoContent {
		t.Fatalf("logout: %d", out.StatusCode)
	}
	if after := b.do(http.MethodGet, base+"/api/v1/me", nil, nil); after.StatusCode != http.StatusUnauthorized {
		t.Fatalf("/me depois do logout: esperava 401, veio %d", after.StatusCode)
	}
	query(`SELECT count(*) FROM auth_session`, &sessions)
	if sessions != 0 {
		t.Fatalf("a sessao deveria ter sido apagada do banco, restam %d", sessions)
	}
}

func TestWrongPasswordDoesNotLogIn(t *testing.T) {
	base, query := startAPI(t)
	b := newBrowser(t)

	res := b.signIn(base, identitytest.UserName, "senha-errada")
	if res.StatusCode == http.StatusFound {
		t.Fatalf("senha errada nao pode concluir o login (Location=%q)", res.Header.Get("Location"))
	}
	if me := b.do(http.MethodGet, base+"/api/v1/me", nil, nil); me.StatusCode != http.StatusUnauthorized {
		t.Fatalf("/me: esperava 401, veio %d", me.StatusCode)
	}
	var sessions int
	query(`SELECT count(*) FROM auth_session`, &sessions)
	if sessions != 0 {
		t.Fatalf("nenhuma sessao pode existir, ha %d", sessions)
	}
}

func TestCallbackWithForgedStateIsRejected(t *testing.T) {
	base, _ := startAPI(t)
	b := newBrowser(t)
	if login := b.do(http.MethodGet, base+"/api/v1/auth/login", nil, nil); login.StatusCode != http.StatusFound {
		t.Fatalf("login: %d", login.StatusCode)
	}
	forged := b.do(http.MethodGet, base+"/api/v1/auth/callback?code=qualquer&state=forjado", nil, nil)
	if forged.StatusCode != http.StatusBadRequest {
		t.Fatalf("state forjado: esperava 400, veio %d", forged.StatusCode)
	}
}
