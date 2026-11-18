package httpapi

import (
	"time"

	openapi_types "github.com/oapi-codegen/runtime/types"

	"github.com/keven-rdr/wordloop/api/internal/api/gen"
	"github.com/keven-rdr/wordloop/api/internal/progress/domain"
)

// day converte a data do pedido em meia-noite UTC, como o dia de estudo e gravado; nil continua nil.
func day(d *openapi_types.Date) *time.Time {
	if d == nil {
		return nil
	}
	t := time.Date(d.Year(), d.Month(), d.Day(), 0, 0, 0, 0, time.UTC)
	return &t
}

func date(t time.Time) openapi_types.Date { return openapi_types.Date{Time: t} }

func dashboard(d domain.Dashboard) gen.Dashboard {
	out := gen.Dashboard{
		Today: date(d.Today), Streak: d.Streak, StudiedToday: d.StudiedToday, KnownWords: d.KnownWords,
		TodayStats: dayStats(d.TodayStats),
		Forecast:   make([]gen.ForecastDay, len(d.Forecast)),
		MostMissed: make([]gen.MissedItem, len(d.MostMissed)),
	}
	if d.Accuracy7d != nil {
		rate := float32(*d.Accuracy7d)
		out.Accuracy7d = &rate
	}
	for i, f := range d.Forecast {
		out.Forecast[i] = gen.ForecastDay{Date: date(f.Date), Due: f.Due}
	}
	for i, m := range d.MostMissed {
		out.MostMissed[i] = gen.MissedItem{ItemId: m.ItemID, Term: m.Term, Gloss: m.Gloss, Misses: m.Misses}
	}
	return out
}

func dayStats(d domain.Day) gen.DayStats {
	return gen.DayStats{Reviews: d.Reviews, Correct: d.Correct, NewItems: d.NewItems, ActiveSeconds: int(d.ActiveMS / 1000)}
}

func heatmap(h domain.Heatmap) gen.Heatmap {
	out := gen.Heatmap{From: date(h.From), To: date(h.To), Days: make([]gen.HeatmapDay, len(h.Days))}
	for i, d := range h.Days {
		out.Days[i] = gen.HeatmapDay{Date: date(d.Date), Reviews: d.Reviews, Correct: d.Correct}
	}
	return out
}
