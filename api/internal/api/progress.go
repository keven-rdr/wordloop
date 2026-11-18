package api

import (
	"context"

	"github.com/keven-rdr/wordloop/api/internal/api/gen"
)

// Progress e o modulo de progresso (painel e heatmap, so leitura); nulo = progresso indisponivel.
type Progress interface {
	GetDashboard(ctx context.Context, req gen.GetDashboardRequestObject) (gen.GetDashboardResponseObject, error)
	GetHeatmap(ctx context.Context, req gen.GetHeatmapRequestObject) (gen.GetHeatmapResponseObject, error)
	GetCurrentGoal(ctx context.Context, req gen.GetCurrentGoalRequestObject) (gen.GetCurrentGoalResponseObject, error)
	GetGoalSuggestions(ctx context.Context, req gen.GetGoalSuggestionsRequestObject) (gen.GetGoalSuggestionsResponseObject, error)
	StartGoal(ctx context.Context, req gen.StartGoalRequestObject) (gen.StartGoalResponseObject, error)
}

func (s Server) GetDashboard(ctx context.Context, req gen.GetDashboardRequestObject) (gen.GetDashboardResponseObject, error) {
	if s.Progress == nil {
		return gen.GetDashboard401Response{}, nil
	}
	return s.Progress.GetDashboard(ctx, req)
}

func (s Server) GetHeatmap(ctx context.Context, req gen.GetHeatmapRequestObject) (gen.GetHeatmapResponseObject, error) {
	if s.Progress == nil {
		return gen.GetHeatmap401Response{}, nil
	}
	return s.Progress.GetHeatmap(ctx, req)
}

func (s Server) GetCurrentGoal(ctx context.Context, req gen.GetCurrentGoalRequestObject) (gen.GetCurrentGoalResponseObject, error) {
	if s.Progress == nil {
		return gen.GetCurrentGoal401Response{}, nil
	}
	return s.Progress.GetCurrentGoal(ctx, req)
}

func (s Server) GetGoalSuggestions(ctx context.Context, req gen.GetGoalSuggestionsRequestObject) (gen.GetGoalSuggestionsResponseObject, error) {
	if s.Progress == nil {
		return gen.GetGoalSuggestions401Response{}, nil
	}
	return s.Progress.GetGoalSuggestions(ctx, req)
}

func (s Server) StartGoal(ctx context.Context, req gen.StartGoalRequestObject) (gen.StartGoalResponseObject, error) {
	if s.Progress == nil {
		return gen.StartGoal401Response{}, nil
	}
	return s.Progress.StartGoal(ctx, req)
}
