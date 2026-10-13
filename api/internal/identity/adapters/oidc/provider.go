// Package oidc e o adaptador do provedor OIDC (code + PKCE) sobre go-oidc e oauth2. Serve qualquer provedor padrao:
// Keycloak em dev e um IdP externo em tst.
package oidc

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"

	gooidc "github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"

	"github.com/keven-rdr/wordloop/api/internal/identity/app"
	"github.com/keven-rdr/wordloop/api/internal/identity/domain"
)

// Config vem do ambiente (OIDC_ISSUER, OIDC_CLIENT_ID, OIDC_CLIENT_SECRET, OIDC_REDIRECT_URL).
type Config struct {
	Issuer, ClientID, ClientSecret, RedirectURL string
}

const httpTimeout = 10 * time.Second

// Provider descobre o provedor sob demanda: se ele estiver fora do ar na subida, a API sobe e tenta de novo no proximo login.
type Provider struct {
	cfg Config

	mu       sync.Mutex
	oauth    *oauth2.Config
	verifier *gooidc.IDTokenVerifier
}

var _ app.Provider = (*Provider)(nil)

// New nao faz rede; a descoberta acontece no primeiro uso.
func New(cfg Config) *Provider { return &Provider{cfg: cfg} }

func (p *Provider) discover() (*oauth2.Config, *gooidc.IDTokenVerifier, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.oauth != nil {
		return p.oauth, p.verifier, nil
	}
	// Contexto proprio: a descoberta e o JWKS vivem alem do pedido que os disparou.
	ctx := gooidc.ClientContext(context.Background(), &http.Client{Timeout: httpTimeout})
	prov, err := gooidc.NewProvider(ctx, p.cfg.Issuer)
	if err != nil {
		return nil, nil, fmt.Errorf("descobrir o provedor OIDC: %w", err)
	}
	p.oauth = &oauth2.Config{
		ClientID: p.cfg.ClientID, ClientSecret: p.cfg.ClientSecret, RedirectURL: p.cfg.RedirectURL,
		Endpoint: prov.Endpoint(), Scopes: []string{gooidc.ScopeOpenID, "profile"},
	}
	p.verifier = prov.Verifier(&gooidc.Config{ClientID: p.cfg.ClientID})
	return p.oauth, p.verifier, nil
}

// AuthURL monta a URL de autorizacao com state, nonce e PKCE (S256).
func (p *Provider) AuthURL(_ context.Context, state, nonce, verifier string) (string, error) {
	cfg, _, err := p.discover()
	if err != nil {
		return "", err
	}
	return cfg.AuthCodeURL(state, gooidc.Nonce(nonce), oauth2.S256ChallengeOption(verifier)), nil
}

// Exchange troca o code por tokens, valida o ID token (assinatura, emissor, audiencia, validade e nonce) e le quem e.
func (p *Provider) Exchange(ctx context.Context, code, verifier, nonce string) (app.Tokens, error) {
	cfg, idv, err := p.discover()
	if err != nil {
		return app.Tokens{}, err
	}
	ctx = gooidc.ClientContext(ctx, &http.Client{Timeout: httpTimeout})
	tok, err := cfg.Exchange(ctx, code, oauth2.VerifierOption(verifier))
	if err != nil {
		return app.Tokens{}, fmt.Errorf("trocar o code: %w", err)
	}
	raw, _ := tok.Extra("id_token").(string)
	if raw == "" {
		return app.Tokens{}, errors.New("o provedor nao devolveu id_token")
	}
	idt, err := idv.Verify(ctx, raw)
	if err != nil {
		return app.Tokens{}, fmt.Errorf("validar o id_token: %w", err)
	}
	if idt.Nonce != nonce {
		return app.Tokens{}, errors.New("nonce do id_token diferente do esperado")
	}
	var claims struct {
		Name              string `json:"name"`
		PreferredUsername string `json:"preferred_username"`
	}
	if err := idt.Claims(&claims); err != nil {
		return app.Tokens{}, fmt.Errorf("ler as claims: %w", err)
	}
	name := claims.Name
	if name == "" {
		name = claims.PreferredUsername
	}
	return app.Tokens{
		Identity:     domain.Identity{Issuer: idt.Issuer, Subject: idt.Subject, DisplayName: name},
		RefreshToken: tok.RefreshToken,
	}, nil
}
