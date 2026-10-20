package app_test

import (
	"context"
	"errors"
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

func TestPlanCreatesTheNextNewCardsInOrderWithContentAndOptions(t *testing.T) {
	s := store("house", "cat", "dog", "tree", "water")
	pack, err := planner(s, student).Run(context.Background(), app.PlanRequest{Minutes: 15})
	if err != nil {
		t.Fatal(err)
	}
	if pack.NewLimit != 3 || pack.ReviewLimit != 60 || len(pack.Cards) != 3 || len(s.Sessions) != 1 || pack.SessionID != s.Sessions[0] {
		t.Fatalf("15 minutos: 3 novas e 60 revisoes: %+v", pack)
	}
	for i, want := range []string{"house", "cat", "dog"} {
		c := pack.Cards[i]
		if c.Content.Term != want || c.Card.Direction != domain.Receptive || c.Card.State.State != srs.StateNew || c.Card.Exercise() != domain.Choose {
			t.Errorf("carta %d: %+v", i, c)
		}
		if len(c.Options) != 4 || !contains(c.Options, c.Card.ItemID) {
			t.Errorf("a pergunta de escolher tem a resposta certa e 3 distratores: %+v", c.Options)
		}
	}
	if !pack.ExpiresAt.After(now) || len(s.Cards) != 3 {
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

func TestPlanIsStableAndDoesNotDuplicateUnansweredCards(t *testing.T) {
	s := store("house", "cat", "dog", "tree")
	p := planner(s, student)
	first, _ := p.Run(context.Background(), app.PlanRequest{Minutes: 15})
	second, err := p.Run(context.Background(), app.PlanRequest{Minutes: 15})
	if err != nil || len(s.Cards) != 3 || len(second.Cards) != 3 {
		t.Fatalf("o segundo plano reaproveita os cartoes ainda nao respondidos: %d cartoes %v", len(s.Cards), err)
	}
	for i := range first.Cards {
		if first.Cards[i].Card.ID != second.Cards[i].Card.ID {
			t.Errorf("carta %d trocou de id", i)
		}
		for j := range first.Cards[i].Options {
			if first.Cards[i].Options[j].ItemID != second.Cards[i].Options[j].ItemID {
				t.Errorf("a ordem das alternativas muda entre planos (carta %d)", i)
			}
		}
	}
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

func TestPlanOrdersDueCardsByLowestRetrievabilityAndCapsThem(t *testing.T) {
	s := store("house", "cat", "dog")
	fresh := addDue(s, s.Items[0].ID, 5, 6, 1)   // revisado ha 6 dias com S=5: R mais alta
	stale := addDue(s, s.Items[1].ID, 2, 20, 18) // revisado ha 20 dias com S=2: R baixissima
	mid := addDue(s, s.Items[2].ID, 4, 9, 5)
	pack, err := planner(s, student).Run(context.Background(), app.PlanRequest{Minutes: 5, Mode: domain.ModeReview})
	if err != nil || len(pack.Cards) != 3 {
		t.Fatalf("revisoes: %d %v", len(pack.Cards), err)
	}
	got := []uuid.UUID{pack.Cards[0].Card.ID, pack.Cards[1].Card.ID, pack.Cards[2].Card.ID}
	if got[0] != stale.ID || got[1] != mid.ID || got[2] != fresh.ID {
		t.Errorf("a menor retencao vem primeiro: %v", got)
	}
	if pack.Cards[0].Mastery <= 0 || pack.Cards[0].Mastery >= 1 {
		t.Errorf("dominio entre 0 e 1: %v", pack.Cards[0].Mastery)
	}
}

func TestPlanLimitsReviewsToTheTimeBudgetAndStopsNewWordsOnAvalanche(t *testing.T) {
	s := learningtest.New(words("house")...)
	for i := 0; i < 50; i++ { // 5 minutos = 20 revisoes; 50 vencidas > 2 x 20: avalanche
		addDue(s, uuid.New(), 2, 10, 5)
	}
	pack, _ := planner(s, student).Run(context.Background(), app.PlanRequest{Minutes: 5})
	if pack.ReviewLimit != 20 || pack.NewLimit != 0 {
		t.Fatalf("sem novas na avalanche: %+v", pack)
	}
	// as vencidas que nao tem item no catalogo ficam de fora do pacote (conteudo indisponivel)
	if len(pack.Cards) != 0 {
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

func TestPlanRejectsBadInputAndPropagatesFailures(t *testing.T) {
	p := planner(store("house"), student)
	if _, err := p.Run(context.Background(), app.PlanRequest{Minutes: 7}); !errors.Is(err, domain.ErrInvalidMinutes) {
		t.Errorf("minutos invalidos: %v", err)
	}
	if _, err := p.Run(context.Background(), app.PlanRequest{Minutes: 5, Mode: "tudo"}); !errors.Is(err, domain.ErrInvalidMode) {
		t.Errorf("modo invalido: %v", err)
	}
	for _, method := range []string{"Settings", "DueCards", "NewItems", "ProductiveItems", "CreateCard", "CreateSession", "Contents", "Distractors"} {
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
