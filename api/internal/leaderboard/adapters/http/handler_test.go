package httpapi_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/keven-rdr/wordloop/api/internal/api"
	httpapi "github.com/keven-rdr/wordloop/api/internal/leaderboard/adapters/http"
	"github.com/keven-rdr/wordloop/api/internal/leaderboard/app"
	"github.com/keven-rdr/wordloop/api/internal/leaderboard/domain"
)

var me = uuid.MustParse("00000000-0000-0000-0000-0000000000aa")

type who struct{ id uuid.UUID }

func (w who) Who(context.Context) (domain.Viewer, error) { return domain.Viewer{UserID: w.id}, nil }

type store struct{}

func (store) Clock(context.Context, uuid.UUID) (string, int, error) { return "UTC", 0, nil }
func (store) Ranking(context.Context, uuid.UUID, time.Time, time.Time, int) ([]domain.Entry, error) {
	return []domain.Entry{{UserID: uuid.New(), Position: 1, Nickname: "bia", Points: 50}, {UserID: me, Position: 2, Points: 10}}, nil
}
func (store) DaysMet(context.Context, uuid.UUID, time.Time, time.Time) (int, error) { return 2, nil }
func (store) Profile(context.Context, uuid.UUID) (domain.Profile, bool, error) {
	return domain.Profile{}, false, nil
}
func (store) SaveProfile(_ context.Context, _ uuid.UUID, p domain.Profile) (domain.Profile, error) {
	if p.Nickname == "bia" {
		return domain.Profile{}, domain.ErrNicknameTaken
	}
	return p, nil
}

func server(id uuid.UUID) http.Handler {
	return api.Handler(api.Server{Leaderboard: httpapi.Handler{Service: app.Service{Identity: who{id}, Store: store{}}}})
}

func do(h http.Handler, method, path, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func TestLeaderboardEndpoints(t *testing.T) {
	h := server(me)
	rec := do(h, http.MethodGet, "/api/v1/leaderboard?period=all&limit=10", "")
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"nickname":"bia"`) || !strings.Contains(rec.Body.String(), `"me":{"me":true,"points":10,"position":2}`) {
		t.Errorf("ranking: %d %s", rec.Code, rec.Body.String())
	}
	if rec := do(h, http.MethodGet, "/api/v1/leaderboard?period=month", ""); rec.Code != http.StatusBadRequest {
		t.Errorf("periodo invalido: %d", rec.Code)
	}
	if rec := do(h, http.MethodGet, "/api/v1/leaderboard/me", ""); rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"week_days_met":2`) {
		t.Errorf("resumo: %d %s", rec.Code, rec.Body.String())
	}
	cases := map[string]int{`{"nickname":"ana_1","visible":true}`: 200, `{"nickname":"a b","visible":true}`: 400, `{"nickname":"bia","visible":true}`: 409}
	for body, want := range cases {
		if rec := do(h, http.MethodPut, "/api/v1/leaderboard/profile", body); rec.Code != want {
			t.Errorf("%s: %d, esperava %d", body, rec.Code, want)
		}
	}
	if rec := do(h, http.MethodGet, "/api/v1/leaderboard/profile", ""); rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"visible":false`) {
		t.Errorf("perfil: %d %s", rec.Code, rec.Body.String())
	}
}

func TestLeaderboardEndpointsWithoutASessionAreUnauthorized(t *testing.T) {
	h := server(uuid.Nil)
	for _, c := range []struct{ method, path, body string }{
		{http.MethodGet, "/api/v1/leaderboard", ""},
		{http.MethodGet, "/api/v1/leaderboard/me", ""},
		{http.MethodGet, "/api/v1/leaderboard/profile", ""},
		{http.MethodPut, "/api/v1/leaderboard/profile", `{"nickname":"ana_1","visible":true}`},
	} {
		if rec := do(h, c.method, c.path, c.body); rec.Code != http.StatusUnauthorized {
			t.Errorf("%s %s sem sessao: %d", c.method, c.path, rec.Code)
		}
	}
}

// failing e o banco que sempre falha: toda rota vira 500, sem corpo pela metade.
type failing struct{ store }

var errDown = errors.New("banco fora")

func (failing) Clock(context.Context, uuid.UUID) (string, int, error) { return "", 0, errDown }
func (failing) Profile(context.Context, uuid.UUID) (domain.Profile, bool, error) {
	return domain.Profile{}, false, errDown
}
func (failing) SaveProfile(context.Context, uuid.UUID, domain.Profile) (domain.Profile, error) {
	return domain.Profile{}, errDown
}

func TestLeaderboardEndpointsTurnStoreFailuresInto500(t *testing.T) {
	h := api.Handler(api.Server{Leaderboard: httpapi.Handler{Service: app.Service{Identity: who{me}, Store: failing{}}}})
	for _, c := range []struct{ method, path, body string }{
		{http.MethodGet, "/api/v1/leaderboard", ""},
		{http.MethodGet, "/api/v1/leaderboard/me", ""},
		{http.MethodGet, "/api/v1/leaderboard/profile", ""},
		{http.MethodPut, "/api/v1/leaderboard/profile", `{"nickname":"ana_1","visible":true}`},
	} {
		if rec := do(h, c.method, c.path, c.body); rec.Code != http.StatusInternalServerError {
			t.Errorf("%s %s com o banco fora: %d", c.method, c.path, rec.Code)
		}
	}
}
