package app_test

import (
	"context"
	"slices"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/keven-rdr/wordloop/api/internal/learning/app"
	"github.com/keven-rdr/wordloop/api/internal/learning/domain"
	"github.com/keven-rdr/wordloop/api/internal/learning/learningtest"
	"github.com/keven-rdr/wordloop/api/internal/srs"
)

const sentenceTerm = "I like the red house."

func sentenceStore() *learningtest.Store {
	return learningtest.New(learningtest.Item{ID: uuid.New(), Content: domain.Content{
		Term: sentenceTerm, Gloss: "Eu gosto da casa vermelha.", Kind: domain.SentenceKind,
	}})
}

func assembled(c *domain.Card, text string, answeredAt time.Time) domain.Answer {
	a := typed(c, text, answeredAt)
	a.Exercise = domain.Assemble
	return a
}

func TestPlanSendsTheTilesOfASentenceInProductiveCards(t *testing.T) {
	s := sentenceStore()
	addDue(s, s.Items[0].ID, 3.5, 3, -5) // receptivo com S >= 3 e revisado ha mais de 1 dia: nasce o produtivo da frase
	pack, err := planner(s, student).Run(context.Background(), app.PlanRequest{Minutes: 15, Mode: domain.ModeLearn})
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range pack.Cards {
		if c.Card.Direction != domain.Productive {
			continue
		}
		want := []string{"I", "like", "the", "red", "house"}
		if c.Exercise != domain.Assemble || len(c.Options) != 0 || c.ClozeText != "" {
			t.Fatalf("frase no produtivo vira montar, sem alternativas: %+v", c)
		}
		if slices.Equal(c.Tiles, want) || len(c.Tiles) != len(want) {
			t.Fatalf("as pecas sao as palavras, fora da ordem certa: %v", c.Tiles)
		}
		return
	}
	t.Fatal("o cartao produtivo da frase deveria nascer com S >= 3")
}

func TestAWordOnlyGetsItsProductiveCardAtTheHigherStability(t *testing.T) {
	s := store("house")
	addDue(s, s.Items[0].ID, 3.5, 3, -5) // S = 3,5: ainda cedo para palavra (7)
	pack, err := planner(s, student).Run(context.Background(), app.PlanRequest{Minutes: 15, Mode: domain.ModeLearn})
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range pack.Cards {
		if c.Card.Direction == domain.Productive {
			t.Fatalf("palavra com S = 3,5 nao ganha o produtivo ainda: %+v", c)
		}
	}
}

func TestSubmitCorrectsAnAssembledSentenceExactlyAndAcceptsEasyBecauseItIsProduction(t *testing.T) {
	s := sentenceStore()
	u := submitter(s, now)
	right := newCard(s, domain.Productive)
	a := assembled(right, "i LIKE the red house", now.Add(-time.Minute))
	a.Difficulty = srs.DifficultyEasy
	res := one(t, u, a)
	if res.Status != domain.Applied || !res.Correct || res.Grade != srs.Easy {
		t.Fatalf("certo, sem diferenciar maiusculas nem pontuacao; Facil so vale em producao: %+v", res)
	}
	typo := newCard(s, domain.Productive)
	res = one(t, u, assembled(typo, "I like the red hose", now.Add(-time.Minute)))
	if res.Status != domain.Applied || res.Correct || res.Grade != srs.Again {
		t.Fatalf("peca errada e erro, nunca 'quase': %+v", res)
	}
	order := newCard(s, domain.Productive)
	res = one(t, u, assembled(order, "like I the red house", now.Add(-time.Minute)))
	if res.Correct {
		t.Fatalf("ordem errada: %+v", res)
	}
	blank := newCard(s, domain.Productive)
	if res = one(t, u, assembled(blank, "", now.Add(-time.Minute))); res.Status != domain.Applied || res.Correct {
		t.Fatalf("em branco e 'nao sei': conta como erro: %+v", res)
	}
	for _, row := range s.Log {
		if !row.Production || row.Exercise != domain.Assemble {
			t.Errorf("o log guarda a variante e que foi producao: %+v", row)
		}
	}
}

func TestSubmitRejectsAssembleOnTheReceptiveCard(t *testing.T) {
	s := sentenceStore()
	receptive := newCard(s, domain.Receptive)
	res := one(t, submitter(s, now), assembled(receptive, sentenceTerm, now.Add(-time.Minute)))
	if res.Status != domain.Rejected || res.Reason != domain.ReasonInvalidAnswer || len(s.Log) != 0 {
		t.Fatalf("montar frase no cartao receptivo: %+v", res)
	}
}
