package main

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/keven-rdr/wordloop/api/internal/api"
	httpapi "github.com/keven-rdr/wordloop/api/internal/identity/adapters/http"
	"github.com/keven-rdr/wordloop/api/internal/identity/adapters/oidc"
	"github.com/keven-rdr/wordloop/api/internal/identity/adapters/postgres"
	"github.com/keven-rdr/wordloop/api/internal/identity/app"
	"github.com/keven-rdr/wordloop/api/internal/platform/seal"
)

// buildIdentity liga o modulo de login. Sem OIDC_ISSUER o login fica desligado (503) e a API sobe normalmente;
// com OIDC_ISSUER, todas as outras variaveis sao obrigatorias.
func buildIdentity(pool *pgxpool.Pool) (api.Identity, error) {
	users, sessions := postgres.NewUsers(pool), postgres.NewSessions(pool)
	h := httpapi.Handler{Me: app.GetMe{Sessions: sessions, Users: users}, Out: app.Logout{Sessions: sessions}}

	cfg := oidc.Config{
		Issuer: os.Getenv("OIDC_ISSUER"), ClientID: os.Getenv("OIDC_CLIENT_ID"),
		ClientSecret: os.Getenv("OIDC_CLIENT_SECRET"), RedirectURL: os.Getenv("OIDC_REDIRECT_URL"),
	}
	if cfg.Issuer == "" {
		slog.Warn("OIDC_ISSUER nao definido: o login esta desligado neste ambiente")
		return h, nil
	}
	if cfg.ClientID == "" || cfg.ClientSecret == "" || cfg.RedirectURL == "" {
		return nil, errors.New("com OIDC_ISSUER, defina tambem OIDC_CLIENT_ID, OIDC_CLIENT_SECRET e OIDC_REDIRECT_URL")
	}
	key, err := seal.ParseKey(os.Getenv("SESSION_KEY"))
	if err != nil {
		return nil, err
	}
	sealer, err := seal.New(key)
	if err != nil {
		return nil, fmt.Errorf("SESSION_KEY: %w", err)
	}
	provider := oidc.New(cfg)
	h.Start = &app.StartLogin{Provider: provider, Sealer: sealer}
	h.Finish = &app.FinishLogin{Provider: provider, Users: users, Sessions: sessions, Sealer: sealer}
	h.Secure = strings.HasPrefix(cfg.RedirectURL, "https://")
	return h, nil
}
