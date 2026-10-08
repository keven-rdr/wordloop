package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/keven-rdr/wordloop/api/internal/api"
	"github.com/keven-rdr/wordloop/api/internal/platform/version"
)

const shutdownTimeout = 15 * time.Second

func main() {
	addr := envOr("API_ADDR", ":8080")
	env := envOr("APP_ENV", version.Env)
	srv := &http.Server{
		Addr: addr,
		Handler: api.Handler(api.Server{
			Version: version.Version, Commit: version.Commit, Env: env,
			DocsEnabled: os.Getenv("API_DOCS_ENABLED") != "false", // ligado por padrao; "false" desliga
		}),
		ReadHeaderTimeout: 5 * time.Second,
	}
	slog.Info("api iniciando", "addr", addr, "version", version.Version, "commit", version.Commit, "env", env)
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	if err := run(ctx, srv); err != nil {
		slog.Error("servidor encerrou com erro", "err", err)
		os.Exit(1)
	}
	slog.Info("api encerrada")
}

// run sobe o servidor e o encerra com elegancia quando ctx termina (SIGTERM/SIGINT; docker stop envia SIGTERM).
func run(ctx context.Context, srv *http.Server) error {
	errCh := make(chan error, 1)
	go func() { errCh <- srv.ListenAndServe() }()

	select {
	case err := <-errCh:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	case <-ctx.Done():
		slog.Info("sinal de encerramento recebido; aguardando requisicoes em andamento")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()
		return srv.Shutdown(shutdownCtx)
	}
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
