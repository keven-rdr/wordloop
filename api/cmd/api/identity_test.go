package main

import (
	"bytes"
	"encoding/base64"
	"testing"

	"github.com/keven-rdr/wordloop/api/internal/platform/seal"
)

func setOIDC(t *testing.T, issuer, id, secret, redirect, key string) {
	t.Helper()
	t.Setenv("OIDC_ISSUER", issuer)
	t.Setenv("OIDC_CLIENT_ID", id)
	t.Setenv("OIDC_CLIENT_SECRET", secret)
	t.Setenv("OIDC_REDIRECT_URL", redirect)
	t.Setenv("SESSION_KEY", key)
}

func TestBuildIdentityWithoutIssuerKeepsLoginOff(t *testing.T) {
	setOIDC(t, "", "", "", "", "")
	if _, err := buildIdentity(nil); err != nil {
		t.Fatalf("sem OIDC_ISSUER a API deve subir com o login desligado: %v", err)
	}
}

func TestBuildIdentityRequiresTheRestOfTheConfiguration(t *testing.T) {
	key := base64.StdEncoding.EncodeToString(bytes.Repeat([]byte{1}, seal.KeySize))
	cases := map[string][5]string{
		"sem client id":     {"https://idp.test", "", "s", "https://app/cb", key},
		"sem client secret": {"https://idp.test", "id", "", "https://app/cb", key},
		"sem redirect":      {"https://idp.test", "id", "s", "", key},
		"sem chave":         {"https://idp.test", "id", "s", "https://app/cb", ""},
		"chave curta":       {"https://idp.test", "id", "s", "https://app/cb", base64.StdEncoding.EncodeToString([]byte("curta"))},
	}
	for name, c := range cases {
		setOIDC(t, c[0], c[1], c[2], c[3], c[4])
		if _, err := buildIdentity(nil); err == nil {
			t.Errorf("%s: esperava erro de configuracao", name)
		}
	}
}

func TestBuildIdentityWithCompleteConfigurationSucceeds(t *testing.T) {
	key := base64.StdEncoding.EncodeToString(bytes.Repeat([]byte{1}, seal.KeySize))
	setOIDC(t, "https://idp.test", "id", "s", "https://app/api/v1/auth/callback", key)
	if _, err := buildIdentity(nil); err != nil {
		t.Fatalf("configuracao completa deveria funcionar (a descoberta so acontece no primeiro login): %v", err)
	}
}
