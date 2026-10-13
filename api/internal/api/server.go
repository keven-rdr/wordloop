package api

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/keven-rdr/wordloop/api/internal/api/gen"
	"github.com/keven-rdr/wordloop/api/internal/platform/httpx"
)

// Pinger confirma que o banco responde (implementado por *pgxpool.Pool). Declarada aqui, pelo consumidor.
type Pinger interface {
	Ping(ctx context.Context) error
}

// Server implementa a interface strict gerada a partir do contrato.
type Server struct {
	Version, Commit, Env string
	// DocsEnabled publica o contrato e o Swagger UI em /api/v1/openapi.yaml e /api/v1/docs/.
	DocsEnabled bool
	// DB e o banco real; sem ele a API nunca fica "pronta" (readiness 503).
	DB Pinger
	// Identity e o modulo de login (BFF); nulo = login indisponivel.
	Identity Identity
}

var _ gen.StrictServerInterface = Server{}

// GetVersion devolve a identidade do build.
func (s Server) GetVersion(_ context.Context, _ gen.GetVersionRequestObject) (gen.GetVersionResponseObject, error) {
	return gen.GetVersion200JSONResponse{Api: s.Version, Commit: s.Commit, Env: s.Env}, nil
}

// GetReadiness responde 200 quando o banco responde e 503 quando nao (ou quando nao ha banco configurado).
func (s Server) GetReadiness(ctx context.Context, _ gen.GetReadinessRequestObject) (gen.GetReadinessResponseObject, error) {
	if s.DB == nil {
		return gen.GetReadiness503Response{}, nil
	}
	if err := s.DB.Ping(ctx); err != nil {
		slog.Warn("readiness: banco indisponivel", "err", err)
		return gen.GetReadiness503Response{}, nil
	}
	return gen.GetReadiness200Response{}, nil
}

// Handler monta a API sob /api/v1 com o ServeMux do net/http.
func Handler(s Server) http.Handler {
	mux := http.NewServeMux()
	if s.DocsEnabled {
		mountDocs(mux)
	}
	strict := gen.NewStrictHandlerWithOptions(s, nil, gen.StrictHTTPServerOptions{
		RequestErrorHandlerFunc:  func(w http.ResponseWriter, _ *http.Request, _ error) { w.WriteHeader(http.StatusBadRequest) },
		ResponseErrorHandlerFunc: internalError,
	})
	api := gen.HandlerWithOptions(strict, gen.StdHTTPServerOptions{BaseURL: basePath, BaseRouter: mux})
	return httpx.WithRequest(api)
}

// internalError registra o erro e responde 500 sem corpo: o texto do erro nunca vai ao cliente.
func internalError(w http.ResponseWriter, r *http.Request, err error) {
	slog.Error("erro interno", "method", r.Method, "path", r.URL.Path, "err", err)
	w.WriteHeader(http.StatusInternalServerError)
}
