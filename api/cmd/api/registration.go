package main

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"

	httpapi "github.com/keven-rdr/wordloop/api/internal/identity/adapters/http"
	"github.com/keven-rdr/wordloop/api/internal/identity/adapters/postgres"
	"github.com/keven-rdr/wordloop/api/internal/identity/registration"
	"github.com/keven-rdr/wordloop/api/internal/platform/mail"
)

func buildMail() (mail.Sender, error) {
	if strings.TrimSpace(os.Getenv("MAIL_DRIVER")) == "log" {
		if strings.HasPrefix(strings.TrimSpace(os.Getenv("APP_PUBLIC_URL")), "https://") {
			return nil, errors.New("MAIL_DRIVER=log escreve os links no log e so vale em desenvolvimento (APP_PUBLIC_URL com http)")
		}
		slog.Warn("MAIL_DRIVER=log: os e-mails (com os links) vao para o log; use so em desenvolvimento")
		return mail.Log{}, nil
	}
	key := strings.TrimSpace(os.Getenv("BREVO_API_KEY"))
	if key == "" {
		return nil, nil
	}
	sender, err := mail.NewBrevo(key, strings.TrimSpace(os.Getenv("MAIL_FROM")), strings.TrimSpace(os.Getenv("MAIL_FROM_NAME")))
	if err != nil {
		return nil, fmt.Errorf("com BREVO_API_KEY, defina tambem MAIL_FROM (o endereco remetente): %w", err)
	}
	return sender, nil
}

func buildRegistration(h *httpapi.Handler, pool *pgxpool.Pool) error {
	sender, err := buildMail()
	if err != nil {
		return err
	}
	publicURL := strings.TrimSpace(os.Getenv("APP_PUBLIC_URL"))
	if sender == nil || publicURL == "" {
		slog.Warn("cadastro e recuperacao por e-mail desligados neste ambiente: defina BREVO_API_KEY, MAIL_FROM e APP_PUBLIC_URL")
		return nil
	}
	issuer := registration.Issuer{Store: postgres.NewRegistration(pool), Mail: sender, PublicURL: publicURL, Async: true}
	h.Registration = &httpapi.Registration{Register: registration.Register{Issuer: issuer}, Forgot: registration.Forgot{Issuer: issuer}}
	return nil
}
