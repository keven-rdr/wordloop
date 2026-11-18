package app_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/keven-rdr/wordloop/api/internal/learning/app"
	"github.com/keven-rdr/wordloop/api/internal/learning/domain"
	"github.com/keven-rdr/wordloop/api/internal/learning/learningtest"
	"github.com/keven-rdr/wordloop/api/internal/srs"
)

// variantAnswer e uma resposta digitada com o exercicio dado.
func variantAnswer(c *domain.Card, ex domain.Exercise, text string, answeredAt time.Time) domain.Answer {
	a := typed(c, text, answeredAt)
	a.Exercise = ex
	return a
}

func TestPlanSendsTheClozeSentenceForFunctionalWords(t *testing.T) {
	item := learningtest.Item{ID: uuid.New(), Content: domain.Content{
		Term: "the", Gloss: "o", Example: "The cat sleeps.", ExampleGloss: "O gato dorme.", Theme: domain.FunctionalTheme,
	}}
	s := learningtest.New(item)
	addDue(s, item.ID, 10, 3, -5) // receptivo forte e antigo: nasce o produtivo
	pack, err := planner(s, student).Run(context.Background(), app.PlanRequest{Minutes: 15, Mode: domain.ModeLearn})
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range pack.Cards {
		if c.Card.Direction != domain.Productive {
			continue
		}
		if c.Exercise != domain.Cloze || c.ClozeText != "____ cat sleeps." || len(c.Options) != 0 {
			t.Fatalf("palavra funcional vira lacuna, sem alternativas: %+v", c)
		}
		return
	}
	t.Fatal("o cartao produtivo deveria estar no plano")
}

func TestPlanKeepsReceptiveCardsAsChoiceEvenWithExampleAndImage(t *testing.T) {
	item := learningtest.Item{ID: uuid.New(), Content: domain.Content{
		Term: "house", Gloss: "casa", Example: "This is a house.", ImageKey: "house.png", Theme: domain.FunctionalTheme,
	}}
	s := learningtest.New(item, learningtest.Item{ID: uuid.New(), Content: domain.Content{Term: "cat", Gloss: "gato"}})
	pack, err := planner(s, student).Run(context.Background(), app.PlanRequest{Minutes: 5, Mode: domain.ModeLearn})
	if err != nil || len(pack.Cards) == 0 {
		t.Fatalf("plano: %v %v", pack, err)
	}
	for _, c := range pack.Cards {
		if c.Exercise != domain.Choose || c.ClozeText != "" || len(c.Options) == 0 {
			t.Errorf("novo e receptivo: escolher, com alternativas: %+v", c)
		}
	}
}

func TestSubmitCorrectsClozeAndImageAnswersLikeTyping(t *testing.T) {
	s := store("house", "cat")
	for _, ex := range []domain.Exercise{domain.Type, domain.Cloze, domain.Image} {
		c := newCard(s, domain.Productive)
		res := one(t, submitter(s, now), variantAnswer(c, ex, "House", now.Add(-time.Minute)))
		if res.Status != domain.Applied || !res.Correct || res.Grade != srs.Good {
			t.Errorf("%s certo: %+v", ex, res)
		}
		wrong := newCard(s, domain.Productive)
		res = one(t, submitter(s, now), variantAnswer(wrong, ex, "cat", now.Add(-time.Minute)))
		if res.Status != domain.Applied || res.Correct || res.Grade != srs.Again {
			t.Errorf("%s errado: %+v", ex, res)
		}
	}
	for _, row := range s.Log {
		if !row.Production {
			t.Errorf("toda variante digitada e producao: %+v", row.Exercise)
		}
	}
}

func TestSubmitRejectsAVariantThatDoesNotFitTheCardDirection(t *testing.T) {
	s := store("house", "cat")
	u := submitter(s, now)
	receptive := newCard(s, domain.Receptive)
	for _, ex := range []domain.Exercise{domain.Cloze, domain.Image} {
		res := one(t, u, variantAnswer(receptive, ex, "house", now.Add(-time.Minute)))
		if res.Status != domain.Rejected || res.Reason != domain.ReasonInvalidAnswer {
			t.Errorf("%s no cartao receptivo: %+v", ex, res)
		}
	}
	productive := newCard(s, domain.Productive)
	res := one(t, u, choose(productive, productive.ItemID, now.Add(-time.Minute)))
	if res.Status != domain.Rejected || res.Reason != domain.ReasonInvalidAnswer {
		t.Errorf("escolher no cartao produtivo: %+v", res)
	}
	if len(s.Log) != 0 {
		t.Error("rejeitada nao grava nada")
	}
}
