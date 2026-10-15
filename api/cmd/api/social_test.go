package main

import (
	"bytes"
	"encoding/base64"
	"slices"
	"testing"

	httpapi "github.com/keven-rdr/wordloop/api/internal/identity/adapters/http"
	"github.com/keven-rdr/wordloop/api/internal/identity/identitytest"
	"github.com/keven-rdr/wordloop/api/internal/platform/seal"
)

func clearSocialEnv(t *testing.T) {
	t.Helper()
	for _, key := range []string{
		"OIDC_ISSUER", "OIDC_CLIENT_ID", "OIDC_CLIENT_SECRET", "OIDC_REDIRECT_URL", "APP_PUBLIC_URL",
		"GOOGLE_CLIENT_ID", "GOOGLE_CLIENT_SECRET", "GITHUB_CLIENT_ID", "GITHUB_CLIENT_SECRET", "SESSION_KEY",
	} {
		t.Setenv(key, "")
	}
}

func sessionKey() string {
	return base64.StdEncoding.EncodeToString(bytes.Repeat([]byte{4}, seal.KeySize))
}

func TestNoSocialProviderConfiguredLeavesTheOptionOff(t *testing.T) {
	clearSocialEnv(t)
	providers, err := socialProviders()
	if err != nil || len(providers) != 0 {
		t.Fatalf("sem configuracao: %v %v", providers, err)
	}
	h := httpapi.Handler{}
	if err := buildSocial(&h, &identitytest.Users{}, &identitytest.Sessions{}); err != nil || h.Start != nil || h.Finish != nil {
		t.Fatalf("o login social fica desligado: %v", err)
	}
}

func TestEachProviderNeedsItsWholePairAndACallback(t *testing.T) {
	clearSocialEnv(t)
	t.Setenv("GOOGLE_CLIENT_ID", "id")
	if _, err := socialProviders(); err == nil {
		t.Fatal("Google so com o id deve falhar")
	}
	t.Setenv("GOOGLE_CLIENT_ID", "")
	t.Setenv("GITHUB_CLIENT_SECRET", "segredo")
	if _, err := socialProviders(); err == nil {
		t.Fatal("GitHub so com o segredo deve falhar")
	}
	t.Setenv("GITHUB_CLIENT_ID", "id")
	if _, err := socialProviders(); err == nil {
		t.Fatal("sem APP_PUBLIC_URL nem OIDC_REDIRECT_URL nao ha para onde voltar")
	}
}

func TestProvidersAreBuiltFromTheEnvironmentWithTheSameCallback(t *testing.T) {
	clearSocialEnv(t)
	t.Setenv("APP_PUBLIC_URL", "https://tst.wordloop.com.br/")
	t.Setenv("GOOGLE_CLIENT_ID", "g-id")
	t.Setenv("GOOGLE_CLIENT_SECRET", "g-segredo")
	t.Setenv("GITHUB_CLIENT_ID", "h-id")
	t.Setenv("GITHUB_CLIENT_SECRET", "h-segredo")
	if got := callbackURL(); got != "https://tst.wordloop.com.br/api/v1/auth/callback" {
		t.Fatalf("callback derivado da URL publica: %q", got)
	}
	t.Setenv("OIDC_REDIRECT_URL", "https://outro.test/cb")
	if got := callbackURL(); got != "https://outro.test/cb" {
		t.Fatalf("OIDC_REDIRECT_URL tem prioridade: %q", got)
	}
	t.Setenv("OIDC_REDIRECT_URL", "")
	providers, err := socialProviders()
	if err != nil || len(providers) != 2 || providers["google"] == nil || providers["github"] == nil {
		t.Fatalf("provedores: %v %v", providers, err)
	}
}

func TestBuildSocialNeedsTheSessionKeyAndSealsTheFlow(t *testing.T) {
	clearSocialEnv(t)
	t.Setenv("APP_PUBLIC_URL", "https://tst.wordloop.com.br")
	t.Setenv("GITHUB_CLIENT_ID", "id")
	t.Setenv("GITHUB_CLIENT_SECRET", "segredo")
	h := httpapi.Handler{}
	if err := buildSocial(&h, &identitytest.Users{}, &identitytest.Sessions{}); err == nil {
		t.Fatal("com provedor e sem SESSION_KEY a API nao sobe")
	}
	t.Setenv("SESSION_KEY", sessionKey())
	if err := buildSocial(&h, &identitytest.Users{}, &identitytest.Sessions{}); err != nil {
		t.Fatal(err)
	}
	if h.Start == nil || h.Finish == nil || !slices.Equal(h.Start.Names(), []string{"github"}) || !h.Secure {
		t.Fatalf("login ligado em https: %+v", h)
	}
}

func TestTheGenericOIDCProviderStillNeedsItsTrio(t *testing.T) {
	clearSocialEnv(t)
	t.Setenv("APP_PUBLIC_URL", "https://tst.wordloop.com.br")
	t.Setenv("OIDC_ISSUER", "https://idp.test")
	if _, err := socialProviders(); err == nil {
		t.Fatal("OIDC_ISSUER sem client id e segredo deve falhar")
	}
	t.Setenv("OIDC_CLIENT_ID", "id")
	t.Setenv("OIDC_CLIENT_SECRET", "segredo")
	providers, err := socialProviders()
	if err != nil || providers["oidc"] == nil {
		t.Fatalf("oidc: %v %v", providers, err)
	}
}
