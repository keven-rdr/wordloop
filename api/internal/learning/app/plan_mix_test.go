package app_test

import (
	"context"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/keven-rdr/wordloop/api/internal/learning/app"
	"github.com/keven-rdr/wordloop/api/internal/learning/domain"
	"github.com/keven-rdr/wordloop/api/internal/learning/learningtest"
)

const (
	seedA = "0b8f3c6e-2d1a-4c55-9a77-1f3e5d7c9b21"
	seedB = "7c2e9a14-5b3d-4f6a-8e01-c4d2b6a8f3e9"
)

// seeded e o planner com ids de sessao (sementes) fixos, na ordem: o teste fica determinista.
func seeded(s *learningtest.Store, seeds ...string) app.PlanSession {
	p := planner(s, student)
	next := 0
	p.NewID = func() uuid.UUID {
		id := uuid.MustParse(seeds[next%len(seeds)])
		next++
		return id
	}
	return p
}

func order(pack domain.Pack) string {
	out := make([]string, 0, len(pack.Cards))
	for _, c := range pack.Cards {
		out = append(out, c.Content.Term)
	}
	return strings.Join(out, ",")
}

// answerAt e a posicao da resposta certa em cada pergunta do pacote, por cartao.
func answerAt(p domain.Pack) map[uuid.UUID]int {
	out := map[uuid.UUID]int{}
	for _, c := range p.Cards {
		for j, o := range c.Options {
			if o.ItemID == c.Card.ItemID {
				out[c.Card.ID] = j
			}
		}
	}
	return out
}

func sameAnswers(a, b domain.Pack) bool {
	wa, wb := answerAt(a), answerAt(b)
	for id, j := range wa {
		if wb[id] != j {
			return false
		}
	}
	return true
}

// Outro plano reaproveita os cartoes ainda nao respondidos (sem duplicar), mas a ordem dos cartoes e das
// alternativas muda de uma sessao para outra; a mesma semente da sempre a mesma ordem.
func TestPlanReusesUnansweredCardsButShufflesEachSession(t *testing.T) {
	s := learningtest.New(many(9)...)
	p := seeded(s, seedA, seedB, seedA)
	ctx := context.Background()
	first, _ := p.Run(ctx, app.PlanRequest{Minutes: 5})
	second, _ := p.Run(ctx, app.PlanRequest{Minutes: 5})
	third, err := p.Run(ctx, app.PlanRequest{Minutes: 5})
	if err != nil || len(s.Cards) != 9 || len(second.Cards) != 9 {
		t.Fatalf("reaproveita os cartoes nao respondidos: %d cartoes %v", len(s.Cards), err)
	}
	if order(first) == order(second) {
		t.Errorf("outra sessao, outra ordem: %s", order(first))
	}
	if order(first) != order(third) || !sameAnswers(first, third) {
		t.Errorf("a mesma semente da a mesma ordem: %s x %s", order(first), order(third))
	}
	if sameAnswers(first, second) {
		t.Error("entre sessoes a resposta certa muda de lugar em alguma pergunta")
	}
}

// 1 em cada 5 palavras novas vem de licoes adiante (o desafio), sem tirar a licao atual da frente.
func TestPlanMixesInStretchWordsFromLaterLessons(t *testing.T) {
	items := many(30)
	s := learningtest.New(items...)
	later := items[25:28]
	for _, it := range later {
		s.Stretch = append(s.Stretch, it.ID)
	}
	pack, err := planner(s, student).Run(context.Background(), app.PlanRequest{Minutes: 15}) // 16 novas: 3 de desafio
	if err != nil || len(pack.Cards) != 16 {
		t.Fatalf("16 novas: %d %v", len(pack.Cards), err)
	}
	got := map[uuid.UUID]bool{}
	for _, c := range pack.Cards {
		got[c.Card.ItemID] = true
	}
	for _, it := range later {
		if !got[it.ID] {
			t.Errorf("a palavra de desafio %s entra na sessao", it.Content.Term)
		}
	}
	for _, it := range items[:13] {
		if !got[it.ID] {
			t.Errorf("as 13 primeiras da licao atual entram: faltou %s", it.Content.Term)
		}
	}
}
