// Command api e o composition root: monta as dependencias e sobe o servidor HTTP.
package main

import (
	"errors"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/keven-rdr/wordloop/api/internal/api"
	"github.com/keven-rdr/wordloop/api/internal/platform/version"
)

func main() {
	addr := envOr("API_ADDR", ":8080")
	env := envOr("APP_ENV", version.Env)
	srv := &http.Server{
		Addr:              addr,
		Handler:           api.Handler(api.Server{Version: version.Version, Commit: version.Commit, Env: env}),
		ReadHeaderTimeout: 5 * time.Second,
	}
	slog.Info("api iniciando", "addr", addr, "version", version.Version, "commit", version.Commit, "env", env)
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		slog.Error("servidor encerrou", "err", err)
		os.Exit(1)
	}
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
