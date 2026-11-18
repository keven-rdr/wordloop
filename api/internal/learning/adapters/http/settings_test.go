package httpapi_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/keven-rdr/wordloop/api/internal/api/gen"
	"github.com/keven-rdr/wordloop/api/internal/learning/domain"
)

func put(h http.Handler, body string, csrf bool) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPut, "/api/v1/me/settings", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	if csrf {
		req.Header.Set("X-Requested-With", "XMLHttpRequest")
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func TestSettingsNeedASession(t *testing.T) {
	h, _ := fixture(domain.Viewer{})
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/me/settings", nil))
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("GET sem sessao: %d", rec.Code)
	}
	if rec := put(h, `{"ui_locale":"system","theme":"system","tz":"America/Sao_Paulo","tz_auto":true}`, true); rec.Code != http.StatusUnauthorized {
		t.Errorf("PUT sem sessao: %d", rec.Code)
	}
}

func TestSettingsAreReadAndSaved(t *testing.T) {
	h, store := fixture(domain.Viewer{UserID: user})
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/me/settings", nil))
	first := decode[gen.Settings](t, rec)
	if rec.Code != http.StatusOK || first.UiLocale != gen.SettingsUiLocaleSystem || first.Theme != gen.SettingsThemeSystem || !first.TzAuto || first.Tz != "America/Sao_Paulo" {
		t.Fatalf("padroes: %d %+v", rec.Code, first)
	}

	saved := put(h, `{"ui_locale":"en-US","theme":"dark","tz":"Europe/Lisbon","tz_auto":false}`, true)
	got := decode[gen.Settings](t, saved)
	if saved.Code != http.StatusOK || got.UiLocale != gen.SettingsUiLocaleEnUS || got.Theme != gen.SettingsThemeDark || got.Tz != "Europe/Lisbon" || got.TzAuto {
		t.Fatalf("gravou: %d %+v", saved.Code, got)
	}
	if store.Settings.TZ != "Europe/Lisbon" || store.Settings.Theme != "dark" || store.Settings.TZAuto {
		t.Errorf("no repositorio: %+v", store.Settings)
	}
}

func TestSettingsRejectBadRequests(t *testing.T) {
	h, store := fixture(domain.Viewer{UserID: user})
	before := store.Settings
	cases := map[string]struct {
		body string
		csrf bool
	}{
		"sem o cabecalho de protecao": {`{"ui_locale":"system","theme":"system","tz":"America/Sao_Paulo","tz_auto":true}`, false},
		"fuso desconhecido":           {`{"ui_locale":"system","theme":"system","tz":"Marte/Olympus","tz_auto":true}`, true},
		"idioma inexistente":          {`{"ui_locale":"fr-FR","theme":"system","tz":"America/Sao_Paulo","tz_auto":true}`, true},
		"tema inexistente":            {`{"ui_locale":"system","theme":"sepia","tz":"America/Sao_Paulo","tz_auto":true}`, true},
		"corpo ilegivel":              {`nao e json`, true},
		"faltam campos":               {`{}`, true},
	}
	for name, c := range cases {
		if rec := put(h, c.body, c.csrf); rec.Code != http.StatusBadRequest {
			t.Errorf("%s: esperava 400, veio %d (%s)", name, rec.Code, rec.Body.String())
		}
	}
	if store.Settings != before {
		t.Errorf("pedido invalido nao grava: %+v", store.Settings)
	}
}
