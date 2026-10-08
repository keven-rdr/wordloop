package api_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/keven-rdr/wordloop/api/internal/api"
	"github.com/keven-rdr/wordloop/api/openapi"
)

func getDocs(t *testing.T, enabled bool, path string) *httptest.ResponseRecorder {
	t.Helper()
	rec := httptest.NewRecorder()
	api.Handler(api.Server{DocsEnabled: enabled}).
		ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
	return rec
}

func TestSpecServesEmbeddedContract(t *testing.T) {
	rec := getDocs(t, true, "/api/v1/openapi.yaml")
	if rec.Code != http.StatusOK {
		t.Fatalf("esperava 200, veio %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/yaml" {
		t.Fatalf("content-type inesperado: %q", ct)
	}
	if rec.Body.String() != string(openapi.Spec) || !strings.Contains(rec.Body.String(), "openapi: 3.0.3") {
		t.Fatal("o corpo nao e o contrato embutido")
	}
}

func TestDocsPageReferencesSpec(t *testing.T) {
	rec := getDocs(t, true, "/api/v1/docs/")
	if rec.Code != http.StatusOK {
		t.Fatalf("esperava 200, veio %d", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "/api/v1/openapi.yaml") || !strings.Contains(body, "swagger-ui") {
		t.Fatalf("a pagina nao aponta para o contrato nem carrega o Swagger UI: %.200s", body)
	}
}

func TestDocsServesStaticAssetsFromBinary(t *testing.T) {
	if rec := getDocs(t, true, "/api/v1/docs/swagger-ui.css"); rec.Code != http.StatusOK {
		t.Fatalf("esperava 200 para o CSS embutido, veio %d", rec.Code)
	}
}

func TestDocsWithoutTrailingSlashRedirects(t *testing.T) {
	rec := getDocs(t, true, "/api/v1/docs")
	if rec.Code != http.StatusTemporaryRedirect || rec.Header().Get("Location") != "/api/v1/docs/" {
		t.Fatalf("esperava redirecionar para /api/v1/docs/, veio %d (%s)", rec.Code, rec.Header().Get("Location"))
	}
}

func TestDocsDisabledReturns404(t *testing.T) {
	for _, path := range []string{"/api/v1/openapi.yaml", "/api/v1/docs/"} {
		if rec := getDocs(t, false, path); rec.Code != http.StatusNotFound {
			t.Fatalf("%s: esperava 404 com a documentacao desligada, veio %d", path, rec.Code)
		}
	}
}
