package ratelimit_test

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/keven-rdr/wordloop/api/internal/platform/ratelimit"
)

func TestAllowSpendsTheBurstThenRefillsOverTime(t *testing.T) {
	clock := time.Date(2026, 10, 15, 12, 0, 0, 0, time.UTC)
	l := ratelimit.New(60, 3)
	l.Now = func() time.Time { return clock }
	for i := range 3 {
		if ok, _ := l.Allow("a"); !ok {
			t.Fatalf("pedido %d dentro da rajada deve passar", i+1)
		}
	}
	ok, wait := l.Allow("a")
	if ok || wait < 900*time.Millisecond || wait > 1100*time.Millisecond {
		t.Fatalf("o quarto pedido deve ser barrado e esperar ~1s: ok=%v wait=%v", ok, wait)
	}
	if other, _ := l.Allow("b"); !other {
		t.Fatal("outra chave tem a propria rajada")
	}
	clock = clock.Add(2 * time.Second)
	if ok, _ := l.Allow("a"); !ok {
		t.Fatal("depois de 2s ja ha fichas de novo")
	}
}

func TestMiddlewareReturns429WithRetryAfterOnlyUnderThePrefix(t *testing.T) {
	l := ratelimit.New(60, 1)
	l.TrustProxy = true
	h := l.Middleware("/api/v1/auth/", http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusNoContent) }))
	do := func(path, ip string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodPost, path, nil)
		req.Header.Set("X-Forwarded-For", ip)
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		return rec
	}
	if rec := do("/api/v1/auth/login", "203.0.113.5"); rec.Code != http.StatusNoContent {
		t.Fatalf("primeiro pedido: %d", rec.Code)
	}
	rec := do("/api/v1/auth/login", "203.0.113.5")
	if secs, err := strconv.Atoi(rec.Header().Get("Retry-After")); rec.Code != http.StatusTooManyRequests || err != nil || secs < 1 {
		t.Fatalf("segundo pedido: %d Retry-After=%q", rec.Code, rec.Header().Get("Retry-After"))
	}
	if rec := do("/api/v1/version", "203.0.113.5"); rec.Code != http.StatusNoContent {
		t.Fatalf("rotas fora do prefixo nao entram no limite: %d", rec.Code)
	}
	if rec := do("/api/v1/auth/login", "203.0.113.6"); rec.Code != http.StatusNoContent {
		t.Fatalf("outro IP tem o proprio limite: %d", rec.Code)
	}
}

func TestClientIPUsesTheLastForwardedEntryAndFallsBackToTheSocket(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "10.0.0.9:5555"
	if got := ratelimit.ClientIP(req, true); got != "10.0.0.9" {
		t.Fatalf("sem proxy: %q", got)
	}
	req.Header.Set("X-Forwarded-For", "1.2.3.4, 198.51.100.7")
	if got := ratelimit.ClientIP(req, true); got != "198.51.100.7" {
		t.Fatalf("o ultimo valor (posto pelo Caddy) e o que vale, nao o que o cliente mandou: %q", got)
	}
	if got := ratelimit.ClientIP(req, false); got != "10.0.0.9" {
		t.Fatalf("sem confiar no proxy, o cabecalho e ignorado (nao da para burlar o limite): %q", got)
	}
	req.Header.Set("X-Forwarded-For", "lixo")
	if got := ratelimit.ClientIP(req, true); got != "10.0.0.9" {
		t.Fatalf("XFF invalido cai para o socket: %q", got)
	}
}

func TestOldBucketsAreSweptWhenTheMapIsFull(t *testing.T) {
	clock := time.Date(2026, 10, 15, 12, 0, 0, 0, time.UTC)
	l := ratelimit.New(60, 1)
	l.Now = func() time.Time { return clock }
	for i := range 10000 {
		l.Allow("ip-" + strconv.Itoa(i))
	}
	clock = clock.Add(11 * time.Minute)
	if ok, _ := l.Allow("novo"); !ok {
		t.Fatal("o mapa cheio deve limpar entradas antigas e aceitar a nova")
	}
}
