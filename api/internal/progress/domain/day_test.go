package domain_test

import (
	"testing"
	"time"

	"github.com/keven-rdr/wordloop/api/internal/progress/domain"
)

func date(s string) time.Time {
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		panic(err)
	}
	return t
}

func studied(dates ...string) []domain.Day {
	out := make([]domain.Day, len(dates))
	for i, d := range dates {
		out[i] = domain.Day{Date: date(d), Reviews: 5, Correct: 4}
	}
	return out
}

func TestTodayFollowsTheTimeZoneAndTheCutoff(t *testing.T) {
	cases := map[string]struct {
		now  time.Time
		tz   string
		want string
	}{
		"tarde em Sao Paulo":                    {time.Date(2026, 10, 19, 18, 0, 0, 0, time.UTC), "America/Sao_Paulo", "2026-10-19"},
		"madrugada (02:00 local) e dia antes":   {time.Date(2026, 10, 19, 5, 0, 0, 0, time.UTC), "America/Sao_Paulo", "2026-10-18"},
		"depois do corte (04:30 local)":         {time.Date(2026, 10, 19, 7, 30, 0, 0, time.UTC), "America/Sao_Paulo", "2026-10-19"},
		"Toquio, 18:00 UTC e 03:00 do dia 20":   {time.Date(2026, 10, 19, 18, 0, 0, 0, time.UTC), "Asia/Tokyo", "2026-10-19"},
		"Toquio, 20:00 UTC e 05:00 do dia 20":   {time.Date(2026, 10, 19, 20, 0, 0, 0, time.UTC), "Asia/Tokyo", "2026-10-20"},
		"Lisboa, madrugada no horario de verao": {time.Date(2026, 7, 1, 2, 0, 0, 0, time.UTC), "Europe/Lisbon", "2026-06-30"},
	}
	for name, c := range cases {
		got, err := domain.Today(c.now, c.tz, 4)
		if err != nil || got.Format("2006-01-02") != c.want {
			t.Errorf("%s: %v %v, esperava %s", name, got, err, c.want)
		}
	}
	if _, err := domain.Today(time.Now(), "Marte/Olympus", 4); err == nil {
		t.Error("fuso desconhecido e erro")
	}
	if _, err := domain.Today(time.Now(), "", 4); err == nil {
		t.Error("fuso vazio e erro")
	}
}

func TestStreakCountsConsecutiveStudiedDaysUpToToday(t *testing.T) {
	today := date("2026-10-19")
	cases := map[string]struct {
		days []domain.Day
		want int
	}{
		"sem estudo":                       {nil, 0},
		"so hoje":                          {studied("2026-10-19"), 1},
		"3 dias seguidos ate hoje":         {studied("2026-10-17", "2026-10-18", "2026-10-19"), 3},
		"hoje ainda nao: ontem segura":     {studied("2026-10-17", "2026-10-18"), 2},
		"um dia inteiro sem estudo quebra": {studied("2026-10-15", "2026-10-16", "2026-10-17"), 0},
		"buraco no meio zera o que passou": {studied("2026-10-14", "2026-10-15", "2026-10-18", "2026-10-19"), 2},
		"dia sem resposta nao conta":       {append(studied("2026-10-17"), domain.Day{Date: date("2026-10-18")}, domain.Day{Date: today, Reviews: 1}), 1},
		"fora de ordem":                    {studied("2026-10-19", "2026-10-17", "2026-10-18"), 3},
	}
	for name, c := range cases {
		if got := domain.Streak(c.days, today); got != c.want {
			t.Errorf("%s: %d, esperava %d", name, got, c.want)
		}
	}
}

func TestAccuracyIsNilWithoutAnswers(t *testing.T) {
	if domain.Accuracy(nil) != nil || domain.Accuracy([]domain.Day{{Date: date("2026-10-19")}}) != nil {
		t.Error("sem respostas nao ha taxa")
	}
	got := domain.Accuracy([]domain.Day{{Reviews: 10, Correct: 8}, {Reviews: 10, Correct: 6}})
	if got == nil || *got != 0.7 {
		t.Errorf("14 de 20: %v", got)
	}
}

func TestWindowKeepsOnlyTheDaysInsideTheRange(t *testing.T) {
	days := studied("2026-10-10", "2026-10-13", "2026-10-19", "2026-10-20")
	got := domain.Window(days, date("2026-10-13"), date("2026-10-19"))
	if len(got) != 2 || got[0].Date != date("2026-10-13") || got[1].Date != date("2026-10-19") {
		t.Errorf("janela inclusiva nas duas pontas: %+v", got)
	}
}

func TestForecastFillsEveryDay(t *testing.T) {
	got := domain.Forecast(date("2026-10-19"), map[int]int{0: 12, 3: 4}, 7)
	if len(got) != 7 || got[0].Due != 12 || got[3].Due != 4 || got[1].Due != 0 || got[6].Date != date("2026-10-25") {
		t.Errorf("previsao: %+v", got)
	}
}
