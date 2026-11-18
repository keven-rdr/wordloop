package app_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/keven-rdr/wordloop/api/internal/progress/app"
	"github.com/keven-rdr/wordloop/api/internal/progress/domain"
	"github.com/keven-rdr/wordloop/api/internal/progress/progresstest"
)

var (
	// 15:00 UTC = 12:00 em Sao Paulo: o dia de estudo de hoje e 2026-10-19.
	now  = time.Date(2026, 10, 19, 15, 0, 0, 0, time.UTC)
	user = uuid.MustParse("00000000-0000-0000-0000-0000000000aa")
)

func date(s string) time.Time {
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		panic(err)
	}
	return t
}

func dashboard(s *progresstest.Store) app.GetDashboard {
	return app.GetDashboard{Identity: &progresstest.Identity{Is: domain.Viewer{UserID: user}}, Reader: s, Now: func() time.Time { return now }}
}

func heatmap(s *progresstest.Store) app.GetHeatmap {
	return app.GetHeatmap{Identity: &progresstest.Identity{Is: domain.Viewer{UserID: user}}, Reader: s, Now: func() time.Time { return now }}
}

func TestDashboardBringsTheWholePicture(t *testing.T) {
	s := progresstest.New()
	s.Rows = []domain.Day{
		{Date: date("2026-10-17"), Reviews: 10, Correct: 5},
		{Date: date("2026-10-18"), Reviews: 10, Correct: 9},
		{Date: date("2026-10-19"), Reviews: 20, Correct: 16, NewItems: 4, ActiveMS: 125_000},
		{Date: date("2026-10-01"), Reviews: 50, Correct: 0}, // fora da janela de 7 dias
	}
	s.Due, s.Known = map[int]int{0: 9, 2: 3}, 42
	s.Missed = []domain.Missed{{ItemID: uuid.New(), Term: "house", Gloss: "casa", Misses: 3}}

	d, err := dashboard(s).Run(context.Background(), "pt-BR")
	if err != nil {
		t.Fatal(err)
	}
	if d.Today != date("2026-10-19") || d.Streak != 3 || !d.StudiedToday || d.KnownWords != 42 || len(d.MostMissed) != 1 {
		t.Errorf("painel: %+v", d)
	}
	if d.TodayStats.Reviews != 20 || d.TodayStats.NewItems != 4 || d.TodayStats.ActiveMS != 125_000 {
		t.Errorf("hoje: %+v", d.TodayStats)
	}
	if d.Accuracy7d == nil || *d.Accuracy7d != 30.0/40.0 {
		t.Errorf("30 acertos em 40 respostas nos 7 dias: %v", d.Accuracy7d)
	}
	if len(d.Forecast) != 7 || d.Forecast[0].Due != 9 || d.Forecast[2].Due != 3 || d.Forecast[6].Date != date("2026-10-25") {
		t.Errorf("previsao: %+v", d.Forecast)
	}
}

func TestDashboardOfSomeoneWhoNeverStudied(t *testing.T) {
	d, err := dashboard(progresstest.New()).Run(context.Background(), "pt-BR")
	if err != nil || d.Streak != 0 || d.StudiedToday || d.Accuracy7d != nil || d.KnownWords != 0 || len(d.Forecast) != 7 {
		t.Errorf("painel vazio: %+v %v", d, err)
	}
}

func TestDashboardYesterdayKeepsTheStreakAlive(t *testing.T) {
	s := progresstest.New()
	s.Rows = []domain.Day{{Date: date("2026-10-17"), Reviews: 3, Correct: 3}, {Date: date("2026-10-18"), Reviews: 3, Correct: 3}}
	d, err := dashboard(s).Run(context.Background(), "pt-BR")
	if err != nil || d.Streak != 2 || d.StudiedToday {
		t.Errorf("hoje ainda nao estudou, a sequencia segue: %+v %v", d, err)
	}
}

// A virada do dia segue o fuso e o corte: 02:00 em Sao Paulo ainda e o dia anterior.
func TestDashboardDayRollsOverByTimeZoneAndCutoff(t *testing.T) {
	s := progresstest.New()
	s.Rows = []domain.Day{{Date: date("2026-10-18"), Reviews: 5, Correct: 5}}
	u := dashboard(s)
	u.Now = func() time.Time { return time.Date(2026, 10, 19, 5, 0, 0, 0, time.UTC) } // 02:00 em Sao Paulo
	d, err := u.Run(context.Background(), "pt-BR")
	if err != nil || d.Today != date("2026-10-18") || !d.StudiedToday || d.Streak != 1 {
		t.Errorf("madrugada conta no dia anterior: %+v %v", d, err)
	}
	s.Config = domain.Settings{TZ: "Asia/Tokyo", CutoffHour: 4}
	u.Now = func() time.Time { return time.Date(2026, 10, 19, 20, 0, 0, 0, time.UTC) } // 05:00 do dia 20 em Toquio
	d, err = u.Run(context.Background(), "pt-BR")
	if err != nil || d.Today != date("2026-10-20") || d.StudiedToday {
		t.Errorf("outro fuso, outro dia: %+v %v", d, err)
	}
}

func TestDashboardNeedsASessionAndPropagatesFailures(t *testing.T) {
	u := dashboard(progresstest.New())
	u.Identity = &progresstest.Identity{}
	if _, err := u.Run(context.Background(), "pt-BR"); !errors.Is(err, app.ErrNoSession) {
		t.Errorf("sem sessao: %v", err)
	}
	boom := errors.New("falha simulada")
	for _, method := range []string{"Settings", "Days", "Forecast", "KnownWords", "MostMissed"} {
		s := progresstest.New()
		s.FailOn, s.Err = method, boom
		if _, err := dashboard(s).Run(context.Background(), "pt-BR"); !errors.Is(err, boom) {
			t.Errorf("falha em %s deveria subir: %v", method, err)
		}
	}
	s := progresstest.New()
	s.Config.TZ = "Marte/Olympus"
	if _, err := dashboard(s).Run(context.Background(), "pt-BR"); err == nil {
		t.Error("fuso salvo invalido e erro")
	}
}

func TestHeatmapDefaultsToTwelveWeeksAndListsOnlyStudiedDays(t *testing.T) {
	s := progresstest.New()
	s.Rows = []domain.Day{{Date: date("2026-10-19"), Reviews: 4, Correct: 3}, {Date: date("2026-08-01"), Reviews: 2, Correct: 2}, {Date: date("2026-10-10")}}
	h, err := heatmap(s).Run(context.Background(), nil, nil)
	if err != nil || h.From != date("2026-07-28") || h.To != date("2026-10-19") || len(h.Days) != 2 {
		t.Errorf("84 dias terminando hoje: %+v %v", h, err)
	}
}

func TestHeatmapValidatesTheRange(t *testing.T) {
	from, to := date("2026-10-10"), date("2026-10-01")
	far := date("2025-01-01")
	for name, r := range map[string][2]*time.Time{"invertido": {&from, &to}, "mais de 400 dias": {&far, nil}} {
		if _, err := heatmap(progresstest.New()).Run(context.Background(), r[0], r[1]); !errors.Is(err, app.ErrInvalidRange) {
			t.Errorf("%s: %v", name, err)
		}
	}
	ok, edge := date("2026-10-01"), date("2025-09-14") // exatamente 400 dias ate hoje
	if _, err := heatmap(progresstest.New()).Run(context.Background(), &ok, nil); err != nil {
		t.Errorf("intervalo valido: %v", err)
	}
	if _, err := heatmap(progresstest.New()).Run(context.Background(), &edge, nil); err != nil {
		t.Errorf("400 dias e o limite: %v", err)
	}
}

func TestHeatmapNeedsASession(t *testing.T) {
	u := heatmap(progresstest.New())
	u.Identity = &progresstest.Identity{}
	if _, err := u.Run(context.Background(), nil, nil); !errors.Is(err, app.ErrNoSession) {
		t.Errorf("sem sessao: %v", err)
	}
}
