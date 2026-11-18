package app

import (
	"context"

	"github.com/keven-rdr/wordloop/api/internal/progress/domain"
)

const (
	// streakLookbackDays e quanto do passado se olha para contar a sequencia.
	streakLookbackDays = 400
	forecastDays       = 7
	missedDays         = 30
	missedLimit        = 5
	accuracyDays       = 7
)

// GetDashboard monta o painel da pessoa: tudo calculado a partir do que o estudo ja gravou.
type GetDashboard struct {
	Identity Identity
	Reader   Reader
	Now      Clock
}

func (u GetDashboard) Run(ctx context.Context, locale string) (domain.Dashboard, error) {
	v, err := asker(ctx, u.Identity)
	if err != nil {
		return domain.Dashboard{}, err
	}
	s, err := u.Reader.Settings(ctx, v.UserID)
	if err != nil {
		return domain.Dashboard{}, err
	}
	today, err := domain.Today(u.Now.now(), s.TZ, s.CutoffHour)
	if err != nil {
		return domain.Dashboard{}, err
	}
	days, err := u.Reader.Days(ctx, v.UserID, domain.AddDays(today, -streakLookbackDays), today)
	if err != nil {
		return domain.Dashboard{}, err
	}
	d := domain.Dashboard{
		Today: today, Streak: domain.Streak(days, today),
		Accuracy7d: domain.Accuracy(domain.Window(days, domain.AddDays(today, 1-accuracyDays), today)),
	}
	for _, day := range domain.Window(days, today, today) {
		d.TodayStats, d.StudiedToday = day, day.Studied()
	}
	return u.fill(ctx, v, s, d, locale)
}

// fill completa o painel com o que vem de outras consultas: previsao, palavras conhecidas e mais erradas.
func (u GetDashboard) fill(ctx context.Context, v domain.Viewer, s domain.Settings, d domain.Dashboard, locale string) (domain.Dashboard, error) {
	due, err := u.Reader.Forecast(ctx, v.UserID, s, d.Today, forecastDays)
	if err != nil {
		return domain.Dashboard{}, err
	}
	d.Forecast = domain.Forecast(d.Today, due, forecastDays)
	if d.KnownWords, err = u.Reader.KnownWords(ctx, v.UserID); err != nil {
		return domain.Dashboard{}, err
	}
	d.MostMissed, err = u.Reader.MostMissed(ctx, v.UserID, domain.AddDays(d.Today, 1-missedDays), locale, missedLimit)
	return d, err
}
