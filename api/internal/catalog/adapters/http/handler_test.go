package httpapi_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"

	"github.com/keven-rdr/wordloop/api/internal/api"
	httpapi "github.com/keven-rdr/wordloop/api/internal/catalog/adapters/http"
	"github.com/keven-rdr/wordloop/api/internal/catalog/app"
	"github.com/keven-rdr/wordloop/api/internal/catalog/catalogtest"
	"github.com/keven-rdr/wordloop/api/internal/catalog/domain"
)

var houseID = uuid.MustParse("11111111-1111-1111-1111-111111111111")

func fixture() (*catalogtest.Viewers, *catalogtest.Repo) {
	house := domain.Summary{ID: houseID, Kind: "word", Term: "house", Gloss: "casa", CEFR: "A1", Status: "published", ImageKey: "openmoji/1F3E0.svg"}
	return &catalogtest.Viewers{Who: domain.Viewer{Authenticated: true}}, &catalogtest.Repo{
		Page: domain.Page{Items: []domain.Summary{house}, Page: 1, Size: 20, Total: 1},
		Detail: domain.Detail{
			Summary:  house,
			Examples: []domain.Example{{Text: "This is a big house.", Translation: "Esta é uma casa grande."}},
			Images:   []domain.Image{{Key: "openmoji/1F3E0.svg", Credit: "OpenMoji"}},
			Related:  []domain.Related{{ID: uuid.New(), Term: "big", Type: "collocation"}},
		},
		UnitsOf: []domain.Unit{{ID: uuid.New(), Position: 1, ThemeKey: "home-and-family", Items: 12}},
		Sources: []domain.Credit{{Code: "openmoji", Name: "OpenMoji", License: "CC-BY-SA-4.0", Notice: "All emojis designed by OpenMoji"}},
	}
}

func serve(viewers *catalogtest.Viewers, repo *catalogtest.Repo) http.Handler {
	h := httpapi.Handler{
		Units: app.ListUnits{Viewers: viewers, Repo: repo}, Items: app.ListItems{Viewers: viewers, Repo: repo},
		Item: app.GetItem{Viewers: viewers, Repo: repo}, Credits: app.ListCredits{Repo: repo},
	}
	return api.Handler(api.Server{Catalog: h})
}

func get(h http.Handler, path string) *httptest.ResponseRecorder {
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
	return rec
}

func decode[T any](t *testing.T, rec *httptest.ResponseRecorder) T {
	t.Helper()
	var out T
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("corpo: %v: %s", err, rec.Body.String())
	}
	return out
}

func TestWithoutASessionTheCatalogIsClosedButTheCreditsAreOpen(t *testing.T) {
	_, repo := fixture()
	h := serve(&catalogtest.Viewers{}, repo)
	for _, path := range []string{"/api/v1/catalog/units", "/api/v1/catalog/items", "/api/v1/catalog/items/" + houseID.String()} {
		if rec := get(h, path); rec.Code != http.StatusUnauthorized {
			t.Errorf("%s: esperava 401, veio %d", path, rec.Code)
		}
	}
	rec := get(h, "/api/v1/catalog/credits")
	if rec.Code != http.StatusOK || len(decode[[]map[string]any](t, rec)) != 1 {
		t.Fatalf("creditos: %d %s", rec.Code, rec.Body.String())
	}
}

func TestListingAnswersWithAPageAndTheMediaPath(t *testing.T) {
	viewers, repo := fixture()
	rec := get(serve(viewers, repo), "/api/v1/catalog/items?page=1&size=20&sort=cefr,desc&q=casa&cefr=A1&kind=word&locale=pt-BR")
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d: %s", rec.Code, rec.Body.String())
	}
	page := decode[struct {
		Items []map[string]any `json:"items"`
		Total int              `json:"total"`
	}](t, rec)
	if page.Total != 1 || len(page.Items) != 1 || page.Items[0]["term"] != "house" || page.Items[0]["image_url"] != "/media/openmoji/1F3E0.svg" {
		t.Fatalf("pagina: %+v", page)
	}
	if q := repo.LastQuery; q.Sort != domain.CEFRDesc || q.Text != "casa" || q.CEFR != "A1" || q.Kind != "word" {
		t.Fatalf("filtros nao chegaram ao caso de uso: %+v", q)
	}
}

func TestAnInvalidListingIsABadRequest(t *testing.T) {
	viewers, repo := fixture()
	h := serve(viewers, repo)
	for _, path := range []string{"/api/v1/catalog/items?sort=id,asc", "/api/v1/catalog/items?size=1000", "/api/v1/catalog/items?cefr=Z9", "/api/v1/catalog/items?page=abc"} {
		if rec := get(h, path); rec.Code != http.StatusBadRequest {
			t.Errorf("%s: esperava 400, veio %d", path, rec.Code)
		}
	}
}

func TestItemDetailAndNotFound(t *testing.T) {
	viewers, repo := fixture()
	h := serve(viewers, repo)
	rec := get(h, "/api/v1/catalog/items/"+houseID.String()+"?locale=en-US")
	if rec.Code != http.StatusOK {
		t.Fatalf("status %d: %s", rec.Code, rec.Body.String())
	}
	d := decode[struct {
		Examples []map[string]any `json:"examples"`
		Images   []map[string]any `json:"images"`
		Related  []map[string]any `json:"related"`
	}](t, rec)
	if len(d.Examples) != 1 || len(d.Related) != 1 || len(d.Images) != 1 || d.Images[0]["credit"] != "OpenMoji" || repo.LastLocale != "en-US" {
		t.Fatalf("detalhe: %+v (idioma %s)", d, repo.LastLocale)
	}
	if rec := get(h, "/api/v1/catalog/items/"+uuid.NewString()); rec.Code != http.StatusNotFound {
		t.Fatalf("item inexistente: %d", rec.Code)
	}
	if rec := get(h, "/api/v1/catalog/items/"+houseID.String()+"?locale=fr-FR"); rec.Code != http.StatusBadRequest {
		t.Fatalf("idioma invalido: %d", rec.Code)
	}
	if rec := get(h, "/api/v1/catalog/items/nao-e-uuid"); rec.Code != http.StatusBadRequest {
		t.Fatalf("id invalido: %d", rec.Code)
	}
}

func TestUnitsAndAnUnexpectedErrorIsA500(t *testing.T) {
	viewers, repo := fixture()
	h := serve(viewers, repo)
	rec := get(h, "/api/v1/catalog/units")
	units := decode[[]map[string]any](t, rec)
	if rec.Code != http.StatusOK || len(units) != 1 || units[0]["theme_key"] != "home-and-family" || units[0]["item_count"] != float64(12) {
		t.Fatalf("unidades: %d %s", rec.Code, rec.Body.String())
	}
	repo.Err = http.ErrAbortHandler
	for _, path := range []string{"/api/v1/catalog/units", "/api/v1/catalog/items", "/api/v1/catalog/credits"} {
		if rec := get(h, path); rec.Code != http.StatusInternalServerError {
			t.Errorf("%s: esperava 500, veio %d", path, rec.Code)
		}
	}
}

func TestWithoutTheModuleTheCatalogIsClosedAndCreditsEmpty(t *testing.T) {
	h := api.Handler(api.Server{})
	if rec := get(h, "/api/v1/catalog/items"); rec.Code != http.StatusUnauthorized {
		t.Fatalf("itens: %d", rec.Code)
	}
	if rec := get(h, "/api/v1/catalog/credits"); rec.Code != http.StatusOK || rec.Body.String() == "" {
		t.Fatalf("creditos: %d", rec.Code)
	}
}
