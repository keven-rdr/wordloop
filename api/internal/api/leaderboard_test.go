package api_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/keven-rdr/wordloop/api/internal/api"
)

// Sem o modulo do ranking ligado, as rotas respondem 401 (como os demais modulos de quem estuda).
func TestLeaderboardRoutesWithoutTheModuleAreUnauthorized(t *testing.T) {
	h := api.Handler(api.Server{})
	cases := []struct{ method, path, body string }{
		{http.MethodGet, "/api/v1/leaderboard", ""},
		{http.MethodGet, "/api/v1/leaderboard/me", ""},
		{http.MethodGet, "/api/v1/leaderboard/profile", ""},
		{http.MethodPut, "/api/v1/leaderboard/profile", `{"nickname":"ana_1","visible":true}`},
	}
	for _, c := range cases {
		req := httptest.NewRequest(c.method, c.path, strings.NewReader(c.body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Requested-With", "XMLHttpRequest")
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		if rec.Code != http.StatusUnauthorized {
			t.Errorf("%s %s: %d, esperava 401", c.method, c.path, rec.Code)
		}
	}
}
