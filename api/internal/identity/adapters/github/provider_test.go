package github

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"golang.org/x/oauth2"
)

type fake struct {
	srv       *httptest.Server
	tokenForm url.Values
	user      string
	emails    string
	userCode  int
	badToken  bool
}

func newFake(t *testing.T) (*fake, *Provider) {
	t.Helper()
	f := &fake{
		user:     `{"id": 42, "login": "ana-dev", "name": "Ana Souza"}`,
		emails:   `[{"email":"outro@exemplo.com","primary":false,"verified":true},{"email":"ana@exemplo.com","primary":true,"verified":true}]`,
		userCode: http.StatusOK,
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/login/oauth/access_token", func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		f.tokenForm = r.PostForm
		w.Header().Set("Content-Type", "application/json")
		if f.badToken {
			_, _ = w.Write([]byte(`{"error":"bad_verification_code"}`))
			return
		}
		_, _ = w.Write([]byte(`{"access_token":"gho_teste","token_type":"bearer","scope":"read:user,user:email"}`))
	})
	mux.HandleFunc("/user", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer gho_teste" || r.Header.Get("User-Agent") == "" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.WriteHeader(f.userCode)
		_, _ = w.Write([]byte(f.user))
	})
	mux.HandleFunc("/user/emails", func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write([]byte(f.emails)) })
	f.srv = httptest.NewServer(mux)
	t.Cleanup(f.srv.Close)
	p := New(Config{ClientID: "id", ClientSecret: "segredo", RedirectURL: "https://app.test/api/v1/auth/callback"})
	p.oauth.Endpoint = oauth2.Endpoint{AuthURL: f.srv.URL + "/login/oauth/authorize", TokenURL: f.srv.URL + "/login/oauth/access_token"}
	p.apiBase = f.srv.URL
	return f, p
}

func TestAuthURLCarriesStateScopesAndPKCE(t *testing.T) {
	_, p := newFake(t)
	raw, err := p.AuthURL(context.Background(), "estado-1", "ignorado", "verificador-pkce-bem-longo-0123456789012345678901")
	if err != nil {
		t.Fatal(err)
	}
	u, _ := url.Parse(raw)
	q := u.Query()
	if q.Get("state") != "estado-1" || q.Get("client_id") != "id" || q.Get("code_challenge_method") != "S256" || q.Get("code_challenge") == "" {
		t.Fatalf("url: %s", raw)
	}
	if !strings.Contains(q.Get("scope"), "user:email") || q.Get("redirect_uri") != "https://app.test/api/v1/auth/callback" {
		t.Fatalf("escopo ou redirecionamento: %s", raw)
	}
}

func TestExchangeReturnsThePrimaryVerifiedEmailAndSendsThePKCEVerifier(t *testing.T) {
	f, p := newFake(t)
	tokens, err := p.Exchange(context.Background(), "codigo-1", "verificador-pkce-bem-longo-0123456789012345678901", "")
	if err != nil {
		t.Fatal(err)
	}
	id := tokens.Identity
	if id.Issuer != "github" || id.Subject != "42" || id.DisplayName != "Ana Souza" || id.Email != "ana@exemplo.com" || !id.EmailVerified || id.MFA {
		t.Fatalf("identidade: %+v", id)
	}
	if f.tokenForm.Get("code") != "codigo-1" || f.tokenForm.Get("code_verifier") == "" {
		t.Fatalf("troca do code: %v", f.tokenForm)
	}
	if tokens.RefreshToken != "" {
		t.Fatal("o GitHub nao tem refresh token")
	}
}

func TestOnlyVerifiedEmailsCount(t *testing.T) {
	cases := map[string]struct {
		emails   string
		want     string
		verified bool
	}{
		"nenhum confirmado":       {`[{"email":"x@exemplo.com","primary":true,"verified":false}]`, "", false},
		"sem e-mails":             {`[]`, "", false},
		"so um confirmado":        {`[{"email":"a@exemplo.com","primary":true,"verified":false},{"email":"b@exemplo.com","primary":false,"verified":true}]`, "b@exemplo.com", true},
		"primario nao confirmado": {`[{"email":"a@exemplo.com","primary":true,"verified":false}]`, "", false},
	}
	for name, c := range cases {
		f, p := newFake(t)
		f.emails = c.emails
		tokens, err := p.Exchange(context.Background(), "c", "verificador-pkce-bem-longo-0123456789012345678901", "")
		if err != nil || tokens.Identity.Email != c.want || tokens.Identity.EmailVerified != c.verified {
			t.Errorf("%s: %+v %v", name, tokens.Identity, err)
		}
	}
}

func TestLoginFallsBackToTheUsernameWhenThereIsNoName(t *testing.T) {
	f, p := newFake(t)
	f.user = `{"id": 7, "login": "so-login", "name": ""}`
	tokens, err := p.Exchange(context.Background(), "c", "verificador-pkce-bem-longo-0123456789012345678901", "")
	if err != nil || tokens.Identity.DisplayName != "so-login" || tokens.Identity.Subject != "7" {
		t.Fatalf("%+v %v", tokens.Identity, err)
	}
}

func TestExchangeFailuresAreReported(t *testing.T) {
	v := "verificador-pkce-bem-longo-0123456789012345678901"
	f, p := newFake(t)
	f.badToken = true
	if _, err := p.Exchange(context.Background(), "ruim", v, ""); err == nil {
		t.Fatal("code recusado deve falhar")
	}
	f, p = newFake(t)
	f.userCode = http.StatusForbidden
	if _, err := p.Exchange(context.Background(), "c", v, ""); err == nil {
		t.Fatal("API do GitHub recusando deve falhar")
	}
	f, p = newFake(t)
	f.user = `{"login": "sem-id"}`
	if _, err := p.Exchange(context.Background(), "c", v, ""); err == nil {
		t.Fatal("resposta sem id deve falhar")
	}
	f, p = newFake(t)
	f.user = `nao e json`
	if _, err := p.Exchange(context.Background(), "c", v, ""); err == nil {
		t.Fatal("JSON invalido deve falhar")
	}
}
