package api_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/keven-rdr/wordloop/api/internal/api"
)

func get(t *testing.T, path string) *httptest.ResponseRecorder {
	t.Helper()
	rec := httptest.NewRecorder()
	api.Handler(api.Server{Version: "0.1.0", Commit: "abc1234", Env: "tst"}).
		ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
	return rec
}

func TestVersion(t *testing.T) {
	rec := get(t, "/api/v1/version")
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d: %s", rec.Code, rec.Body.String())
	}
	var v map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &v); err != nil {
		t.Fatal(err)
	}
	if v["api"] != "0.1.0" || v["commit"] != "abc1234" || v["env"] != "tst" {
		t.Fatalf("corpo inesperado: %v", v)
	}
}

// pinger e o banco de mentira SO neste teste de unidade; o banco real e coberto pelo teste de integracao.
type pinger struct{ err error }

func (p pinger) Ping(context.Context) error { return p.err }

func readiness(s api.Server) int {
	rec := httptest.NewRecorder()
	api.Handler(s).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/health/ready", nil))
	return rec.Code
}

func TestReadiness(t *testing.T) {
	cases := map[string]struct {
		db   api.Pinger
		want int
	}{
		"banco responde":     {pinger{}, http.StatusOK},
		"banco indisponivel": {pinger{errors.New("fora do ar")}, http.StatusServiceUnavailable},
		"sem banco":          {nil, http.StatusServiceUnavailable},
	}
	for name, c := range cases {
		if got := readiness(api.Server{DB: c.db}); got != c.want {
			t.Errorf("%s: esperava %d, veio %d", name, c.want, got)
		}
	}
}

func TestWrongMethodIs405(t *testing.T) {
	rec := httptest.NewRecorder()
	api.Handler(api.Server{}).ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/v1/version", nil))
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("esperava 405, veio %d", rec.Code)
	}
}
