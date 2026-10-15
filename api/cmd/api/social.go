package main

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/keven-rdr/wordloop/api/internal/identity/adapters/github"
	httpapi "github.com/keven-rdr/wordloop/api/internal/identity/adapters/http"
	"github.com/keven-rdr/wordloop/api/internal/identity/adapters/oidc"
	"github.com/keven-rdr/wordloop/api/internal/identity/app"
	"github.com/keven-rdr/wordloop/api/internal/platform/seal"
)

const (
	googleIssuer = "https://accounts.google.com"
	callbackPath = "/api/v1/auth/callback"
)

func callbackURL() string {
	if url := strings.TrimSpace(os.Getenv("OIDC_REDIRECT_URL")); url != "" {
		return url
	}
	if base := strings.TrimRight(strings.TrimSpace(os.Getenv("APP_PUBLIC_URL")), "/"); base != "" {
		return base + callbackPath
	}
	return ""
}

func pair(idKey, secretKey string) (id, secret string, set bool, err error) {
	id, secret = strings.TrimSpace(os.Getenv(idKey)), strings.TrimSpace(os.Getenv(secretKey))
	if id == "" && secret == "" {
		return "", "", false, nil
	}
	if id == "" || secret == "" {
		return "", "", false, fmt.Errorf("defina %s e %s juntos", idKey, secretKey)
	}
	return id, secret, true, nil
}

func socialProviders() (app.Providers, error) {
	providers := app.Providers{}
	redirect := callbackURL()
	for _, add := range []func(app.Providers, string) error{addOIDC, addGoogle, addGitHub} {
		if err := add(providers, redirect); err != nil {
			return nil, err
		}
	}
	if len(providers) > 0 && redirect == "" {
		return nil, errors.New("defina APP_PUBLIC_URL (ou OIDC_REDIRECT_URL): e para onde o provedor devolve o navegador")
	}
	return providers, nil
}

func addOIDC(providers app.Providers, redirect string) error {
	issuer := strings.TrimSpace(os.Getenv("OIDC_ISSUER"))
	if issuer == "" {
		return nil
	}
	id, secret, set, err := pair("OIDC_CLIENT_ID", "OIDC_CLIENT_SECRET")
	if err != nil || !set || redirect == "" {
		return errors.New("com OIDC_ISSUER, defina tambem OIDC_CLIENT_ID, OIDC_CLIENT_SECRET e OIDC_REDIRECT_URL")
	}
	providers["oidc"] = oidc.New(oidc.Config{Issuer: issuer, ClientID: id, ClientSecret: secret, RedirectURL: redirect})
	return nil
}

func addGoogle(providers app.Providers, redirect string) error {
	id, secret, set, err := pair("GOOGLE_CLIENT_ID", "GOOGLE_CLIENT_SECRET")
	if err != nil || !set {
		return err
	}
	providers["google"] = oidc.New(oidc.Config{Issuer: googleIssuer, ClientID: id, ClientSecret: secret, RedirectURL: redirect})
	return nil
}

func addGitHub(providers app.Providers, redirect string) error {
	id, secret, set, err := pair("GITHUB_CLIENT_ID", "GITHUB_CLIENT_SECRET")
	if err != nil || !set {
		return err
	}
	providers["github"] = github.New(github.Config{ClientID: id, ClientSecret: secret, RedirectURL: redirect})
	return nil
}

func buildSocial(h *httpapi.Handler, users app.Users, sessions app.Sessions) error {
	providers, err := socialProviders()
	if err != nil {
		return err
	}
	if len(providers) == 0 {
		slog.Warn("nenhum provedor de login social configurado (GOOGLE_*, GITHUB_*, OIDC_*): so e-mail e senha e passkey")
		return nil
	}
	key, err := seal.ParseKey(os.Getenv("SESSION_KEY"))
	if err != nil {
		return err
	}
	sealer, err := seal.New(key)
	if err != nil {
		return fmt.Errorf("SESSION_KEY: %w", err)
	}
	h.Start = &app.StartLogin{Providers: providers, Sealer: sealer}
	h.Finish = &app.FinishLogin{Providers: providers, Users: users, Sessions: sessions, Sealer: sealer}
	h.Secure = h.Secure || strings.HasPrefix(callbackURL(), "https://")
	slog.Info("login social ligado", "provedores", h.Start.Names())
	return nil
}
