package api

import (
	"context"
	"net/http"

	"example.com/wordloop/api/internal/api/gen"
)

// Server implementa só o que o spike exercita; o resto do contrato fica "não implementado" (embed da interface).
type Server struct {
	gen.StrictServerInterface
	Version, Commit, Env string
}

func (s Server) GetVersion(_ context.Context, _ gen.GetVersionRequestObject) (gen.GetVersionResponseObject, error) {
	return gen.GetVersion200JSONResponse{Api: &s.Version, Commit: &s.Commit, Env: &s.Env}, nil
}

// Handler monta a API sob /api/v1 com ServeMux (net/http), sem roteador de terceiros.
func Handler(s Server) http.Handler {
	strict := gen.NewStrictHandler(s, nil)
	return gen.HandlerWithOptions(strict, gen.StdHTTPServerOptions{BaseURL: "/api/v1", BaseRouter: http.NewServeMux()})
}
