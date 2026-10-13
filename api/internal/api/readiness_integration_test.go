//go:build integration

package api_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/keven-rdr/wordloop/api/internal/api"
	"github.com/keven-rdr/wordloop/api/internal/platform/db"
	"github.com/keven-rdr/wordloop/api/internal/platform/db/dbtest"
)

// Banco REAL: 200 enquanto o Postgres responde e 503 depois que ele cai.
func TestReadinessFollowsRealDatabase(t *testing.T) {
	ctx := context.Background()
	url, stop := dbtest.Postgres(t)
	pool, err := db.Open(ctx, url)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	h := api.Handler(api.Server{DB: pool})
	status := func() int {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/health/ready", nil))
		return rec.Code
	}
	if got := status(); got != http.StatusOK {
		t.Fatalf("com o banco no ar: esperava 200, veio %d", got)
	}
	stop()
	if got := status(); got != http.StatusServiceUnavailable {
		t.Fatalf("com o banco fora do ar: esperava 503, veio %d", got)
	}
}
