package api

import (
	"net/http"

	"github.com/swaggest/swgui/v5emb"

	"github.com/keven-rdr/wordloop/api/openapi"
)

const (
	basePath    = "/api/v1"
	specPath    = basePath + "/openapi.yaml"
	docsPath    = basePath + "/docs/"
	docsTitle   = "wordloop API"
	specContent = "application/yaml"
)

// mountDocs publica o contrato (openapi.yaml) e a tela do Swagger UI, ambos embutidos no binario
// (sem CDN). Ficam na mesma origem da API: o "Try it out" dispensa CORS e usa o cookie da sessao.
func mountDocs(mux *http.ServeMux) {
	mux.HandleFunc(http.MethodGet+" "+specPath, serveSpec)
	mux.Handle(http.MethodGet+" "+docsPath, v5emb.New(docsTitle, specPath, docsPath))
}

func serveSpec(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", specContent)
	w.Header().Set("X-Content-Type-Options", "nosniff")
	_, _ = w.Write(openapi.Spec)
}
