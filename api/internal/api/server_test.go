package api_test

import (
	"encoding/json"
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

func TestReadiness(t *testing.T) {
	if rec := get(t, "/api/v1/health/ready"); rec.Code != http.StatusOK {
		t.Fatalf("esperava 200, veio %d", rec.Code)
	}
}

func TestWrongMethodIs405(t *testing.T) {
	rec := httptest.NewRecorder()
	api.Handler(api.Server{}).ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/v1/version", nil))
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("esperava 405, veio %d", rec.Code)
	}
}
