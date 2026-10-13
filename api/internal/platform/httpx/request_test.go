package httpx_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/keven-rdr/wordloop/api/internal/platform/httpx"
)

func TestWithRequestExposesCookiesAndUserAgent(t *testing.T) {
	var cookie, ua string
	h := httpx.WithRequest(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		cookie, ua = httpx.Cookie(r.Context(), "wl_session"), httpx.UserAgent(r.Context())
	}))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: "wl_session", Value: "abc"})
	req.Header.Set("User-Agent", strings.Repeat("x", 500))
	h.ServeHTTP(httptest.NewRecorder(), req)
	if cookie != "abc" || len(ua) != 200 {
		t.Fatalf("cookie=%q, tamanho do user agent=%d", cookie, len(ua))
	}
}

func TestWithoutRequestInContextReturnsEmptyValues(t *testing.T) {
	if httpx.Cookie(context.Background(), "x") != "" || httpx.UserAgent(context.Background()) != "" {
		t.Fatal("sem pedido no contexto, os valores devem ser vazios")
	}
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	var got string
	httpx.WithRequest(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) { got = httpx.Cookie(r.Context(), "ausente") })).
		ServeHTTP(httptest.NewRecorder(), req)
	if got != "" {
		t.Fatalf("cookie ausente deveria ser vazio, veio %q", got)
	}
}
