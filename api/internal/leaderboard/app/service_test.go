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

var (
	me    = uuid.MustParse("00000000-0000-0000-0000-0000000000aa")
	other = uuid.MustParse("00000000-0000-0000-0000-0000000000bb")
	now   = time.Date(2026, 11, 27, 15, 0, 0, 0, time.UTC) // sexta
)

type who struct{ id uuid.UUID }

func (w who) Who(context.Context) (domain.Viewer, error) { return domain.Viewer{UserID: w.id}, nil }

// store e o banco de mentira: guarda os pedidos e devolve as linhas configuradas.
type store struct {
	rows     map[bool][]domain.Entry // por "e o periodo geral?"
	asked    []time.Time
	limits   []int
	profiles map[uuid.UUID]domain.Profile
	met      int
}

func (s *store) Clock(context.Context, uuid.UUID) (string, int, error) {
	return "America/Sao_Paulo", 4, nil
}

func (s *store) Ranking(_ context.Context, _ uuid.UUID, from, _ time.Time, limit int) ([]domain.Entry, error) {
	s.asked, s.limits = append(s.asked, from), append(s.limits, limit)
	return s.rows[from.Year() == 1], nil
}

func (s *store) DaysMet(context.Context, uuid.UUID, time.Time, time.Time) (int, error) {
	return s.met, nil
}

func (s *store) Profile(_ context.Context, u uuid.UUID) (domain.Profile, bool, error) {
	p, ok := s.profiles[u]
	return p, ok, nil
}

func (s *store) SaveProfile(_ context.Context, u uuid.UUID, p domain.Profile) (domain.Profile, error) {
	for id, q := range s.profiles {
		if id != u && q.Nickname == p.Nickname {
			return domain.Profile{}, domain.ErrNicknameTaken
		}
	}
	s.profiles[u] = p
	return p, nil
}

func service(s *store, id uuid.UUID) app.Service {
	return app.Service{Identity: who{id}, Store: s, Now: func() time.Time { return now }}
}

func TestBoardSeparatesTheTopFromMyRowAndUsesMyWeek(t *testing.T) {
	s := &store{rows: map[bool][]domain.Entry{false: {
		{UserID: other, Position: 1, Nickname: "bia", Points: 300},
		{UserID: me, Position: 4, Nickname: "", Points: 40}, // sem participar: so a propria linha
	}}}
	b, err := service(s, me).Board(context.Background(), domain.Week, 3)
	if err != nil {
		t.Fatal(err)
	}
	if b.From != time.Date(2026, 11, 23, 0, 0, 0, 0, time.UTC) || len(b.Entries) != 1 || b.Me == nil || b.Me.Position != 4 {
		t.Errorf("semana de segunda a domingo, podio e a minha linha: %+v", b)
	}
	if s.limits[0] != 3 {
		t.Errorf("limite pedido: %v", s.limits)
	}
}

func TestBoardKeepsMyVisibleRowInThePodiumAndClampsTheLimit(t *testing.T) {
	s := &store{rows: map[bool][]domain.Entry{true: {{UserID: me, Position: 1, Nickname: "ana", Points: 900}}}}
	b, err := service(s, me).Board(context.Background(), domain.All, 999)
	if err != nil || len(b.Entries) != 1 || b.Me == nil || s.limits[0] != app.MaxLimit {
		t.Errorf("participando, apareco no podio e na minha linha; limite no maximo: %+v %v %v", b, s.limits, err)
	}
	if _, err := service(s, me).Board(context.Background(), "month", 10); !errors.Is(err, domain.ErrInvalidPeriod) {
		t.Errorf("periodo desconhecido: %v", err)
	}
}

func TestScoreSumsWeekAndTotalAndComputesTheLevel(t *testing.T) {
	s := &store{met: 3, rows: map[bool][]domain.Entry{
		false: {{UserID: me, Position: 2, Points: 120}},
		true:  {{UserID: me, Position: 5, Points: 260}},
	}}
	sc, err := service(s, me).Score(context.Background())
	if err != nil || sc.WeekPoints != 120 || sc.TotalPoints != 260 || sc.WeekDaysMet != 3 || sc.Level.Number != 3 {
		t.Errorf("resumo: %+v %v", sc, err)
	}
	if s.limits[0] != 0 {
		t.Errorf("o resumo so pede a propria linha: %v", s.limits)
	}
	empty, _ := service(&store{rows: map[bool][]domain.Entry{}}, me).Score(context.Background())
	if empty.TotalPoints != 0 || empty.Level.Number != 1 {
		t.Errorf("sem pontos ainda: %+v", empty)
	}
}

func TestProfileValidatesTheNicknameAndRefusesOneInUse(t *testing.T) {
	s := &store{profiles: map[uuid.UUID]domain.Profile{other: {Nickname: "bia", Visible: true}}}
	svc := service(s, me)
	if p, err := svc.Profile(context.Background()); err != nil || p.Nickname != "" || p.Visible {
		t.Errorf("sem perfil, fora do ranking: %+v %v", p, err)
	}
	if _, err := svc.SaveProfile(context.Background(), domain.Profile{Nickname: "a b"}); !errors.Is(err, domain.ErrInvalidNickname) {
		t.Errorf("apelido invalido: %v", err)
	}
	if _, err := svc.SaveProfile(context.Background(), domain.Profile{Nickname: "bia"}); !errors.Is(err, domain.ErrNicknameTaken) {
		t.Errorf("apelido de outra pessoa: %v", err)
	}
	if p, err := svc.SaveProfile(context.Background(), domain.Profile{Nickname: "ana_1", Visible: true}); err != nil || !p.Visible {
		t.Errorf("salvou: %+v %v", p, err)
	}
}

func TestEverythingNeedsASession(t *testing.T) {
	svc := service(&store{}, uuid.Nil)
	if _, err := svc.Board(context.Background(), domain.Week, 10); !errors.Is(err, app.ErrNoSession) {
		t.Errorf("ranking: %v", err)
	}
	if _, err := svc.Score(context.Background()); !errors.Is(err, app.ErrNoSession) {
		t.Errorf("resumo: %v", err)
	}
	if _, err := svc.SaveProfile(context.Background(), domain.Profile{Nickname: "ana"}); !errors.Is(err, app.ErrNoSession) {
		t.Errorf("perfil: %v", err)
	}
}
