package app

import (
	"context"
	"time"

	"github.com/keven-rdr/wordloop/api/internal/progress/domain"
)

const (
	// defaultHeatmapDays sao as 12 semanas do painel.
	defaultHeatmapDays = 84
	maxHeatmapDays     = 400
)

// GetHeatmap lista os dias em que a pessoa estudou num intervalo (padrao: as ultimas 12 semanas).
type GetHeatmap struct {
	Identity Identity
	Reader   Reader
	Now      Clock
}

// Run devolve os dias entre from e to; datas nulas usam o padrao. Intervalo invertido ou maior que 400 dias
// e ErrInvalidRange.
func (u GetHeatmap) Run(ctx context.Context, from, to *time.Time) (domain.Heatmap, error) {
	who, err := resolve(ctx, u.Identity, u.Reader, u.Now)
	if err != nil {
		return domain.Heatmap{}, err
	}
	v, today := who.Viewer, who.Today
	start, end := today.AddDate(0, 0, 1-defaultHeatmapDays), today
	if from != nil {
		start = *from
	}
	if to != nil {
		end = *to
	}
	if end.Before(start) || end.Sub(start) > maxHeatmapDays*24*time.Hour {
		return domain.Heatmap{}, ErrInvalidRange
	}
	days, err := u.Reader.Days(ctx, v.UserID, start, end)
	if err != nil {
		return domain.Heatmap{}, err
	}
	return domain.Heatmap{From: start, To: end, Days: days}, nil
}
