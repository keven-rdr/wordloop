package httpapi_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/keven-rdr/wordloop/api/internal/api"
	"github.com/keven-rdr/wordloop/api/internal/api/gen"
	httpapi "github.com/keven-rdr/wordloop/api/internal/progress/adapters/http"
	"github.com/keven-rdr/wordloop/api/internal/progress/app"
	"github.com/keven-rdr/wordloop/api/internal/progress/domain"
	"github.com/keven-rdr/wordloop/api/internal/progress/progresstest"
)

var (
	now  = time.Date(2026, 10, 19, 15, 0, 0, 0, time.UTC)
	user = uuid.MustParse("00000000-0000-0000-0000-0000000000aa")
)

func day(s string) time.Time {
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		panic(err)
	}
	return t
}

func fixture(who domain.Viewer) (http.Handler, *progresstest.Store) {
	store := progresstest.New()
	identity := &progresstest.Identity{Is: who}
	clock := func() time.Time { return now }
	h := httpapi.Handler{
		Dashboard: app.GetDashboard{Identity: identity, Reader: store, Now: clock},
		Heatmap:   app.GetHeatmap{Identity: identity, Reader: store, Now: clock},
		Goals:     app.GoalService{Identity: identity, Reader: store, Goals: store, Now: clock},
	}
	return api.Handler(api.Server{Progress: h}), store
}

func get(h http.Handler, path string) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
	return rec
}

func TestProgressEndpointsNeedASession(t *testing.T) {
	h, _ := fixture(domain.Viewer{})
	for _, path := range []string{"/api/v1/progress/dashboard", "/api/v1/progress/heatmap"} {
		if rec := get(h, path); rec.Code != http.StatusUnauthorized {
			t.Errorf("%s sem sessao: esperava 401, veio %d", path, rec.Code)
		}
	}
	if rec := get(api.Handler(api.Server{}), "/api/v1/progress/dashboard"); rec.Code != http.StatusUnauthorized {
		t.Errorf("modulo desligado responde 401, veio %d", rec.Code)
	}
}

func TestDashboardSpeaksTheContract(t *testing.T) {
	h, store := fixture(domain.Viewer{UserID: user})
	store.Rows = []domain.Day{{Date: day("2026-10-18"), Reviews: 10, Correct: 8}, {Date: day("2026-10-19"), Reviews: 4, Correct: 3, NewItems: 2, ActiveMS: 61_000}}
	store.Due, store.Known = map[int]int{1: 5}, 7
	store.Missed = []domain.Missed{{ItemID: uuid.New(), Term: "house", Gloss: "casa", Misses: 2}}

	rec := get(h, "/api/v1/progress/dashboard?locale=en-US")
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d: %s", rec.Code, rec.Body.String())
	}
	var d gen.Dashboard
	if err := json.Unmarshal(rec.Body.Bytes(), &d); err != nil {
		t.Fatal(err)
	}
	if d.Today.Format("2006-01-02") != "2026-10-19" || d.Streak != 2 || !d.StudiedToday || d.KnownWords != 7 || len(d.Forecast) != 7 {
		t.Errorf("painel: %+v", d)
	}
	if d.TodayStats.ActiveSeconds != 61 || d.TodayStats.NewItems != 2 || d.Forecast[1].Due != 5 || d.MostMissed[0].Term != "house" {
		t.Errorf("detalhes: %+v", d)
	}
	if d.Accuracy7d == nil || *d.Accuracy7d < 0.78 || *d.Accuracy7d > 0.79 {
		t.Errorf("11 de 14: %v", d.Accuracy7d)
	}
}

func TestDashboardOmitsAccuracyWhenThereAreNoAnswers(t *testing.T) {
	h, _ := fixture(domain.Viewer{UserID: user})
	rec := get(h, "/api/v1/progress/dashboard")
	var raw map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &raw); err != nil || rec.Code != http.StatusOK {
		t.Fatalf("%d %v", rec.Code, err)
	}
	if _, has := raw["accuracy_7d"]; has {
		t.Errorf("sem respostas, sem taxa: %v", raw)
	}
}

func TestHeatmapListsStudiedDaysAndRejectsBadRanges(t *testing.T) {
	h, store := fixture(domain.Viewer{UserID: user})
	store.Rows = []domain.Day{{Date: day("2026-10-10"), Reviews: 3, Correct: 2}, {Date: day("2026-09-01"), Reviews: 9, Correct: 9}}

	rec := get(h, "/api/v1/progress/heatmap?from=2026-10-01&to=2026-10-19")
	var hm gen.Heatmap
	if err := json.Unmarshal(rec.Body.Bytes(), &hm); err != nil || rec.Code != http.StatusOK {
		t.Fatalf("%d %v %s", rec.Code, err, rec.Body.String())
	}
	if len(hm.Days) != 1 || hm.Days[0].Reviews != 3 || hm.From.Format("2006-01-02") != "2026-10-01" {
		t.Errorf("mapa: %+v", hm)
	}
	for _, q := range []string{"from=2026-10-19&to=2026-10-01", "from=2024-01-01&to=2026-10-19"} {
		if rec := get(h, "/api/v1/progress/heatmap?"+q); rec.Code != http.StatusBadRequest {
			t.Errorf("%s: esperava 400, veio %d", q, rec.Code)
		}
	}
	if rec := get(h, "/api/v1/progress/heatmap?from=ontem"); rec.Code != http.StatusBadRequest {
		t.Errorf("data ilegivel: esperava 400, veio %d", rec.Code)
	}
}
