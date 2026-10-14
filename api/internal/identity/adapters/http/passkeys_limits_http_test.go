package httpapi_test

import (
	"net/http"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/keven-rdr/wordloop/api/internal/identity/domain"
	"github.com/keven-rdr/wordloop/api/internal/platform/ratelimit"
)

func TestPasskeysAreOffWhenNotConfigured(t *testing.T) {
	f := newPasskeyFixture(t, domain.RoleUser, false, nil)
	for _, path := range []string{"register/begin", "register/finish", "login/begin", "login/finish"} {
		if res := post(f.h, "/api/v1/auth/passkeys/"+path, map[string]any{"credential": map[string]any{}}, csrf); res.StatusCode != http.StatusServiceUnavailable {
			t.Errorf("%s: esperava 503, veio %d", path, res.StatusCode)
		}
	}
}

func TestAuthEndpointsAreRateLimitedPerIPAndOversizedBodiesAreRefused(t *testing.T) {
	limiter := ratelimit.New(60, 3)
	limiter.TrustProxy = true
	f := newPasskeyFixture(t, domain.RoleUser, true, limiter)
	hdr := map[string]string{"X-Requested-With": "XMLHttpRequest", "X-Forwarded-For": "203.0.113.9"}
	for i := range 3 {
		if res := post(f.h, "/api/v1/auth/passkeys/login/begin", nil, hdr); res.StatusCode != http.StatusOK {
			t.Fatalf("pedido %d: %d", i+1, res.StatusCode)
		}
	}
	res := post(f.h, "/api/v1/auth/passkeys/login/begin", nil, hdr)
	if res.StatusCode != http.StatusTooManyRequests || res.Header.Get("Retry-After") == "" {
		t.Fatalf("o quarto pedido deve receber 429 com Retry-After: %d", res.StatusCode)
	}
	other := map[string]string{"X-Requested-With": "XMLHttpRequest", "X-Forwarded-For": "203.0.113.10"}
	huge := []byte(`{"credential":"` + strings.Repeat("a", 70<<10) + `"}`)
	if res := post(f.h, "/api/v1/auth/passkeys/login/finish", huge, other); res.StatusCode != http.StatusBadRequest {
		t.Fatalf("corpo acima de 64 KB: esperava 400, veio %d", res.StatusCode)
	}
}

func TestSpoofedForwardedForDoesNotEscapeTheLimitWhenTheProxyIsNotTrusted(t *testing.T) {
	f := newPasskeyFixture(t, domain.RoleUser, true, ratelimit.New(60, 2))
	for i := range 2 {
		hdr := map[string]string{"X-Requested-With": "XMLHttpRequest", "X-Forwarded-For": "198.51.100." + string(rune('1'+i))}
		if res := post(f.h, "/api/v1/auth/passkeys/login/begin", nil, hdr); res.StatusCode != http.StatusOK {
			t.Fatalf("pedido %d: %d", i+1, res.StatusCode)
		}
	}
	hdr := map[string]string{"X-Requested-With": "XMLHttpRequest", "X-Forwarded-For": "198.51.100.99"}
	if res := post(f.h, "/api/v1/auth/passkeys/login/begin", nil, hdr); res.StatusCode != http.StatusTooManyRequests {
		t.Fatalf("trocar o X-Forwarded-For nao pode dar uma rajada nova: %d", res.StatusCode)
	}
}

func TestPasskeyListAndDeleteStillRequireASessionWhenPasskeysAreOff(t *testing.T) {
	f := newPasskeyFixture(t, domain.RoleUser, false, nil)
	if res := do(f.h, http.MethodGet, "/api/v1/me/passkeys", nil); res.StatusCode != http.StatusUnauthorized {
		t.Fatalf("lista sem sessao: %d", res.StatusCode)
	}
	if res := do(f.h, http.MethodDelete, "/api/v1/me/passkeys/"+uuid.NewString(), csrf); res.StatusCode != http.StatusUnauthorized {
		t.Fatalf("apagar sem sessao: %d", res.StatusCode)
	}
}
