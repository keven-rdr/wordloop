package api_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"example.com/wordloop/api/internal/api"
)

func TestVersion(t *testing.T) {
	h := api.Handler(api.Server{Version: "0.1.0", Commit: "abc1234", Env: "tst"})
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/version", nil))
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

func TestBadUUIDPathParamIs400(t *testing.T) {
	h := api.Handler(api.Server{})
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/items/nao-e-uuid", nil))
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("esperava 400, veio %d", rec.Code)
	}
}

func TestMethodMuxRouting(t *testing.T) {
	h := api.Handler(api.Server{})
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/v1/version", nil)) // método errado
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("esperava 405, veio %d", rec.Code)
	}
}
