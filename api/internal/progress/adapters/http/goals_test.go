package httpapi_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/keven-rdr/wordloop/api/internal/api/gen"
	"github.com/keven-rdr/wordloop/api/internal/progress/domain"
)

func send(h http.Handler, method, path, body string, csrf bool) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	if csrf {
		req.Header.Set("X-Requested-With", "XMLHttpRequest")
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func TestGoalEndpointsNeedASession(t *testing.T) {
	h, _ := fixture(domain.Viewer{})
	for _, c := range []struct{ method, path, body string }{
		{http.MethodGet, "/api/v1/goals/current", ""},
		{http.MethodGet, "/api/v1/goals/suggestions", ""},
		{http.MethodPost, "/api/v1/goals", `{"daily_minutes":10,"days_target":5}`},
	} {
		if rec := send(h, c.method, c.path, c.body, true); rec.Code != http.StatusUnauthorized {
			t.Errorf("%s %s sem sessao: esperava 401, veio %d", c.method, c.path, rec.Code)
		}
	}
}

func TestStartAndReadTheCurrentGoal(t *testing.T) {
	h, store := fixture(domain.Viewer{UserID: user})
	store.Rows = []domain.Day{{Date: day("2026-10-19"), Reviews: 5, ActiveMS: 120_000, NewItems: 2}}

	empty := send(h, http.MethodGet, "/api/v1/goals/current", "", false)
	var none gen.CurrentGoal
	if err := json.Unmarshal(empty.Body.Bytes(), &none); err != nil || empty.Code != http.StatusOK || none.Goal != nil || none.Ended {
		t.Fatalf("sem meta: %d %s", empty.Code, empty.Body.String())
	}

	rec := send(h, http.MethodPost, "/api/v1/goals", `{"preset":"steady","daily_minutes":10}`, true)
	if rec.Code != http.StatusCreated {
		t.Fatalf("status %d: %s", rec.Code, rec.Body.String())
	}
	var cur gen.CurrentGoal
	if err := json.Unmarshal(rec.Body.Bytes(), &cur); err != nil || cur.Goal == nil || cur.Progress == nil {
		t.Fatalf("corpo: %v %s", err, rec.Body.String())
	}
	if cur.Goal.PeriodStart.Format("2006-01-02") != "2026-10-19" || cur.Goal.PeriodEnd.Format("2006-01-02") != "2026-10-25" ||
		*cur.Goal.Preset != gen.GoalPresetSteady || len(cur.Goal.Weekdays) != 7 || cur.Goal.DaysTarget != 7 {
		t.Errorf("meta: %+v", cur.Goal)
	}
	if p := cur.Progress; p.DaysStudied != 1 || p.MinutesStudied != 2 || p.NewWords != 2 || p.MinutesGoal != 70 || p.DaysLeft != 6 || p.Met {
		t.Errorf("progresso: %+v", p)
	}

	again := send(h, http.MethodGet, "/api/v1/goals/current", "", false)
	var read gen.CurrentGoal
	if err := json.Unmarshal(again.Body.Bytes(), &read); err != nil || read.Goal == nil || read.Goal.Id != cur.Goal.Id {
		t.Errorf("a meta aberta e a que volta: %d %s", again.Code, again.Body.String())
	}
}

func TestStartGoalRejectsBadRequests(t *testing.T) {
	h, store := fixture(domain.Viewer{UserID: user})
	cases := map[string]struct {
		body string
		csrf bool
	}{
		"sem o cabecalho de protecao": {`{"daily_minutes":10,"days_target":5}`, false},
		"minutos demais":              {`{"daily_minutes":999,"days_target":5}`, true},
		"dias demais":                 {`{"daily_minutes":10,"days_target":9}`, true},
		"menos de 1 minuto":           {`{"daily_minutes":0}`, true},
		"preset desconhecido":         {`{"preset":"turbo","daily_minutes":10,"days_target":3}`, true},
		"corpo ilegivel":              {`nao e json`, true},
		"faltam campos":               {`{}`, true},
	}
	for name, c := range cases {
		if rec := send(h, http.MethodPost, "/api/v1/goals", c.body, c.csrf); rec.Code != http.StatusBadRequest {
			t.Errorf("%s: esperava 400, veio %d (%s)", name, rec.Code, rec.Body.String())
		}
	}
	if len(store.Cycles) != 0 {
		t.Errorf("pedido invalido nao grava: %+v", store.Cycles)
	}
}

func TestGoalSuggestionsAreThreeAndSayWhereTheyCameFrom(t *testing.T) {
	h, store := fixture(domain.Viewer{UserID: user})
	var s gen.GoalSuggestions
	rec := send(h, http.MethodGet, "/api/v1/goals/suggestions", "", false)
	if err := json.Unmarshal(rec.Body.Bytes(), &s); err != nil || rec.Code != http.StatusOK || len(s.Suggestions) != 3 || s.BasedOnHistory {
		t.Fatalf("valores iniciais: %d %s", rec.Code, rec.Body.String())
	}
	if s.Suggestions[0].Preset != gen.GoalSuggestionPresetLight || s.Suggestions[2].Preset != gen.GoalSuggestionPresetIntense {
		t.Errorf("da mais leve a mais intensa: %+v", s.Suggestions)
	}
	for i := 0; i < 6; i++ {
		store.Rows = append(store.Rows, domain.Day{Date: day("2026-10-19").AddDate(0, 0, -i*2), Reviews: 4, ActiveMS: 600_000})
	}
	rec = send(h, http.MethodGet, "/api/v1/goals/suggestions", "", false)
	if err := json.Unmarshal(rec.Body.Bytes(), &s); err != nil || !s.BasedOnHistory || s.Suggestions[1].DailyMinutes != 10 {
		t.Errorf("a partir do historico: %s", rec.Body.String())
	}
}
