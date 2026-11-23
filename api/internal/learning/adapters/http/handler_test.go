package httpapi_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/keven-rdr/wordloop/api/internal/api"
	httpapi "github.com/keven-rdr/wordloop/api/internal/learning/adapters/http"
	"github.com/keven-rdr/wordloop/api/internal/learning/app"
	"github.com/keven-rdr/wordloop/api/internal/learning/domain"
	"github.com/keven-rdr/wordloop/api/internal/learning/learningtest"
	"github.com/keven-rdr/wordloop/api/internal/srs"
)

var (
	now  = time.Date(2026, 10, 19, 15, 0, 0, 0, time.UTC)
	user = uuid.MustParse("00000000-0000-0000-0000-0000000000aa")
)

func fixture(who domain.Viewer) (http.Handler, *learningtest.Store) {
	items := []learningtest.Item{
		{ID: uuid.New(), Content: domain.Content{Term: "house", Gloss: "casa", Example: "This is a big house.", ExampleGloss: "Esta é uma casa grande.", ImageKey: "openmoji/1F3E0.svg"}},
		{ID: uuid.New(), Content: domain.Content{Term: "cat", Gloss: "gato"}},
		{ID: uuid.New(), Content: domain.Content{Term: "dog", Gloss: "cachorro"}},
		{ID: uuid.New(), Content: domain.Content{Term: "tree", Gloss: "arvore"}},
	}
	store := learningtest.New(items...)
	viewers := &learningtest.Viewers{Who: who}
	clock := func() time.Time { return now }
	h := httpapi.Handler{
		Plan:     app.PlanSession{Viewers: viewers, Repo: store, Scheduler: srs.SimpleV1{NoFuzz: true}, Now: clock, NewID: uuid.New},
		Submit:   app.SubmitReviews{Viewers: viewers, Repo: store, Scheduler: srs.SimpleV1{NoFuzz: true}, Now: clock},
		Settings: app.SettingsService{Viewers: viewers, Repo: store},
	}
	return api.Handler(api.Server{Learning: h}), store
}

func post(h http.Handler, path, body string, csrf bool) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, path, bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	if csrf {
		req.Header.Set("X-Requested-With", "XMLHttpRequest")
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
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

func TestStudyEndpointsNeedASession(t *testing.T) {
	h, _ := fixture(domain.Viewer{})
	for path, body := range map[string]string{
		"/api/v1/study/plan":    `{"minutes":15}`,
		"/api/v1/study/reviews": `{"reviews":[]}`,
	} {
		if rec := post(h, path, body, true); rec.Code != http.StatusUnauthorized {
			t.Errorf("%s sem sessao: esperava 401, veio %d", path, rec.Code)
		}
	}
	closed := api.Handler(api.Server{})
	if rec := post(closed, "/api/v1/study/plan", `{"minutes":15}`, true); rec.Code != http.StatusUnauthorized {
		t.Errorf("estudo indisponivel: %d", rec.Code)
	}
	if rec := post(closed, "/api/v1/study/reviews", `{"reviews":[]}`, true); rec.Code != http.StatusUnauthorized {
		t.Errorf("estudo indisponivel: %d", rec.Code)
	}
}

func TestPlanReturnsThePackWithContentAndOptions(t *testing.T) {
	h, _ := fixture(domain.Viewer{UserID: user})
	rec := post(h, "/api/v1/study/plan", `{"minutes":15,"tz":"America/Sao_Paulo","locale":"pt-BR"}`, true)
	if rec.Code != http.StatusOK {
		t.Fatalf("plano: %d %s", rec.Code, rec.Body.String())
	}
	pack := decode[map[string]any](t, rec)
	cards := pack["cards"].([]any)
	if len(cards) != 4 || pack["new_limit"].(float64) != 16 || pack["review_limit"].(float64) != 60 || pack["session_id"] == "" {
		t.Fatalf("pacote: %s", rec.Body.String())
	}
	byTerm := map[string]map[string]any{}
	for _, c := range cards {
		card := c.(map[string]any)
		byTerm[card["term"].(string)] = card
	}
	first := byTerm["house"] // a ordem muda a cada sessao: procura pelo termo
	if first["term"] != "house" || first["gloss"] != "casa" || first["exercise"] != "choose" || first["direction"] != "receptive" ||
		first["state"] != "new" || first["image_url"] != "/media/openmoji/1F3E0.svg" || len(first["options"].([]any)) != 4 {
		t.Errorf("carta: %+v", first)
	}
	if ex := first["example"].(map[string]any); ex["text"] != "This is a big house." || ex["translation"] != "Esta é uma casa grande." {
		t.Errorf("exemplo: %+v", ex)
	}
	if second := byTerm["cat"]; second["image_url"] != nil || second["example"] != nil {
		t.Errorf("sem imagem nem exemplo, os campos nao aparecem: %+v", second)
	}
}

func TestPlanSendsTheClozeFieldsForProductiveFunctionalWords(t *testing.T) {
	item := learningtest.Item{ID: uuid.New(), Content: domain.Content{
		Term: "the", Gloss: "o", Example: "The cat sleeps.", ExampleGloss: "O gato dorme.", Theme: domain.FunctionalTheme,
	}}
	store := learningtest.New(item)
	id := uuid.New()
	store.Cards[id] = &domain.Card{ID: id, ItemID: item.ID, Direction: domain.Receptive, Scheduler: "simple_v1", State: srs.CardState{
		ID: id.String(), State: srs.StateReview, Stability: 10, Reps: 3, LastReviewAt: now.Add(-72 * time.Hour), DueAt: now.Add(24 * time.Hour),
	}}
	viewers := &learningtest.Viewers{Who: domain.Viewer{UserID: user}}
	clock := func() time.Time { return now }
	h := api.Handler(api.Server{Learning: httpapi.Handler{
		Plan: app.PlanSession{Viewers: viewers, Repo: store, Scheduler: srs.SimpleV1{NoFuzz: true}, Now: clock, NewID: uuid.New},
	}})
	pack := decode[map[string]any](t, post(h, "/api/v1/study/plan", `{"minutes":15,"mode":"learn"}`, true))
	for _, c := range pack["cards"].([]any) {
		card := c.(map[string]any)
		if card["direction"] != "productive" {
			continue
		}
		if card["exercise"] != "cloze" || card["cloze_text"] != "____ cat sleeps." || card["options"] != nil {
			t.Errorf("lacuna: %+v", card)
		}
		return
	}
	t.Fatal("o cartao produtivo deveria estar no plano")
}

func TestPlanValidation(t *testing.T) {
	h, _ := fixture(domain.Viewer{UserID: user})
	cases := map[string]struct {
		body string
		csrf bool
	}{
		"sem o cabecalho anti-CSRF": {`{"minutes":15}`, false},
		"minutos abaixo de 5":       {`{"minutes":4}`, true},
		"minutos acima de 240":      {`{"minutes":241}`, true},
		"modo desconhecido":         {`{"minutes":15,"mode":"tudo"}`, true},
		"corpo quebrado":            {`{"minutes":`, true},
		"sem corpo":                 {``, true},
	}
	for name, c := range cases {
		if rec := post(h, "/api/v1/study/plan", c.body, c.csrf); rec.Code != http.StatusBadRequest {
			t.Errorf("%s: esperava 400, veio %d", name, rec.Code)
		}
	}
}

func reviewBody(card, item uuid.UUID, id uuid.UUID) string {
	b, _ := json.Marshal(map[string]any{"reviews": []any{map[string]any{
		"id": id, "card_id": card, "exercise": "choose", "chosen_item_id": item,
		"shown_at": now.Add(-time.Minute).Format(time.RFC3339), "answered_at": now.Add(-50 * time.Second).Format(time.RFC3339),
		"difficulty": "medium", "client_correct": true, "response_ms": 3200, "client_seq": 1,
	}}})
	return string(b)
}

func TestSubmitReviewsEndToEnd(t *testing.T) {
	h, store := fixture(domain.Viewer{UserID: user})
	plan := decode[map[string]any](t, post(h, "/api/v1/study/plan", `{"minutes":15}`, true))
	card := plan["cards"].([]any)[0].(map[string]any)
	cardID, itemID := uuid.MustParse(card["id"].(string)), uuid.MustParse(card["item_id"].(string))
	id := uuid.New()

	rec := post(h, "/api/v1/study/reviews", reviewBody(cardID, itemID, id), true)
	if rec.Code != http.StatusOK {
		t.Fatalf("revisao: %d %s", rec.Code, rec.Body.String())
	}
	res := decode[map[string][]map[string]any](t, rec)["results"][0]
	progress := res["card"].(map[string]any)
	if res["status"] != "applied" || res["correct"] != true || res["grade"].(float64) != 3 ||
		progress["state"] != "review" || progress["stability"].(float64) != 2 || progress["mastery"].(float64) <= 0 || progress["leech"] != false {
		t.Fatalf("resultado: %+v", res)
	}
	if len(store.Log) != 1 || store.Log[0].ClientSeq != 1 || store.Log[0].ResponseMS != 3200 {
		t.Errorf("log: %+v", store.Log)
	}

	again := decode[map[string][]map[string]any](t, post(h, "/api/v1/study/reviews", reviewBody(cardID, itemID, id), true))["results"][0]
	if again["status"] != "duplicate" || len(store.Log) != 1 || again["grade"] != nil {
		t.Errorf("reenvio: %+v", again)
	}

	missing := decode[map[string][]map[string]any](t, post(h, "/api/v1/study/reviews", reviewBody(uuid.New(), itemID, uuid.New()), true))["results"][0]
	if missing["status"] != "rejected" || missing["reason"] != "card_not_found" || missing["card"] != nil {
		t.Errorf("cartao inexistente: %+v", missing)
	}
}

func TestSubmitReviewsValidation(t *testing.T) {
	h, _ := fixture(domain.Viewer{UserID: user})
	if rec := post(h, "/api/v1/study/reviews", `{"reviews":[]}`, true); rec.Code != http.StatusBadRequest {
		t.Errorf("lote vazio: %d", rec.Code)
	}
	if rec := post(h, "/api/v1/study/reviews", reviewBody(uuid.New(), uuid.New(), uuid.New()), false); rec.Code != http.StatusBadRequest {
		t.Errorf("sem CSRF: %d", rec.Code)
	}
	if rec := post(h, "/api/v1/study/reviews", `não é json`, true); rec.Code != http.StatusBadRequest {
		t.Errorf("corpo quebrado: %d", rec.Code)
	}
	if rec := post(h, "/api/v1/study/reviews", ``, true); rec.Code != http.StatusBadRequest {
		t.Errorf("sem corpo: %d", rec.Code)
	}
}

func TestRepositoryFailuresBecome500WithoutDetails(t *testing.T) {
	h, store := fixture(domain.Viewer{UserID: user})
	store.FailOn = "Settings"
	rec := post(h, "/api/v1/study/plan", `{"minutes":15}`, true)
	if rec.Code != http.StatusInternalServerError || rec.Body.Len() != 0 {
		t.Errorf("plano: %d %q", rec.Code, rec.Body.String())
	}
	rec = post(h, "/api/v1/study/reviews", reviewBody(uuid.New(), uuid.New(), uuid.New()), true)
	if rec.Code != http.StatusOK { // cartao nao existe: rejeitada, nao e erro
		t.Errorf("revisao de cartao inexistente: %d", rec.Code)
	}
}

func TestPlanSendsTilesAndTheServerGradesAnAssembledSentence(t *testing.T) {
	item := learningtest.Item{ID: uuid.New(), Content: domain.Content{Term: "I like the red house.", Gloss: "Eu gosto da casa vermelha.", Kind: domain.SentenceKind}}
	store := learningtest.New(item)
	id := uuid.New()
	store.Cards[id] = &domain.Card{ID: id, ItemID: item.ID, Direction: domain.Receptive, Scheduler: "simple_v1", State: srs.CardState{
		ID: id.String(), State: srs.StateReview, Stability: 4, Reps: 3, LastReviewAt: now.Add(-72 * time.Hour), DueAt: now.Add(24 * time.Hour),
	}}
	viewers := &learningtest.Viewers{Who: domain.Viewer{UserID: user}}
	clock := func() time.Time { return now }
	h := api.Handler(api.Server{Learning: httpapi.Handler{
		Plan:   app.PlanSession{Viewers: viewers, Repo: store, Scheduler: srs.SimpleV1{NoFuzz: true}, Now: clock, NewID: uuid.New},
		Submit: app.SubmitReviews{Viewers: viewers, Repo: store, Scheduler: srs.SimpleV1{NoFuzz: true}, Now: clock},
	}})
	pack := decode[map[string]any](t, post(h, "/api/v1/study/plan", `{"minutes":15,"mode":"learn"}`, true))
	var card map[string]any
	for _, c := range pack["cards"].([]any) {
		if c.(map[string]any)["direction"] == "productive" {
			card = c.(map[string]any)
		}
	}
	if card == nil || card["exercise"] != "assemble" || card["term"] == nil || card["options"] != nil {
		t.Fatalf("frase no produtivo: %+v", card)
	}
	tiles := card["tiles"].([]any)
	if len(tiles) != 5 || card["gloss"] != "Eu gosto da casa vermelha." {
		t.Fatalf("5 pecas e a traducao como pergunta: %+v", card)
	}
	body, _ := json.Marshal(map[string]any{"reviews": []any{map[string]any{
		"id": uuid.New(), "card_id": card["id"], "exercise": "assemble", "answer": "I like the red house",
		"shown_at": now.Add(-time.Minute).Format(time.RFC3339), "answered_at": now.Add(-50 * time.Second).Format(time.RFC3339),
		"difficulty": "medium", "client_seq": 1,
	}}})
	rec := post(h, "/api/v1/study/reviews", string(body), true)
	if rec.Code != http.StatusOK {
		t.Fatalf("revisao: %d %s", rec.Code, rec.Body.String())
	}
	res := decode[map[string][]map[string]any](t, rec)["results"][0]
	if res["status"] != "applied" || res["correct"] != true {
		t.Fatalf("o servidor corrige a frase montada: %+v", res)
	}
}
