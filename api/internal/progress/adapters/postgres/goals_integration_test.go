//go:build integration

package postgres_test

import (
	"context"
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/keven-rdr/wordloop/api/internal/progress/domain"
)

func (f fixture) goal(t *testing.T, user uuid.UUID, start string, minutes, days int) domain.Goal {
	t.Helper()
	g, err := domain.NewGoal(day(start), domain.GoalRequest{Preset: domain.Steady, DailyMinutes: minutes, DaysTarget: days})
	if err != nil {
		t.Fatal(err)
	}
	saved, err := f.store.StartGoal(context.Background(), user, g)
	if err != nil {
		t.Fatal(err)
	}
	return saved
}

func (f fixture) statuses(t *testing.T, user uuid.UUID) []string {
	t.Helper()
	rows, err := f.pool.Query(context.Background(), `SELECT status FROM goal_cycle WHERE user_id = $1 ORDER BY period_start`, user)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var s string
		if err := rows.Scan(&s); err != nil {
			t.Fatal(err)
		}
		out = append(out, s)
	}
	return out
}

func TestGoalIsSavedAndReadBackForTheWholeWeek(t *testing.T) {
	f := setup(t)
	ctx := context.Background()
	if _, found, err := f.store.CurrentGoal(ctx, f.user); err != nil || found {
		t.Fatalf("sem meta: %v %v", found, err)
	}
	saved := f.goal(t, f.user, "2026-10-19", 10, 5)
	got, found, err := f.store.CurrentGoal(ctx, f.user)
	if err != nil || !found || got.ID != saved.ID {
		t.Fatalf("meta: %+v %v %v", got, found, err)
	}
	if got.PeriodStart != day("2026-10-19") || got.PeriodEnd != day("2026-10-25") || got.Preset != domain.Steady ||
		got.DailyMinutes != 10 || got.DaysTarget != 5 || !slices.Equal(got.Weekdays, []int{1, 2, 3, 4, 5, 6, 7}) {
		t.Errorf("campos: %+v", got)
	}
}

func TestOpeningANewCycleClosesTheOlderOnesAndTheNewestIsTheCurrent(t *testing.T) {
	f := setup(t)
	f.goal(t, f.user, "2026-10-05", 10, 5)
	f.goal(t, f.user, "2026-10-12", 15, 4)
	newest := f.goal(t, f.user, "2026-10-19", 20, 3)

	if got := f.statuses(t, f.user); !slices.Equal(got, []string{"closed", "closed", "active"}) {
		t.Errorf("so o mais novo fica aberto: %v", got)
	}
	got, _, err := f.store.CurrentGoal(context.Background(), f.user)
	if err != nil || got.ID != newest.ID {
		t.Errorf("vale o ciclo mais novo: %+v %v", got, err)
	}
}

func TestOpeningACycleOnTheSameDayReplacesIt(t *testing.T) {
	f := setup(t)
	first := f.goal(t, f.user, "2026-10-19", 10, 5)
	second := f.goal(t, f.user, "2026-10-19", 25, 2)
	var n int
	if err := f.pool.QueryRow(context.Background(), `SELECT count(*) FROM goal_cycle WHERE user_id = $1`, f.user).Scan(&n); err != nil || n != 1 {
		t.Fatalf("um ciclo por dia de inicio: %d %v", n, err)
	}
	if second.ID != first.ID || second.DailyMinutes != 25 || second.DaysTarget != 2 {
		t.Errorf("substituiu os valores no mesmo ciclo: %+v -> %+v", first, second)
	}
}

func TestTheNewestStartWinsEvenIfItWasWrittenFirst(t *testing.T) {
	f := setup(t)
	f.exec(t, `INSERT INTO goal_cycle (user_id, period_start, period_end, daily_minutes, days_target, status)
		VALUES ($1, '2026-10-19', '2026-10-25', 20, 3, 'active'), ($1, '2026-10-05', '2026-10-11', 5, 1, 'closed')`, f.user)
	got, found, err := f.store.CurrentGoal(context.Background(), f.user)
	if err != nil || !found || got.DailyMinutes != 20 {
		t.Errorf("o mais recente por inicio: %+v %v %v", got, found, err)
	}
}

// Um ciclo antigo de 7 dias corridos (aberto numa quarta) comeca depois da segunda da meta nova: a meta da semana
// o fecha e passa a ser a atual.
func TestTheWeeklyGoalReplacesAnOldRollingCycleThatStartedLaterInTheWeek(t *testing.T) {
	f := setup(t)
	f.exec(t, `INSERT INTO goal_cycle (user_id, period_start, period_end, daily_minutes, days_target, status)
		VALUES ($1, '2026-10-21', '2026-10-27', 10, 3, 'active')`, f.user)
	weekly := f.goal(t, f.user, "2026-10-23", 30, 7) // sexta: a meta vale de segunda (19) a domingo (25)
	if weekly.PeriodStart != day("2026-10-19") {
		t.Fatalf("a meta comeca na segunda: %+v", weekly)
	}
	if got := f.statuses(t, f.user); !slices.Equal(got, []string{"active", "closed"}) {
		t.Errorf("o ciclo antigo fecha: %v", got)
	}
	got, _, err := f.store.CurrentGoal(context.Background(), f.user)
	if err != nil || got.ID != weekly.ID || got.DailyMinutes != 30 {
		t.Errorf("vale a meta da semana: %+v %v", got, err)
	}
}

func TestGoalsOfOnePersonDoNotLeakToAnother(t *testing.T) {
	f := setup(t)
	var other uuid.UUID
	if err := f.pool.QueryRow(context.Background(), `INSERT INTO app_user (display_name) VALUES ('Bia') RETURNING id`).Scan(&other); err != nil {
		t.Fatal(err)
	}
	f.goal(t, f.user, "2026-10-19", 10, 5)
	f.goal(t, other, "2026-10-19", 30, 2) // mesmo dia de inicio: nao conflita entre pessoas
	if got, _, _ := f.store.CurrentGoal(context.Background(), f.user); got.DailyMinutes != 10 {
		t.Errorf("a meta de uma nao vira a da outra: %+v", got)
	}
	if got := f.statuses(t, other); !slices.Equal(got, []string{"active"}) {
		t.Errorf("abrir meta de uma pessoa nao fecha a da outra: %v", got)
	}
}

func TestTheDatabaseRefusesGoalsOutOfBounds(t *testing.T) {
	f := setup(t)
	for name, sql := range map[string]string{
		"minutos 0":     `INSERT INTO goal_cycle (user_id, period_start, period_end, daily_minutes, days_target) VALUES ($1, '2026-10-19', '2026-10-25', 0, 3)`,
		"minutos 241":   `INSERT INTO goal_cycle (user_id, period_start, period_end, daily_minutes, days_target) VALUES ($1, '2026-10-19', '2026-10-25', 241, 3)`,
		"dias 8":        `INSERT INTO goal_cycle (user_id, period_start, period_end, daily_minutes, days_target) VALUES ($1, '2026-10-19', '2026-10-25', 10, 8)`,
		"fim antes":     `INSERT INTO goal_cycle (user_id, period_start, period_end, daily_minutes, days_target) VALUES ($1, '2026-10-19', '2026-10-18', 10, 3)`,
		"preset errado": `INSERT INTO goal_cycle (user_id, period_start, period_end, daily_minutes, days_target, preset) VALUES ($1, '2026-10-19', '2026-10-25', 10, 3, 'turbo')`,
	} {
		if _, err := f.pool.Exec(context.Background(), sql, f.user); err == nil {
			t.Errorf("%s: o banco deveria recusar", name)
		}
	}
}
