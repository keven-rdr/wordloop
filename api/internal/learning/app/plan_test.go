package app_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/keven-rdr/wordloop/api/internal/learning/app"
	"github.com/keven-rdr/wordloop/api/internal/learning/domain"
	"github.com/keven-rdr/wordloop/api/internal/learning/learningtest"
	"github.com/keven-rdr/wordloop/api/internal/srs"
)

var (
	now  = time.Date(2026, 10, 19, 15, 0, 0, 0, time.UTC)
	user = uuid.MustParse("00000000-0000-0000-0000-0000000000aa")
)

func words(terms ...string) []learningtest.Item {
	glosses := map[string]string{"house": "casa", "cat": "gato", "dog": "cachorro", "tree": "arvore", "water": "agua", "bread": "pao"}
	items := make([]learningtest.Item, 0, len(terms))
	for _, term := range terms {
		items = append(items, learningtest.Item{ID: uuid.New(), Content: domain.Content{Term: term, Gloss: glosses[term]}})
	}
	return items
}

// many cria n itens com termos e traducoes distintos.
func many(n int) []learningtest.Item {
	items := make([]learningtest.Item, 0, n)
	for i := 0; i < n; i++ {
		items = append(items, learningtest.Item{ID: uuid.New(), Content: domain.Content{Term: fmt.Sprintf("w%d", i), Gloss: fmt.Sprintf("g%d", i)}})
	}
	return items
}

func store(terms ...string) *learningtest.Store { return learningtest.New(words(terms...)...) }

func planner(s *learningtest.Store, who domain.Viewer) app.PlanSession {
	return app.PlanSession{
		Viewers: &learningtest.Viewers{Who: who}, Repo: s, Scheduler: srs.SimpleV1{NoFuzz: true},
		Now: func() time.Time { return now }, NewID: uuid.New,
	}
}

var student = domain.Viewer{UserID: user}

func TestPlanRequiresASession(t *testing.T) {
	p := planner(store("house"), domain.Viewer{})
	if _, err := p.Run(context.Background(), app.PlanRequest{Minutes: 15}); !errors.Is(err, app.ErrNoSession) {
		t.Fatalf("sem sessao: %v", err)
	}
}

func TestPlanCreatesTheNextNewCardsWithContentAndOptions(t *testing.T) {
	s := learningtest.New(many(12)...)
	pack, err := planner(s, student).Run(context.Background(), app.PlanRequest{Minutes: 5})
	if err != nil {
		t.Fatal(err)
	}
	if pack.NewLimit != 9 || pack.ReviewLimit != 20 || len(pack.Cards) != 9 || len(s.Sessions) != 1 || pack.SessionID != s.Sessions[0] {
		t.Fatalf("5 minutos: 9 novas e 20 revisoes: %d cartoes, %+v", len(pack.Cards), pack)
	}
	seen := map[uuid.UUID]bool{}
	for i, c := range pack.Cards {
		if c.Card.Direction != domain.Receptive || c.Card.State.State != srs.StateNew || c.Card.Exercise() != domain.Choose || seen[c.Card.ItemID] {
			t.Errorf("carta %d: %+v", i, c)
		}
		seen[c.Card.ItemID] = true
		if len(c.Options) != 4 || !contains(c.Options, c.Card.ItemID) {
			t.Errorf("a pergunta de escolher tem a resposta certa e 3 distratores: %+v", c.Options)
		}
	}
	for _, it := range s.Items[:9] { // as 9 primeiras da ordem de estudo (o duble ignora a semente)
		if !seen[it.ID] {
			t.Errorf("faltou %s, que vem antes na ordem de estudo", it.Content.Term)
		}
	}
	if !pack.ExpiresAt.After(now) || len(s.Cards) != 9 {
		t.Errorf("validade e cartoes criados: %v %d", pack.ExpiresAt, len(s.Cards))
	}
}

func contains(opts []domain.Option, item uuid.UUID) bool {
	for _, o := range opts {
		if o.ItemID == item {
			return true
		}
	}
	return false
}

func TestPlanModes(t *testing.T) {
	s := store("house", "cat", "dog")
	pack, _ := planner(s, student).Run(context.Background(), app.PlanRequest{Minutes: 15, Mode: domain.ModeReview})
	if len(pack.Cards) != 0 || pack.NewLimit != 0 {
		t.Errorf("modo revisao nao traz palavras novas: %+v", pack)
	}
	pack, _ = planner(s, student).Run(context.Background(), app.PlanRequest{Minutes: 15, Mode: domain.ModeLearn})
	if len(pack.Cards) != 3 || pack.ReviewLimit != 0 {
		t.Errorf("modo aprender so traz novas: %+v", pack)
	}
}

func TestPlanHidesDraftsFromStudentsButNotFromAdmins(t *testing.T) {
	items := words("house", "cat")
	items[1].Draft = true
	for who, want := range map[string]int{"aluno": 1, "admin": 2} {
		viewer := domain.Viewer{UserID: user, Admin: who == "admin"}
		pack, err := planner(learningtest.New(items...), viewer).Run(context.Background(), app.PlanRequest{Minutes: 15})
		if err != nil || len(pack.Cards) != want {
			t.Errorf("%s: %d cartoes (esperava %d) %v", who, len(pack.Cards), want, err)
		}
	}
}

func TestPlanRemembersAValidTimezoneOnly(t *testing.T) {
	s := store("house")
	p := planner(s, student)
	_, _ = p.Run(context.Background(), app.PlanRequest{Minutes: 5, TZ: "Europe/Lisbon"})
	if s.Settings.TZ != "Europe/Lisbon" {
		t.Errorf("fuso valido fica guardado: %s", s.Settings.TZ)
	}
	_, _ = p.Run(context.Background(), app.PlanRequest{Minutes: 5, TZ: "Marte/Olympus"})
	if s.Settings.TZ != "Europe/Lisbon" {
		t.Errorf("fuso invalido e ignorado: %s", s.Settings.TZ)
	}
}

func addDue(s *learningtest.Store, item uuid.UUID, stability float64, reviewedDaysAgo, dueDaysAgo float64) *domain.Card {
	id := uuid.New()
	c := &domain.Card{ID: id, ItemID: item, Direction: domain.Receptive, Scheduler: "simple_v1", State: srs.CardState{
		ID: id.String(), State: srs.StateReview, Stability: stability, Reps: 3,
		LastReviewAt: now.Add(-time.Duration(reviewedDaysAgo * 24 * float64(time.Hour))),
		DueAt:        now.Add(-time.Duration(dueDaysAgo * 24 * float64(time.Hour))),
	}}
	s.Cards[id] = c
	return c
}

// Com mais vencidas do que cabem (5 minutos = 20 revisoes), ficam as de menor retencao; a de R mais alta sai.
func TestPlanKeepsTheLowestRetrievabilityDueCardsWithinTheBudget(t *testing.T) {
	s := learningtest.New(many(21)...)
	fresh := addDue(s, s.Items[0].ID, 5, 6, 1) // revisado ha 6 dias com S=5: R mais alta
	for _, it := range s.Items[1:] {
		addDue(s, it.ID, 2, 20, 18) // revisados ha 20 dias com S=2: R baixissima
	}
	pack, err := planner(s, student).Run(context.Background(), app.PlanRequest{Minutes: 5, Mode: domain.ModeReview})
	if err != nil || len(pack.Cards) != 20 {
		t.Fatalf("revisoes: %d %v", len(pack.Cards), err)
	}
	for _, c := range pack.Cards {
		if c.Card.ID == fresh.ID {
			t.Error("a revisao de R mais alta fica para depois")
		}
		if c.Mastery <= 0 || c.Mastery >= 1 {
			t.Errorf("dominio entre 0 e 1: %v", c.Mastery)
		}
	}
}

func TestPlanLimitsReviewsToTheTimeBudgetAndHalvesNewWordsOnAvalanche(t *testing.T) {
	s := learningtest.New(words("house")...)
	for i := 0; i < 50; i++ { // 5 minutos = 20 revisoes; 50 vencidas > 2 x 20: avalanche
		addDue(s, uuid.New(), 2, 10, 5)
	}
	pack, _ := planner(s, student).Run(context.Background(), app.PlanRequest{Minutes: 5})
	if pack.ReviewLimit != 20 || pack.NewLimit != 4 {
		t.Fatalf("na avalanche as novas caem pela metade (9 -> 4): %+v", pack)
	}
	// as vencidas que nao tem item no catalogo ficam de fora do pacote (conteudo indisponivel); sobra a nova
	if len(pack.Cards) != 1 || pack.Cards[0].Content.Term != "house" {
		t.Errorf("cartoes sem conteudo nao entram: %d", len(pack.Cards))
	}
}

func TestPlanCreatesTheProductiveCardOnlyAfterTheReceptiveIsStrongAndOldEnough(t *testing.T) {
	s := store("house", "cat")
	strong := addDue(s, s.Items[0].ID, 10, 3, -5) // S >= 7 e revisado ha 3 dias (nao esta vencido)
	weak := addDue(s, s.Items[1].ID, 3, 3, -1)    // S < 7
	justNow := addDue(s, uuid.New(), 12, 0, -5)   // forte, mas revisado agora ha pouco: nunca na mesma sessao
	_ = weak
	_ = justNow
	pack, err := planner(s, student).Run(context.Background(), app.PlanRequest{Minutes: 15, Mode: domain.ModeLearn})
	if err != nil {
		t.Fatal(err)
	}
	var productive []domain.PackCard
	for _, c := range pack.Cards {
		if c.Card.Direction == domain.Productive {
			productive = append(productive, c)
		}
	}
	if len(productive) != 1 || productive[0].Card.ItemID != strong.ItemID || productive[0].Card.Exercise() != domain.Type || len(productive[0].Options) != 0 {
		t.Fatalf("so o forte ganha o produtivo (digitar, sem alternativas): %+v", productive)
	}
}

// Os produtivos tem cota propria: nao tiram a vaga das palavras novas.
func TestPlanProductiveCardsDoNotTakeTheNewWordsSlots(t *testing.T) {
	s := learningtest.New(many(30)...)
	for _, it := range s.Items[:10] {
		addDue(s, it.ID, 10, 3, -5) // receptivos fortes, nao vencidos: viram produtivos
	}
	pack, err := planner(s, student).Run(context.Background(), app.PlanRequest{Minutes: 5, Mode: domain.ModeLearn})
	if err != nil {
		t.Fatal(err)
	}
	count := map[domain.Direction]int{}
	for _, c := range pack.Cards {
		count[c.Card.Direction]++
	}
	if count[domain.Receptive] != 9 || count[domain.Productive] != 5 {
		t.Errorf("5 minutos: 9 novas e 5 produtivos: %+v", count)
	}
}

func TestPlanRejectsBadInputAndPropagatesFailures(t *testing.T) {
	p := planner(store("house"), student)
	for _, m := range []int{4, 241} {
		if _, err := p.Run(context.Background(), app.PlanRequest{Minutes: m}); !errors.Is(err, domain.ErrInvalidMinutes) {
			t.Errorf("minutos invalidos (%d): %v", m, err)
		}
	}
	if _, err := p.Run(context.Background(), app.PlanRequest{Minutes: 5, Mode: "tudo"}); !errors.Is(err, domain.ErrInvalidMode) {
		t.Errorf("modo invalido: %v", err)
	}
	for _, method := range []string{"Settings", "DueCards", "NewItems", "StretchItems", "ProductiveItems", "CreateCard", "CreateSession", "Contents", "Distractors"} {
		s := store("house", "cat")
		addDue(s, s.Items[0].ID, 10, 3, 1) // garante que ProductiveItems/DueCards sao usados
		s.FailOn = method
		if _, err := planner(s, student).Run(context.Background(), app.PlanRequest{Minutes: 15}); !errors.Is(err, learningtest.ErrBoom) {
			t.Errorf("falha em %s deveria subir: %v", method, err)
		}
	}
	s := store("house")
	s.FailOn = "CreateSession"
	_, _ = planner(s, student).Run(context.Background(), app.PlanRequest{Minutes: 15})
	if len(s.Cards) != 0 {
		t.Error("se o plano falha, nenhum cartao novo fica criado (transacao)")
	}
}
