package domain_test

import (
	"testing"
	"time"

	"github.com/keven-rdr/wordloop/api/internal/identity/domain"
)

func TestLifetimeExpired(t *testing.T) {
	start := time.Date(2026, 10, 14, 9, 0, 0, 0, time.UTC)
	l := domain.Lifetime{Absolute: 8 * time.Hour, Idle: 30 * time.Minute}
	cases := map[string]struct {
		lastUsed time.Time
		now      time.Time
		want     bool
	}{
		"em uso, dentro dos prazos":     {start.Add(10 * time.Minute), start.Add(20 * time.Minute), false},
		"um segundo antes do ocioso":    {start, start.Add(30*time.Minute - time.Second), false},
		"ocioso":                        {start, start.Add(30 * time.Minute), true},
		"em uso mas passou do absoluto": {start.Add(7*time.Hour + 50*time.Minute), start.Add(8 * time.Hour), true},
	}
	for name, c := range cases {
		if got := l.Expired(start, c.lastUsed, c.now); got != c.want {
			t.Errorf("%s: Expired = %v, esperava %v", name, got, c.want)
		}
	}
}

func TestPolicyAdminIsShorterThanUser(t *testing.T) {
	p := domain.DefaultSessionPolicy()
	if p.For(domain.RoleAdmin).Absolute >= p.For(domain.RoleUser).Absolute || p.For(domain.RoleAdmin).Idle >= p.For(domain.RoleUser).Idle {
		t.Fatal("a sessao de admin precisa ser mais curta que a do usuario")
	}
	if p.For(domain.Role("qualquer")) != p.User {
		t.Fatal("papel desconhecido deve receber o prazo do usuario comum")
	}
}

func TestParseRoleFailsSafe(t *testing.T) {
	cases := map[string]domain.Role{"admin": domain.RoleAdmin, "user": domain.RoleUser, "": domain.RoleUser, "ADMIN": domain.RoleUser, "root": domain.RoleUser}
	for in, want := range cases {
		if got := domain.ParseRole(in); got != want {
			t.Errorf("ParseRole(%q) = %q, esperava %q", in, got, want)
		}
	}
}

func TestOrDefaultFillsOnlyTheMissingPieces(t *testing.T) {
	d := domain.DefaultSessionPolicy()
	if got := (domain.SessionPolicy{}).OrDefault(); got != d {
		t.Fatalf("politica vazia deve virar o padrao: %+v", got)
	}
	partial := domain.SessionPolicy{User: domain.Lifetime{Absolute: 24 * time.Hour}} // Idle esquecido; Admin vazio
	got := partial.OrDefault()
	if got.User.Absolute != 24*time.Hour || got.User.Idle != d.User.Idle || got.Admin != d.Admin {
		t.Fatalf("OrDefault deve manter o que foi definido e completar o resto: %+v", got)
	}
	if got.User.Expired(time.Now(), time.Now(), time.Now()) {
		t.Fatal("uma sessao recem-criada nao pode nascer vencida")
	}
}
