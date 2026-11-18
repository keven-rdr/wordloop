package httpapi_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/keven-rdr/wordloop/api/internal/api"
	httpapi "github.com/keven-rdr/wordloop/api/internal/privacy/adapters/http"
	"github.com/keven-rdr/wordloop/api/internal/privacy/app"
	"github.com/keven-rdr/wordloop/api/internal/privacy/domain"
	"github.com/keven-rdr/wordloop/api/internal/privacy/privacytest"
)

var (
	now  = time.Date(2026, 11, 18, 15, 0, 0, 0, time.UTC)
	user = uuid.MustParse("00000000-0000-0000-0000-0000000000aa")
)

func fixture(who domain.Subject) (http.Handler, *privacytest.Store) {
	store := privacytest.New()
	store.Emails[user] = "ana@example.com"
	store.Data = json.RawMessage(`{"profile":{"display_name":"Ana"},"reviews":[]}`)
	h := httpapi.Handler{Service: app.Service{
		Identity: &privacytest.Identity{Is: who}, Store: store, Now: func() time.Time { return now },
	}}
	return api.Handler(api.Server{Privacy: h}), store
}

func recent() domain.Subject {
	return domain.Subject{UserID: user, AuthenticatedAt: now.Add(-time.Minute)}
}

func erase(h http.Handler, body string, csrf bool) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, "/api/v1/me/erase", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	if csrf {
		req.Header.Set("X-Requested-With", "XMLHttpRequest")
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func export(h http.Handler) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/v1/me/export", nil))
	return rec
}

func TestPrivacyEndpointsNeedASession(t *testing.T) {
	h, _ := fixture(domain.Subject{})
	if rec := export(h); rec.Code != http.StatusUnauthorized {
		t.Errorf("exportar sem sessao: %d", rec.Code)
	}
	if rec := erase(h, `{"confirmation":"ana@example.com"}`, true); rec.Code != http.StatusUnauthorized {
		t.Errorf("apagar sem sessao: %d", rec.Code)
	}
	if rec := export(api.Handler(api.Server{})); rec.Code != http.StatusUnauthorized {
		t.Errorf("modulo desligado responde 401: %d", rec.Code)
	}
}

func TestExportIsADownloadableJSONWithTheVersionAndTheDate(t *testing.T) {
	h, _ := fixture(recent())
	rec := export(h)
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d: %s", rec.Code, rec.Body.String())
	}
	if got := rec.Header().Get("Content-Disposition"); got != `attachment; filename="wordloop-dados-2026-11-18.json"` {
		t.Errorf("nome do arquivo: %q", got)
	}
	if !strings.HasPrefix(rec.Header().Get("Content-Type"), "application/json") {
		t.Errorf("tipo: %q", rec.Header().Get("Content-Type"))
	}
	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body["schema_version"] != float64(httpapi.SchemaVersion) || body["generated_at"] != "2026-11-18T15:00:00Z" {
		t.Errorf("versao e data: %v %v", body["schema_version"], body["generated_at"])
	}
	if profile, ok := body["profile"].(map[string]any); !ok || profile["display_name"] != "Ana" {
		t.Errorf("as secoes do banco entram como vieram: %v", body["profile"])
	}
}

func TestEraseAnswersEachCaseWithItsOwnStatus(t *testing.T) {
	cases := map[string]struct {
		who   domain.Subject
		body  string
		csrf  bool
		want  int
		wipes bool
	}{
		"tudo certo":                  {recent(), `{"confirmation":"ana@example.com"}`, true, http.StatusNoContent, true},
		"e-mail com outra caixa":      {recent(), `{"confirmation":"ANA@example.com"}`, true, http.StatusNoContent, true},
		"confirmacao errada":          {recent(), `{"confirmation":"bia@example.com"}`, true, http.StatusBadRequest, false},
		"sem confirmacao":             {recent(), `{}`, true, http.StatusBadRequest, false},
		"corpo ilegivel":              {recent(), `nao e json`, true, http.StatusBadRequest, false},
		"sem o cabecalho de protecao": {recent(), `{"confirmation":"ana@example.com"}`, false, http.StatusBadRequest, false},
		"administrador":               {domain.Subject{UserID: user, Admin: true, AuthenticatedAt: now}, `{"confirmation":"ana@example.com"}`, true, http.StatusForbidden, false},
		"login antigo":                {domain.Subject{UserID: user, AuthenticatedAt: now.Add(-time.Hour)}, `{"confirmation":"ana@example.com"}`, true, http.StatusPreconditionRequired, false},
	}
	for name, c := range cases {
		h, store := fixture(c.who)
		if rec := erase(h, c.body, c.csrf); rec.Code != c.want {
			t.Errorf("%s: esperava %d, veio %d (%s)", name, c.want, rec.Code, rec.Body.String())
		}
		if wiped := len(store.Erased) == 1; wiped != c.wipes {
			t.Errorf("%s: apagou = %v, esperava %v", name, wiped, c.wipes)
		}
	}
}

func TestErasingTwiceTheSecondTimeFindsNothing(t *testing.T) {
	h, _ := fixture(recent())
	if rec := erase(h, `{"confirmation":"ana@example.com"}`, true); rec.Code != http.StatusNoContent {
		t.Fatalf("1a: %d", rec.Code)
	}
	if rec := erase(h, `{"confirmation":"ana@example.com"}`, true); rec.Code != http.StatusUnauthorized {
		t.Errorf("2a: a conta nao existe mais, sem vazar nada: %d", rec.Code)
	}
}

func TestPrivacyModuleOffAnswers401ToErase(t *testing.T) {
	if rec := erase(api.Handler(api.Server{}), `{"confirmation":"ana@example.com"}`, true); rec.Code != http.StatusUnauthorized {
		t.Errorf("modulo desligado responde 401 tambem ao apagar: %d", rec.Code)
	}
}

func TestStoreFailuresBecomeServerErrorsAndNeverLeakDetails(t *testing.T) {
	h, store := fixture(recent())
	store.FailOn = "Export"
	rec := export(h)
	if rec.Code != http.StatusInternalServerError || strings.Contains(rec.Body.String(), "falha") {
		t.Errorf("falha ao exportar: %d %q", rec.Code, rec.Body.String())
	}

	h, store = fixture(recent())
	store.FailOn = "Erase"
	if rec := erase(h, `{"confirmation":"ana@example.com"}`, true); rec.Code != http.StatusInternalServerError {
		t.Errorf("falha ao apagar: %d", rec.Code)
	}
}

func TestAnExportThatIsNotAJSONObjectIsAServerError(t *testing.T) {
	h, store := fixture(recent())
	store.Data = json.RawMessage(`[1, 2, 3]`)
	if rec := export(h); rec.Code != http.StatusInternalServerError {
		t.Errorf("o banco devolveu algo que nao e um objeto: %d", rec.Code)
	}
}
