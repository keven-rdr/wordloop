package app_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/keven-rdr/wordloop/api/internal/progress/app"
	"github.com/keven-rdr/wordloop/api/internal/progress/domain"
	"github.com/keven-rdr/wordloop/api/internal/progress/progresstest"
)

func goals(s *progresstest.Store) app.GoalService {
	return app.GoalService{
		Identity: &progresstest.Identity{Is: domain.Viewer{UserID: user}}, Reader: s, Goals: s,
		Now: func() time.Time { return now },
	}
}

func open(t *testing.T, u app.GoalService, minutes, days int) app.Current {
	t.Helper()
	cur, err := u.Start(context.Background(), domain.GoalRequest{DailyMinutes: minutes, DaysTarget: days})
	if err != nil {
		t.Fatal(err)
	}
	return cur
}

func TestCurrentGoalIsEmptyUntilThePersonOpensOne(t *testing.T) {
	cur, err := goals(progresstest.New()).Current(context.Background())
	if err != nil || cur.Goal != nil || cur.Progress != nil || cur.Ended {
		t.Errorf("sem meta: %+v %v", cur, err)
	}
}

func TestStartOpensASevenDayCycleFromTodayWithTodaysProgress(t *testing.T) {
	s := progresstest.New()
	s.Rows = []domain.Day{{Date: date("2026-10-19"), Reviews: 6, ActiveMS: 180_000, NewItems: 2}}
	cur := open(t, goals(s), 10, 5)
	if cur.Goal == nil || cur.Goal.PeriodStart != date("2026-10-19") || cur.Goal.PeriodEnd != date("2026-10-25") || cur.Ended {
		t.Fatalf("ciclo: %+v", cur)
	}
	if p := cur.Progress; p == nil || p.DaysStudied != 1 || p.MinutesStudied != 3 || p.NewWords != 2 || p.DaysLeft != 6 || p.MinutesGoal != 50 {
		t.Errorf("o que ja estudou hoje conta: %+v", cur.Progress)
	}
}

func TestStartingAnotherGoalClosesTheOldOneWhetherOrNotItWasMet(t *testing.T) {
	s := progresstest.New()
	older := goals(s)
	older.Now = func() time.Time { return now.AddDate(0, 0, -10) } // a meta antiga comecou ha 10 dias
	open(t, older, 10, 5)
	open(t, goals(s), 20, 3)
	if len(s.Cycles) != 2 || !s.Cycles[0].Closed || s.Cycles[1].Closed {
		t.Fatalf("o antigo fecha, o novo fica aberto: %+v", s.Cycles)
	}
	if cur, _ := goals(s).Current(context.Background()); cur.Goal.DailyMinutes != 20 {
		t.Errorf("vale a mais nova: %+v", cur.Goal)
	}
}

func TestOpeningAGoalTwiceInTheSameDayReplacesIt(t *testing.T) {
	s := progresstest.New()
	open(t, goals(s), 10, 5)
	open(t, goals(s), 15, 4)
	if len(s.Cycles) != 1 || s.Cycles[0].Goal.DailyMinutes != 15 {
		t.Errorf("um ciclo por dia de inicio: %+v", s.Cycles)
	}
}

// Passado o fim, o ciclo mostra a retrospectiva e a meta continua valendo: nao responder nao bloqueia nada.
func TestEndedCycleShowsTheRetrospectiveAndKeepsTheGoalUntilTheNextOne(t *testing.T) {
	s := progresstest.New()
	s.Rows = []domain.Day{
		{Date: date("2026-10-12"), Reviews: 8, ActiveMS: 600_000, NewItems: 3},
		{Date: date("2026-10-13"), Reviews: 4, ActiveMS: 300_000},
	}
	older := goals(s)
	older.Now = func() time.Time { return now.AddDate(0, 0, -7) } // ciclo de 12 a 18/10; hoje e 19/10
	open(t, older, 10, 2)

	cur, err := goals(s).Current(context.Background())
	if err != nil || !cur.Ended || cur.Goal == nil || cur.Goal.DaysTarget != 2 {
		t.Fatalf("o ciclo acabou mas a meta fica: %+v %v", cur, err)
	}
	if p := cur.Progress; p.DaysStudied != 2 || p.MinutesStudied != 15 || p.NewWords != 3 || !p.Met || p.DaysLeft != 0 {
		t.Errorf("retrospectiva: %+v", p)
	}
}

func TestStartRejectsAnInvalidGoalWithoutSavingAnything(t *testing.T) {
	s := progresstest.New()
	_, err := goals(s).Start(context.Background(), domain.GoalRequest{DailyMinutes: 500, DaysTarget: 5})
	if !errors.Is(err, domain.ErrInvalidGoal) || len(s.Cycles) != 0 {
		t.Errorf("meta invalida: %v %+v", err, s.Cycles)
	}
}

func TestSuggestUsesTheLastFourWeeks(t *testing.T) {
	s := progresstest.New()
	for i := 0; i < 8; i++ {
		s.Rows = append(s.Rows, domain.Day{Date: date("2026-10-19").AddDate(0, 0, -i*3), Reviews: 5, ActiveMS: 600_000})
	}
	got, fromHistory, err := goals(s).Suggest(context.Background())
	if err != nil || !fromHistory || len(got) != 3 || got[1].DailyMinutes != 10 || got[1].DaysTarget != 2 {
		t.Errorf("sugestoes: %v %v %v", got, fromHistory, err)
	}
	_, fromHistory, err = goals(progresstest.New()).Suggest(context.Background())
	if err != nil || fromHistory {
		t.Errorf("sem historico: %v %v", fromHistory, err)
	}
}

func TestGoalsNeedASessionAndPropagateFailures(t *testing.T) {
	u := goals(progresstest.New())
	u.Identity = &progresstest.Identity{}
	if _, err := u.Current(context.Background()); !errors.Is(err, app.ErrNoSession) {
		t.Errorf("sem sessao: %v", err)
	}
	boom := errors.New("falha simulada")
	for _, method := range []string{"Settings", "CurrentGoal", "Days", "StartGoal"} {
		s := progresstest.New()
		s.FailOn, s.Err = method, boom
		u := goals(s)
		var err error
		switch method {
		case "StartGoal":
			_, err = u.Start(context.Background(), domain.GoalRequest{DailyMinutes: 10, DaysTarget: 3})
		case "Days":
			s.Cycles = []progresstest.Cycle{{Goal: domain.Goal{PeriodStart: date("2026-10-19"), PeriodEnd: date("2026-10-25")}}}
			_, err = u.Current(context.Background())
		default:
			_, err = u.Current(context.Background())
		}
		if !errors.Is(err, boom) {
			t.Errorf("falha em %s deveria subir: %v", method, err)
		}
	}
}
