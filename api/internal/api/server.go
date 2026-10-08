package api

import (
	"context"
	"net/http"

	"github.com/keven-rdr/wordloop/api/internal/api/gen"
)

// Server implementa a interface strict gerada a partir do contrato.
type Server struct {
	Version, Commit, Env string
	// DocsEnabled publica o contrato e o Swagger UI em /api/v1/openapi.yaml e /api/v1/docs/.
	DocsEnabled bool
}

var _ gen.StrictServerInterface = Server{}

// GetVersion devolve a identidade do build.
func (s Server) GetVersion(_ context.Context, _ gen.GetVersionRequestObject) (gen.GetVersionResponseObject, error) {
	return gen.GetVersion200JSONResponse{Api: s.Version, Commit: s.Commit, Env: s.Env}, nil
}

// GetReadiness responde 200 enquanto o processo aceita requisicoes.
// Quando houver banco, passa a checar a conexao.
func (Server) GetReadiness(_ context.Context, _ gen.GetReadinessRequestObject) (gen.GetReadinessResponseObject, error) {
	return gen.GetReadiness200Response{}, nil
}

// Handler monta a API sob /api/v1 com o ServeMux do net/http.
func Handler(s Server) http.Handler {
	mux := http.NewServeMux()
	if s.DocsEnabled {
		mountDocs(mux)
	}
	strict := gen.NewStrictHandler(s, nil)
	return gen.HandlerWithOptions(strict, gen.StdHTTPServerOptions{BaseURL: basePath, BaseRouter: mux})
}
