package main

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/keven-rdr/wordloop/api/internal/api"
	"github.com/keven-rdr/wordloop/api/internal/platform/ratelimit"
	"github.com/keven-rdr/wordloop/api/internal/platform/version"
)

const (
	shutdownTimeout       = 15 * time.Second
	authRequestsPerMinute = 30
	authBurst             = 15
)

func main() { os.Exit(realMain(os.Args[1:])) }

// realMain devolve o codigo de saida do processo: 0 em sucesso e 1 em erro.
func realMain(args []string) int {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	if err := execute(ctx, args); err != nil {
		slog.Error("api encerrou com erro", "err", err)
		return 1
	}
	return 0
}

// execute escolhe o modo: `api migrate` aplica as migracoes e sai; sem argumento, serve a API.
func execute(ctx context.Context, args []string) error {
	if len(args) > 0 {
		switch args[0] {
		case "migrate":
			return migrateUp(ctx)
		case "create-admin":
			return createAdmin(ctx, args[1:], os.Stdout)
		}
	}
	return serve(ctx)
}

func serve(ctx context.Context) error {
	pool, err := openDatabase(ctx)
	if err != nil {
		return err
	}
	defer pool.Close()
	ident, err := buildIdentity(pool)
	if err != nil {
		return err
	}
	info := startupInfo{
		Env: envOr("APP_ENV", version.Env), Version: version.Version, Commit: version.Commit,
		DocsEnabled: os.Getenv("API_DOCS_ENABLED") != "false", // ligado por padrao; "false" desliga
	}
	srv := &http.Server{
		Addr: envOr("API_ADDR", ":8080"),
		Handler: api.Handler(api.Server{
			Version: info.Version, Commit: info.Commit, Env: info.Env, DocsEnabled: info.DocsEnabled, DB: pool, Identity: ident,
			AuthLimiter: authLimiter(),
		}),
		ReadHeaderTimeout: 5 * time.Second,
	}
	if err := run(ctx, srv, func(bound net.Addr) { announce(os.Stdout, info, bound) }); err != nil {
		return err
	}
	slog.Info("api encerrada")
	return nil
}

// run sobe o servidor e o encerra com elegancia quando ctx termina (SIGTERM/SIGINT; docker stop envia SIGTERM).
// onReady e chamado quando a porta ja esta aberta, isto e, quando a API passou a aceitar conexoes.
func run(ctx context.Context, srv *http.Server, onReady func(net.Addr)) error {
	ln, err := net.Listen("tcp", srv.Addr)
	if err != nil {
		return err
	}
	onReady(ln.Addr())
	errCh := make(chan error, 1)
	go func() { errCh <- srv.Serve(ln) }()

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

func authLimiter() *ratelimit.Limiter {
	l := ratelimit.New(authRequestsPerMinute, authBurst)
	l.TrustProxy = os.Getenv("API_TRUST_PROXY") == "true"
	return l
}
