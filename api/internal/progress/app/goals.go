package app

import (
	"context"

	"github.com/keven-rdr/wordloop/api/internal/progress/domain"
)

// GoalService cuida das metas por ciclo: ver a atual, sugerir e abrir outra.
type GoalService struct {
	Identity Identity
	Reader   Reader
	Goals    Goals
	Now      Clock
}

// Current e a meta atual com o quanto ja foi cumprido. Sem meta, so Ended = false.
type Current struct {
	Goal     *domain.Goal
	Progress *domain.GoalProgress
	Ended    bool
}

// Current devolve a meta mais recente e o progresso. Acabado o ciclo, o progresso e a retrospectiva e a meta
// continua valendo ate a pessoa abrir outra.
func (u GoalService) Current(ctx context.Context) (Current, error) {
	who, err := resolve(ctx, u.Identity, u.Reader, u.Now)
	if err != nil {
		return Current{}, err
	}
	return u.currentOf(ctx, who)
}

func (u GoalService) currentOf(ctx context.Context, who Asker) (Current, error) {
	goal, found, err := u.Goals.CurrentGoal(ctx, who.Viewer.UserID)
	if err != nil || !found {
		return Current{}, err
	}
	days, err := u.Reader.Days(ctx, who.Viewer.UserID, goal.PeriodStart, goal.PeriodEnd)
	if err != nil {
		return Current{}, err
	}
	progress := goal.Progress(days, who.Today)
	return Current{Goal: &goal, Progress: &progress, Ended: goal.Ended(who.Today)}, nil
}

// Suggest propoe tres metas pelo que a pessoa ja estudou nos ultimos 28 dias.
func (u GoalService) Suggest(ctx context.Context) ([]domain.Suggestion, bool, error) {
	who, err := resolve(ctx, u.Identity, u.Reader, u.Now)
	if err != nil {
		return nil, false, err
	}
	days, err := u.Reader.Days(ctx, who.Viewer.UserID, domain.AddDays(who.Today, -27), who.Today)
	if err != nil {
		return nil, false, err
	}
	suggestions, fromHistory := domain.Suggest(days, who.Today)
	return suggestions, fromHistory, nil
}

// Start abre um ciclo de 7 dias a partir de hoje; o anterior e encerrado, cumprido ou nao.
func (u GoalService) Start(ctx context.Context, req domain.GoalRequest) (Current, error) {
	who, err := resolve(ctx, u.Identity, u.Reader, u.Now)
	if err != nil {
		return Current{}, err
	}
	goal, err := domain.NewGoal(who.Today, req)
	if err != nil {
		return Current{}, err
	}
	if _, err := u.Goals.StartGoal(ctx, who.Viewer.UserID, goal); err != nil {
		return Current{}, err
	}
	return u.currentOf(ctx, who)
}
