package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/keven-rdr/wordloop/api/internal/progress/adapters/postgres/db"
	"github.com/keven-rdr/wordloop/api/internal/progress/app"
	"github.com/keven-rdr/wordloop/api/internal/progress/domain"
)

var _ app.Goals = (*Store)(nil)

// goal monta a meta a partir dos campos de uma linha de goal_cycle (as consultas do sqlc devolvem tipos parecidos).
func goal(id uuid.UUID, start, end pgtype.Date, preset pgtype.Text, daily, days int16, weekdays []int16) domain.Goal {
	g := domain.Goal{
		ID: id, PeriodStart: start.Time, PeriodEnd: end.Time, Preset: domain.Preset(preset.String),
		DailyMinutes: int(daily), DaysTarget: int(days), Weekdays: make([]int, len(weekdays)),
	}
	for i, d := range weekdays {
		g.Weekdays[i] = int(d)
	}
	return g
}

func (s *Store) CurrentGoal(ctx context.Context, user uuid.UUID) (domain.Goal, bool, error) {
	r, err := s.q.CurrentGoal(ctx, user)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Goal{}, false, nil
	}
	if err != nil {
		return domain.Goal{}, false, fmt.Errorf("ler meta: %w", err)
	}
	return goal(r.ID, r.PeriodStart, r.PeriodEnd, r.Preset, r.DailyMinutes, r.DaysTarget, r.Weekdays), true, nil
}

// StartGoal fecha os ciclos abertos e grava o novo numa transacao so: nunca ficam dois ciclos abertos.
func (s *Store) StartGoal(ctx context.Context, user uuid.UUID, g domain.Goal) (domain.Goal, error) {
	var saved db.UpsertGoalRow
	err := pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		q := db.New(tx)
		if err := q.CloseOlderGoals(ctx, db.CloseOlderGoalsParams{UserID: user, PeriodStart: date(g.PeriodStart)}); err != nil {
			return err
		}
		weekdays := make([]int16, len(g.Weekdays))
		for i, d := range g.Weekdays {
			weekdays[i] = int16(d) //nolint:gosec // 1 a 7, validado no dominio
		}
		var err error
		saved, err = q.UpsertGoal(ctx, db.UpsertGoalParams{
			UserID: user, PeriodStart: date(g.PeriodStart), PeriodEnd: date(g.PeriodEnd),
			Preset:       pgtype.Text{String: string(g.Preset), Valid: g.Preset != ""},
			DailyMinutes: int16(g.DailyMinutes), DaysTarget: int16(g.DaysTarget), Weekdays: weekdays, //nolint:gosec // 1 a 240 e 1 a 7
		})
		return err
	})
	if err != nil {
		return domain.Goal{}, fmt.Errorf("abrir meta: %w", err)
	}
	return goal(saved.ID, saved.PeriodStart, saved.PeriodEnd, saved.Preset, saved.DailyMinutes, saved.DaysTarget, saved.Weekdays), nil
}
