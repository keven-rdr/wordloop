// Package httpapi expoe o progresso pelo contrato gerado (strict server).
package httpapi

import (
	"context"
	"errors"

	"github.com/keven-rdr/wordloop/api/internal/api/gen"
	"github.com/keven-rdr/wordloop/api/internal/progress/app"
)

// Handler junta os casos de uso do progresso.
type Handler struct {
	Dashboard app.GetDashboard
	Heatmap   app.GetHeatmap
	Goals     app.GoalService
}

func (h Handler) GetDashboard(ctx context.Context, req gen.GetDashboardRequestObject) (gen.GetDashboardResponseObject, error) {
	locale := "pt-BR"
	if req.Params.Locale != nil {
		locale = string(*req.Params.Locale)
	}
	d, err := h.Dashboard.Run(ctx, locale)
	switch {
	case errors.Is(err, app.ErrNoSession):
		return gen.GetDashboard401Response{}, nil
	case err != nil:
		return nil, err
	}
	return gen.GetDashboard200JSONResponse(dashboard(d)), nil
}

func (h Handler) GetHeatmap(ctx context.Context, req gen.GetHeatmapRequestObject) (gen.GetHeatmapResponseObject, error) {
	hm, err := h.Heatmap.Run(ctx, day(req.Params.From), day(req.Params.To))
	switch {
	case errors.Is(err, app.ErrNoSession):
		return gen.GetHeatmap401Response{}, nil
	case errors.Is(err, app.ErrInvalidRange):
		return gen.GetHeatmap400Response{}, nil
	case err != nil:
		return nil, err
	}
	return gen.GetHeatmap200JSONResponse(heatmap(hm)), nil
}
