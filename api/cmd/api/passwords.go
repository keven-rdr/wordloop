package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"

	httpapi "github.com/keven-rdr/wordloop/api/internal/identity/adapters/http"
	"github.com/keven-rdr/wordloop/api/internal/identity/adapters/postgres"
	"github.com/keven-rdr/wordloop/api/internal/identity/app"
	"github.com/keven-rdr/wordloop/api/internal/identity/passwords"
	"github.com/keven-rdr/wordloop/api/internal/platform/password"
)

const hashConcurrency = 2

func buildLocalLogin(h *httpapi.Handler, pool *pgxpool.Pool, users app.Users, sessions app.Sessions) error {
	if err := buildPasskeys(h, pool, users, sessions); err != nil {
		return err
	}
	if err := buildPasswords(h, pool, users, sessions); err != nil {
		return err
	}
	if err := buildRegistration(h, pool); err != nil {
		return err
	}
	h.Secure = h.Secure || strings.HasPrefix(strings.TrimSpace(os.Getenv("APP_PUBLIC_URL")), "https://")
	return nil
}

func buildPasswords(h *httpapi.Handler, pool *pgxpool.Pool, users app.Users, sessions app.Sessions) error {
	hasher := password.NewHasher(hashConcurrency)
	decoy := make([]byte, 24)
	if _, err := rand.Read(decoy); err != nil {
		return fmt.Errorf("gerar a senha isca: %w", err)
	}
	dummy, err := hasher.Hash(context.Background(), base64.RawStdEncoding.EncodeToString(decoy))
	if err != nil {
		return fmt.Errorf("gerar o hash isca: %w", err)
	}
	store := postgres.NewPasswords(pool)
	opener := app.SessionOpener{Sessions: sessions}
	h.Passwords = &httpapi.Passwords{
		Login: passwords.Login{Store: store, Hasher: hasher, Users: users, Opener: opener, DummyHash: dummy},
		Set:   passwords.SetPassword{Store: store, Hasher: hasher, Users: users, Opener: opener},
	}
	return nil
}
