package api

import (
	"context"

	"github.com/keven-rdr/wordloop/api/internal/api/gen"
)

// Progress e o modulo de progresso (painel e heatmap, so leitura); nulo = progresso indisponivel.
type Progress interface {
	GetDashboard(ctx context.Context, req gen.GetDashboardRequestObject) (gen.GetDashboardResponseObject, error)
	GetHeatmap(ctx context.Context, req gen.GetHeatmapRequestObject) (gen.GetHeatmapResponseObject, error)
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
