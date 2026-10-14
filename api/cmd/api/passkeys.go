package main

import (
	"errors"
	"log/slog"
	"net/url"
	"os"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"

	httpapi "github.com/keven-rdr/wordloop/api/internal/identity/adapters/http"
	"github.com/keven-rdr/wordloop/api/internal/identity/adapters/postgres"
	"github.com/keven-rdr/wordloop/api/internal/identity/adapters/webauthn"
	"github.com/keven-rdr/wordloop/api/internal/identity/app"
	"github.com/keven-rdr/wordloop/api/internal/identity/passkeys"
)

const defaultRPName = "wordloop"

func buildPasskeys(h *httpapi.Handler, pool *pgxpool.Pool, users app.Users, sessions app.Sessions) error {
	rpID, origin := strings.TrimSpace(os.Getenv("WEBAUTHN_RP_ID")), strings.TrimSpace(os.Getenv("WEBAUTHN_ORIGIN"))
	if rpID == "" && origin == "" {
		slog.Warn("WEBAUTHN_RP_ID e WEBAUTHN_ORIGIN nao definidos: o login por passkey esta desligado neste ambiente")
		return nil
	}
	if err := validateWebAuthn(rpID, origin); err != nil {
		return err
	}
	ceremony, err := webauthn.New(webauthn.Config{RPID: rpID, RPName: envOr("WEBAUTHN_RP_NAME", defaultRPName), Origins: []string{origin}})
	if err != nil {
		return err
	}
	store := postgres.NewPasskeys(pool)
	opener := app.SessionOpener{Sessions: sessions}
	h.Passkeys = &httpapi.Passkeys{
		BeginRegistration:  passkeys.BeginRegistration{Ceremony: ceremony, Store: store, Challenges: store},
		FinishRegistration: passkeys.FinishRegistration{Ceremony: ceremony, Store: store, Challenges: store, Users: users, Opener: opener},
		BeginLogin:         passkeys.BeginLogin{Ceremony: ceremony, Challenges: store},
		FinishLogin:        passkeys.FinishLogin{Ceremony: ceremony, Store: store, Challenges: store, Users: users, Opener: opener},
		List:               passkeys.List{Store: store},
		Delete:             passkeys.Delete{Store: store},
	}
	h.Secure = h.Secure || strings.HasPrefix(origin, "https://")
	return nil
}

func validateWebAuthn(rpID, origin string) error {
	if rpID == "" || origin == "" {
		return errors.New("defina WEBAUTHN_RP_ID e WEBAUTHN_ORIGIN juntos")
	}
	u, err := url.Parse(origin)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" || u.Path != "" {
		return errors.New("WEBAUTHN_ORIGIN invalida: use http(s)://host[:porta], sem caminho")
	}
	if u.Hostname() != rpID && !strings.HasSuffix(u.Hostname(), "."+rpID) {
		return errors.New("WEBAUTHN_RP_ID precisa ser o host de WEBAUTHN_ORIGIN ou um dominio acima dele")
	}
	return nil
}
