package main

import (
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/keven-rdr/wordloop/api/internal/identity/accounts"
	httpapi "github.com/keven-rdr/wordloop/api/internal/identity/adapters/http"
	"github.com/keven-rdr/wordloop/api/internal/identity/adapters/postgres"
	"github.com/keven-rdr/wordloop/api/internal/identity/app"
)

// buildIdentity liga o modulo de login: e-mail e senha, passkeys e, se configurados, Google, GitHub e OIDC generico.
func buildIdentity(pool *pgxpool.Pool) (httpapi.Handler, error) {
	users, sessions := postgres.NewUsers(pool), postgres.NewSessions(pool)
	auth := app.Authenticate{Sessions: sessions, Users: users}
	h := httpapi.Handler{
		Auth: auth, Out: app.Logout{Sessions: sessions},
		List:   app.ListSessions{Auth: auth, Sessions: sessions},
		Revoke: app.RevokeSession{Auth: auth, Sessions: sessions}, RevokeAll: app.RevokeAllSessions{Auth: auth, Sessions: sessions},
		TwoFactor: accounts.TwoFactor{Store: postgres.NewAccounts(pool)},
	}

	if err := buildLocalLogin(&h, pool, users, sessions); err != nil {
		return httpapi.Handler{}, err
	}

	if err := buildSocial(&h, users, sessions); err != nil {
		return httpapi.Handler{}, err
	}
	return h, nil
}
