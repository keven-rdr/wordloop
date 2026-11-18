package domain_test

import (
	"errors"
	"testing"

	"github.com/keven-rdr/wordloop/api/internal/progress/domain"
)

func TestNewGoalOpensASevenDayCycleAndAppliesTheDefaults(t *testing.T) {
	start := date("2026-10-19")
	g, err := domain.NewGoal(start, domain.GoalRequest{DailyMinutes: 10, DaysTarget: 5})
	if err != nil {
		t.Fatal(err)
	}
	if g.PeriodStart != start || g.PeriodEnd != date("2026-10-25") || g.Preset != domain.Custom || len(g.Weekdays) != 7 {
		t.Errorf("meta: %+v", g)
	}
}

func TestNewGoalSortsTheWeekdaysWithoutTouchingTheRequest(t *testing.T) {
	asked := []int{5, 1, 3}
	g, err := domain.NewGoal(date("2026-10-19"), domain.GoalRequest{Preset: domain.Steady, DailyMinutes: 10, DaysTarget: 3, Weekdays: asked})
	if err != nil || g.Weekdays[0] != 1 || g.Weekdays[2] != 5 || asked[0] != 5 {
		t.Errorf("ordenou o pedido do chamador ou nao ordenou a meta: %+v %v %v", g.Weekdays, asked, err)
	}
}

func TestNewGoalRejectsWhatIsOutOfBounds(t *testing.T) {
	cases := map[string]domain.GoalRequest{
		"0 minutos":                      {DailyMinutes: 0, DaysTarget: 3},
		"241 minutos":                    {DailyMinutes: 241, DaysTarget: 3},
		"0 dias":                         {DailyMinutes: 10, DaysTarget: 0},
		"8 dias":                         {DailyMinutes: 10, DaysTarget: 8},
		"preset desconhecido":            {Preset: "turbo", DailyMinutes: 10, DaysTarget: 3},
		"dia da semana 0":                {DailyMinutes: 10, DaysTarget: 1, Weekdays: []int{0, 1}},
		"dia da semana 8":                {DailyMinutes: 10, DaysTarget: 1, Weekdays: []int{8}},
		"dia da semana repetido":         {DailyMinutes: 10, DaysTarget: 2, Weekdays: []int{1, 1, 2}},
		"menos dias permitidos que meta": {DailyMinutes: 10, DaysTarget: 4, Weekdays: []int{1, 2, 3}},
	}
	for name, req := range cases {
		if _, err := domain.NewGoal(date("2026-10-19"), req); !errors.Is(err, domain.ErrInvalidGoal) {
			t.Errorf("%s: %v", name, err)
		}
	}
	for _, req := range []domain.GoalRequest{{DailyMinutes: 1, DaysTarget: 1}, {DailyMinutes: 240, DaysTarget: 7}} {
		if _, err := domain.NewGoal(date("2026-10-19"), req); err != nil {
			t.Errorf("limite aceito: %+v %v", req, err)
		}
	}
}

func goalOf(t *testing.T, days int) domain.Goal {
	t.Helper()
	g, err := domain.NewGoal(date("2026-10-19"), domain.GoalRequest{DailyMinutes: 10, DaysTarget: days})
	if err != nil {
		t.Fatal(err)
	}
	return g
}

func TestProgressCountsStudiedDaysMinutesAndNewWordsOfTheCycle(t *testing.T) {
	days := []domain.Day{
		{Date: date("2026-10-18"), Reviews: 9, ActiveMS: 999_000, NewItems: 9}, // antes do ciclo: nao conta
		{Date: date("2026-10-19"), Reviews: 10, ActiveMS: 330_000, NewItems: 4},
		{Date: date("2026-10-21"), Reviews: 5, ActiveMS: 150_000, NewItems: 1},
		{Date: date("2026-10-26"), Reviews: 7, ActiveMS: 60_000}, // depois do ciclo: nao conta
	}
	p := goalOf(t, 3).Progress(days, date("2026-10-22"))
	if p.DaysStudied != 2 || p.MinutesStudied != 8 || p.NewWords != 5 || p.MinutesGoal != 30 || p.DaysTarget != 3 || p.Met {
		t.Errorf("progresso: %+v", p)
	}
	if p.DaysLeft != 3 {
		t.Errorf("de 22 para 25/10 faltam 3 dias: %d", p.DaysLeft)
	}
}

func TestProgressIsMetWhenTheStudiedDaysReachTheTarget(t *testing.T) {
	days := []domain.Day{{Date: date("2026-10-19"), Reviews: 1}, {Date: date("2026-10-20"), Reviews: 1}}
	if p := goalOf(t, 2).Progress(days, date("2026-10-20")); !p.Met {
		t.Errorf("2 de 2: %+v", p)
	}
	if p := goalOf(t, 3).Progress(days, date("2026-10-20")); p.Met {
		t.Errorf("2 de 3: %+v", p)
	}
}

func TestEndedAndDaysLeftAroundTheLastDay(t *testing.T) {
	g := goalOf(t, 3) // 19 a 25/10
	if g.Ended(date("2026-10-25")) || !g.Ended(date("2026-10-26")) {
		t.Error("acaba depois do ultimo dia, nao nele")
	}
	if got := g.Progress(nil, date("2026-10-25")).DaysLeft; got != 0 {
		t.Errorf("no ultimo dia faltam 0: %d", got)
	}
	if got := g.Progress(nil, date("2026-11-30")).DaysLeft; got != 0 {
		t.Errorf("depois do fim nao fica negativo: %d", got)
	}
	if got := g.Progress(nil, date("2026-10-19")).DaysLeft; got != 6 {
		t.Errorf("no primeiro dia faltam 6: %d", got)
	}
}

func TestRetrospectiveOnlyCountsTheDaysOfTheCycleEvenLongAfterIt(t *testing.T) {
	days := []domain.Day{{Date: date("2026-10-20"), Reviews: 3, ActiveMS: 120_000}, {Date: date("2026-11-10"), Reviews: 50, ActiveMS: 9_000_000}}
	p := goalOf(t, 1).Progress(days, date("2026-11-12"))
	if p.DaysStudied != 1 || p.MinutesStudied != 2 || !p.Met {
		t.Errorf("retrospectiva: %+v", p)
	}
}

// studyDays gera n dias estudados, um por dia, terminando em end, com os minutos pedidos.
func studyDays(end string, minutes []float64) []domain.Day {
	out := make([]domain.Day, len(minutes))
	for i, m := range minutes {
		out[i] = domain.Day{Date: date(end).AddDate(0, 0, -i*2), Reviews: 5, ActiveMS: int64(m * 60_000)}
	}
	return out
}

func TestSuggestFallsBackToFixedValuesWithoutEnoughHistory(t *testing.T) {
	got, fromHistory := domain.Suggest(studyDays("2026-10-19", []float64{10, 12}), date("2026-10-19"))
	want := []domain.Suggestion{{Preset: domain.Light, DailyMinutes: 5, DaysTarget: 3}, {Preset: domain.Steady, DailyMinutes: 10, DaysTarget: 5}, {Preset: domain.Intense, DailyMinutes: 20, DaysTarget: 7}}
	if fromHistory || len(got) != 3 || got[0] != want[0] || got[1] != want[1] || got[2] != want[2] {
		t.Errorf("valores iniciais: %v %v", got, fromHistory)
	}
}

func TestSuggestFollowsTheRealTimeAndTheRealFrequency(t *testing.T) {
	// 12 dias estudados em 28: 3 por semana; mediana 10 min, percentil 75 = 15 min.
	minutes := []float64{5, 8, 8, 9, 10, 10, 10, 12, 14, 15, 18, 25}
	got, fromHistory := domain.Suggest(studyDays("2026-10-19", minutes), date("2026-10-19"))
	if !fromHistory || len(got) != 3 {
		t.Fatalf("historico: %v %v", got, fromHistory)
	}
	light, steady, intense := got[0], got[1], got[2]
	if light.Preset != domain.Light || steady.Preset != domain.Steady || intense.Preset != domain.Intense {
		t.Errorf("ordem: %v", got)
	}
	if light.DailyMinutes != 10 || steady.DailyMinutes != 10 || intense.DailyMinutes != 15 {
		t.Errorf("minutos 0,8x / 1x / p75 (de 5 em 5): %v", got)
	}
	if light.DaysTarget != 2 || steady.DaysTarget != 3 || intense.DaysTarget != 4 {
		t.Errorf("dias por semana ao redor de 3: %v", got)
	}
}

func TestSuggestStaysInsideTheLimitsAndNeverGetsLighterThanLessStudy(t *testing.T) {
	heavy := make([]float64, 28)
	for i := range heavy {
		heavy[i] = 400 // 6h40 por dia
	}
	days := make([]domain.Day, 28)
	for i := range days {
		days[i] = domain.Day{Date: date("2026-10-19").AddDate(0, 0, -i), Reviews: 3, ActiveMS: int64(heavy[i] * 60_000)}
	}
	got, _ := domain.Suggest(days, date("2026-10-19"))
	for _, s := range got {
		if s.DailyMinutes > 240 || s.DaysTarget > 7 || s.DailyMinutes < 5 || s.DaysTarget < 1 {
			t.Errorf("fora dos limites: %+v", s)
		}
		if _, err := domain.NewGoal(date("2026-10-19"), domain.GoalRequest{Preset: s.Preset, DailyMinutes: s.DailyMinutes, DaysTarget: s.DaysTarget}); err != nil {
			t.Errorf("a sugestao tem que ser uma meta valida: %+v %v", s, err)
		}
	}
	if got[0].DailyMinutes > got[1].DailyMinutes || got[1].DailyMinutes > got[2].DailyMinutes {
		t.Errorf("leve <= constante <= intensa: %v", got)
	}
}

func TestSuggestIgnoresDaysOlderThanFourWeeks(t *testing.T) {
	old := []domain.Day{{Date: date("2026-08-01"), Reviews: 5, ActiveMS: 3_000_000}, {Date: date("2026-08-02"), Reviews: 5, ActiveMS: 3_000_000}, {Date: date("2026-08-03"), Reviews: 5, ActiveMS: 3_000_000}}
	if _, fromHistory := domain.Suggest(old, date("2026-10-19")); fromHistory {
		t.Error("estudo de meses atras nao conta como historico")
	}
}
