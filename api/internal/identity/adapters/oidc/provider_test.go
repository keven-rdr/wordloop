package oidc_test

import (
	"context"
	"testing"

	"github.com/keven-rdr/wordloop/api/internal/identity/adapters/oidc"
)

// Provedor fora do ar: o erro volta (a API responde 503) em vez de derrubar o processo, e tenta de novo no proximo login.
func TestUnreachableProviderReturnsErrors(t *testing.T) {
	p := oidc.New(oidc.Config{Issuer: "http://127.0.0.1:1/realms/x", ClientID: "id", ClientSecret: "s", RedirectURL: "http://app/cb"})
	if _, err := p.AuthURL(context.Background(), "s", "n", "v"); err == nil {
		t.Fatal("AuthURL: esperava erro com o provedor inalcancavel")
	}
	if _, err := p.Exchange(context.Background(), "code", "v", "n"); err == nil {
		t.Fatal("Exchange: esperava erro com o provedor inalcancavel")
	}
}
