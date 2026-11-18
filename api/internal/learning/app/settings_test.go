package app_test

import (
	"context"
	"errors"
	"testing"

	"github.com/keven-rdr/wordloop/api/internal/learning/app"
	"github.com/keven-rdr/wordloop/api/internal/learning/domain"
	"github.com/keven-rdr/wordloop/api/internal/learning/learningtest"
)

func settingsOf(s *learningtest.Store, who domain.Viewer) app.SettingsService {
	return app.SettingsService{Viewers: &learningtest.Viewers{Who: who}, Repo: s}
}

func TestGetSettingsReturnsTheDefaultsForANewPerson(t *testing.T) {
	got, err := settingsOf(store("house"), student).Get(context.Background())
	if err != nil || got.UILocale != "system" || got.Theme != "system" || !got.TZAuto || got.TZ != "America/Sao_Paulo" {
		t.Errorf("padroes: %+v %v", got, err)
	}
}

func TestUpdateSettingsSavesThePreferencesAndKeepsTheRest(t *testing.T) {
	s := store("house")
	u := settingsOf(s, student)
	got, err := u.Update(context.Background(), domain.Settings{TZ: "Europe/Lisbon", UILocale: "en-US", Theme: "dark", TZAuto: false})
	if err != nil {
		t.Fatal(err)
	}
	if got.TZ != "Europe/Lisbon" || got.UILocale != "en-US" || got.Theme != "dark" || got.TZAuto {
		t.Errorf("devolveu: %+v", got)
	}
	if got.CutoffHour != 4 || got.Scheduler != "simple_v1" {
		t.Errorf("corte do dia e agendador nao mudam por aqui: %+v", got)
	}
	if s.Settings.TZ != "Europe/Lisbon" || s.Settings.Theme != "dark" || s.Settings.CutoffHour != 4 {
		t.Errorf("gravou: %+v", s.Settings)
	}
	again, _ := u.Get(context.Background())
	if again != got {
		t.Errorf("ler depois devolve o mesmo: %+v != %+v", again, got)
	}
}

func TestUpdateSettingsRejectsWhatIsOutOfTheOptions(t *testing.T) {
	good := domain.Settings{TZ: "America/Sao_Paulo", UILocale: "pt-BR", Theme: "light", TZAuto: true}
	cases := map[string]func(*domain.Settings){
		"fuso desconhecido":  func(s *domain.Settings) { s.TZ = "Marte/Olympus" },
		"fuso vazio":         func(s *domain.Settings) { s.TZ = "" },
		"fuso Local":         func(s *domain.Settings) { s.TZ = "Local" },
		"idioma inexistente": func(s *domain.Settings) { s.UILocale = "fr-FR" },
		"idioma vazio":       func(s *domain.Settings) { s.UILocale = "" },
		"tema inexistente":   func(s *domain.Settings) { s.Theme = "sepia" },
	}
	for name, mutate := range cases {
		s := store("house")
		before := s.Settings
		bad := good
		mutate(&bad)
		if _, err := settingsOf(s, student).Update(context.Background(), bad); !errors.Is(err, domain.ErrInvalidSettings) {
			t.Errorf("%s: %v", name, err)
		}
		if s.Settings != before {
			t.Errorf("%s: nada fica gravado: %+v", name, s.Settings)
		}
	}
	if _, err := settingsOf(store("house"), student).Update(context.Background(), good); err != nil {
		t.Errorf("padroes validos: %v", err)
	}
}

func TestSettingsNeedASessionAndPropagateFailures(t *testing.T) {
	anonymous := settingsOf(store("house"), domain.Viewer{})
	if _, err := anonymous.Get(context.Background()); !errors.Is(err, app.ErrNoSession) {
		t.Errorf("Get sem sessao: %v", err)
	}
	if _, err := anonymous.Update(context.Background(), domain.Settings{TZ: "America/Sao_Paulo", UILocale: "system", Theme: "system"}); !errors.Is(err, app.ErrNoSession) {
		t.Errorf("Update sem sessao: %v", err)
	}
	for _, method := range []string{"Settings", "SavePreferences"} {
		s := store("house")
		s.FailOn = method
		_, err := settingsOf(s, student).Update(context.Background(), domain.Settings{TZ: "Europe/Lisbon", UILocale: "system", Theme: "dark", TZAuto: true})
		if !errors.Is(err, learningtest.ErrBoom) {
			t.Errorf("falha em %s deveria subir: %v", method, err)
		}
		if s.Settings.Theme == "dark" {
			t.Errorf("falha em %s: nada fica gravado", method)
		}
	}
}

// Fuso escolhido nas configuracoes nao e desfeito pelo aparelho; so o fuso que acompanha o aparelho muda sozinho.
func TestPlanOnlyFollowsTheDeviceTimeZoneWhileTheZoneIsAuto(t *testing.T) {
	s := store("house")
	p := planner(s, student)
	_, _ = p.Run(context.Background(), app.PlanRequest{Minutes: 5, TZ: "Europe/Lisbon"})
	if s.Settings.TZ != "Europe/Lisbon" {
		t.Fatalf("fuso automatico acompanha o aparelho: %s", s.Settings.TZ)
	}

	if _, err := settingsOf(s, student).Update(context.Background(), domain.Settings{TZ: "Asia/Tokyo", UILocale: "system", Theme: "system", TZAuto: false}); err != nil {
		t.Fatal(err)
	}
	_, _ = p.Run(context.Background(), app.PlanRequest{Minutes: 5, TZ: "America/Sao_Paulo"})
	if s.Settings.TZ != "Asia/Tokyo" {
		t.Errorf("fuso fixo escolhido pela pessoa nao e desfeito pelo aparelho: %s", s.Settings.TZ)
	}

	if _, err := settingsOf(s, student).Update(context.Background(), domain.Settings{TZ: "Asia/Tokyo", UILocale: "system", Theme: "system", TZAuto: true}); err != nil {
		t.Fatal(err)
	}
	_, _ = p.Run(context.Background(), app.PlanRequest{Minutes: 5, TZ: "America/Sao_Paulo"})
	if s.Settings.TZ != "America/Sao_Paulo" {
		t.Errorf("voltando para automatico, o aparelho manda de novo: %s", s.Settings.TZ)
	}
}
