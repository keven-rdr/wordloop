//go:build integration

package postgres_test

import (
	"context"
	"testing"

	"github.com/keven-rdr/wordloop/api/internal/learning/app"
	"github.com/keven-rdr/wordloop/api/internal/learning/domain"
	"github.com/keven-rdr/wordloop/api/internal/learning/learningtest"
)

func (f fixture) settings() app.SettingsService {
	return app.SettingsService{Viewers: &learningtest.Viewers{Who: domain.Viewer{UserID: f.user}}, Repo: f.store}
}

func TestPreferencesStartWithTheDefaultsAndAreSavedForReal(t *testing.T) {
	f := setup(t, false)
	ctx := context.Background()
	got, err := f.settings().Get(ctx)
	if err != nil || got.UILocale != "system" || got.Theme != "system" || !got.TZAuto || got.TZ != "America/Sao_Paulo" || got.CutoffHour != 4 {
		t.Fatalf("padroes: %+v %v", got, err)
	}
	saved, err := f.settings().Update(ctx, domain.Settings{TZ: "Asia/Tokyo", UILocale: "en-US", Theme: "dark", TZAuto: false})
	if err != nil {
		t.Fatal(err)
	}
	again, err := f.settings().Get(ctx)
	if err != nil || again != saved || again.TZ != "Asia/Tokyo" || again.UILocale != "en-US" || again.Theme != "dark" || again.TZAuto {
		t.Errorf("lido de volta: %+v %v (gravado %+v)", again, err, saved)
	}
	if again.CutoffHour != 4 || again.Scheduler != "simple_v1" {
		t.Errorf("o corte do dia e o agendador nao mudam: %+v", again)
	}
}

func TestPreferencesOfOnePersonDoNotLeakToAnother(t *testing.T) {
	f := setup(t, false)
	ctx := context.Background()
	var other string
	if err := f.pool.QueryRow(ctx, `INSERT INTO app_user (display_name) VALUES ('Bia') RETURNING id::text`).Scan(&other); err != nil {
		t.Fatal(err)
	}
	if _, err := f.settings().Update(ctx, domain.Settings{TZ: "Asia/Tokyo", UILocale: "en-US", Theme: "dark", TZAuto: false}); err != nil {
		t.Fatal(err)
	}
	var theme string
	if err := f.pool.QueryRow(ctx, `INSERT INTO user_settings (user_id) VALUES ($1::uuid) RETURNING theme`, other).Scan(&theme); err != nil || theme != "system" {
		t.Errorf("a outra pessoa tem os padroes: %q %v", theme, err)
	}
}

func TestTheDatabaseRefusesPreferencesOutOfTheOptions(t *testing.T) {
	f := setup(t, false)
	ctx := context.Background()
	if _, err := f.settings().Get(ctx); err != nil { // cria a linha
		t.Fatal(err)
	}
	for name, sql := range map[string]string{
		"idioma": `UPDATE user_settings SET ui_locale = 'fr-FR'`,
		"tema":   `UPDATE user_settings SET theme = 'sepia'`,
	} {
		if _, err := f.pool.Exec(ctx, sql); err == nil {
			t.Errorf("%s: o banco deveria recusar", name)
		}
	}
}

// Com o banco de verdade: o fuso escolhido nas configuracoes nao e desfeito pelo aparelho ao montar uma sessao.
func TestPlanKeepsAFixedTimeZoneChosenInTheSettings(t *testing.T) {
	f := setup(t, false)
	ctx := context.Background()
	if _, err := f.settings().Update(ctx, domain.Settings{TZ: "Asia/Tokyo", UILocale: "system", Theme: "system", TZAuto: false}); err != nil {
		t.Fatal(err)
	}
	if _, err := f.plan.Run(ctx, app.PlanRequest{Minutes: 5, TZ: "Europe/Lisbon"}); err != nil {
		t.Fatal(err)
	}
	got, _ := f.settings().Get(ctx)
	if got.TZ != "Asia/Tokyo" {
		t.Errorf("fuso fixo mantido: %s", got.TZ)
	}

	if _, err := f.settings().Update(ctx, domain.Settings{TZ: "Asia/Tokyo", UILocale: "system", Theme: "system", TZAuto: true}); err != nil {
		t.Fatal(err)
	}
	if _, err := f.plan.Run(ctx, app.PlanRequest{Minutes: 5, TZ: "Europe/Lisbon"}); err != nil {
		t.Fatal(err)
	}
	if got, _ := f.settings().Get(ctx); got.TZ != "Europe/Lisbon" {
		t.Errorf("automatico acompanha o aparelho: %s", got.TZ)
	}
}
