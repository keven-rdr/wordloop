package app_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/keven-rdr/wordloop/api/internal/leaderboard/app"
	"github.com/keven-rdr/wordloop/api/internal/leaderboard/domain"
)

var errBoom = errors.New("falha simulada")

// broken e o banco de mentira que falha no metodo pedido (na n-esima chamada de Ranking, se after > 0).
type broken struct {
	*store
	failOn string
	after  int
	calls  int
}

func (b *broken) fail(method string) error {
	if b.failOn != method {
		return nil
	}
	b.calls++
	if b.calls <= b.after {
		return nil
	}
	return errBoom
}

func (b *broken) Clock(ctx context.Context, u uuid.UUID) (string, int, error) {
	if err := b.fail("Clock"); err != nil {
		return "", 0, err
	}
	return b.store.Clock(ctx, u)
}

func (b *broken) Ranking(ctx context.Context, me uuid.UUID, from, to time.Time, limit int) ([]domain.Entry, error) {
	if err := b.fail("Ranking"); err != nil {
		return nil, err
	}
	return b.store.Ranking(ctx, me, from, to, limit)
}

func (b *broken) DaysMet(ctx context.Context, u uuid.UUID, from, to time.Time) (int, error) {
	if err := b.fail("DaysMet"); err != nil {
		return 0, err
	}
	return b.store.DaysMet(ctx, u, from, to)
}

func (b *broken) Profile(ctx context.Context, u uuid.UUID) (domain.Profile, bool, error) {
	if err := b.fail("Profile"); err != nil {
		return domain.Profile{}, false, err
	}
	return b.store.Profile(ctx, u)
}

func (b *broken) SaveProfile(ctx context.Context, u uuid.UUID, p domain.Profile) (domain.Profile, error) {
	if err := b.fail("SaveProfile"); err != nil {
		return domain.Profile{}, err
	}
	return b.store.SaveProfile(ctx, u, p)
}

func brokenService(failOn string, after int) app.Service {
	s := &store{rows: map[bool][]domain.Entry{}, profiles: map[uuid.UUID]domain.Profile{}}
	return app.Service{Identity: who{me}, Store: &broken{store: s, failOn: failOn, after: after}, Now: func() time.Time { return now }}
}

// Toda falha do banco sobe como erro (vira 500), sem resposta pela metade.
func TestStoreFailuresPropagate(t *testing.T) {
	ctx := context.Background()
	for _, method := range []string{"Clock", "Ranking"} {
		if _, err := brokenService(method, 0).Board(ctx, domain.Week, 10); !errors.Is(err, errBoom) {
			t.Errorf("ranking com falha em %s: %v", method, err)
		}
	}
	cases := []struct {
		method string
		after  int
	}{{"Clock", 0}, {"Ranking", 0}, {"Ranking", 1}, {"DaysMet", 0}} // Ranking 1 = a 2a leitura (o total)
	for _, c := range cases {
		if _, err := brokenService(c.method, c.after).Score(ctx); !errors.Is(err, errBoom) {
			t.Errorf("resumo com falha em %s (%d): %v", c.method, c.after, err)
		}
	}
	if _, err := brokenService("Profile", 0).Profile(ctx); !errors.Is(err, errBoom) {
		t.Errorf("perfil: %v", err)
	}
	if _, err := brokenService("SaveProfile", 0).SaveProfile(ctx, domain.Profile{Nickname: "ana_1"}); !errors.Is(err, errBoom) {
		t.Errorf("salvar perfil: %v", err)
	}
}

type noSession struct{}

func (noSession) Who(context.Context) (domain.Viewer, error) { return domain.Viewer{}, errBoom }

func TestIdentityFailuresPropagate(t *testing.T) {
	svc := app.Service{Identity: noSession{}, Store: &store{}}
	if _, err := svc.Profile(context.Background()); !errors.Is(err, errBoom) {
		t.Errorf("perfil sem identidade: %v", err)
	}
	if _, err := svc.Board(context.Background(), domain.Week, 10); !errors.Is(err, errBoom) {
		t.Errorf("ranking sem identidade: %v", err)
	}
}

func TestTheDefaultClockIsTheRealTime(t *testing.T) {
	svc := app.Service{Identity: who{me}, Store: &store{rows: map[bool][]domain.Entry{}}}
	b, err := svc.Board(context.Background(), domain.Week, 10)
	if err != nil || b.To.Sub(b.From) != 6*24*time.Hour || time.Since(b.From) > 8*24*time.Hour {
		t.Errorf("sem relogio, usa a hora atual: %+v %v", b, err)
	}
}
