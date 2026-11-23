package app_test

import (
	"testing"
	"time"

	"github.com/keven-rdr/wordloop/api/internal/learning/domain"
	"github.com/keven-rdr/wordloop/api/internal/srs"
)

// "Ja sabia" no primeiro contato, com acerto: Easy e a palavra volta so em ~3 semanas. Num cartao ja respondido (sessao
// retomada, outro aparelho) vale como facil, sem perder a resposta; com erro, a dificuldade nao conta; no digitar e
// rejeitado.
func TestSubmitKnownOnFirstSightSendsTheWordFar(t *testing.T) {
	s := store("house", "cat")
	c := newCard(s, domain.Receptive)
	a := choose(c, c.ItemID, now.Add(-time.Minute))
	a.Difficulty = srs.DifficultyKnown
	if res := one(t, submitter(s, now), a); res.Status != domain.Applied || res.Grade != srs.Easy {
		t.Fatalf("ja sabia: %+v", res)
	}
	if got := s.Cards[c.ID].State; got.Stability != srs.KnownStability || got.DueAt.Before(now.Add(20*24*time.Hour)) {
		t.Errorf("volta em ~3 semanas: %+v", got)
	}
	again := choose(c, c.ItemID, now)
	again.Difficulty = srs.DifficultyKnown
	if res := one(t, submitter(s, now), again); res.Status != domain.Applied || res.Grade != srs.Good {
		t.Errorf("num cartao ja respondido vale facil (Good no reconhecimento): %+v", res)
	}
	c2 := newCard(s, domain.Receptive)
	miss := choose(c2, s.Items[1].ID, now.Add(-time.Minute))
	miss.Difficulty = srs.DifficultyKnown
	if res := one(t, submitter(s, now), miss); res.Status != domain.Applied || res.Grade != srs.Again || s.Cards[c2.ID].State.Stability == srs.KnownStability {
		t.Errorf("errou: a dificuldade nao vale, e Again: %+v", res)
	}
	p := newCard(s, domain.Productive)
	typedKnown := typed(p, "house", now.Add(-time.Minute))
	typedKnown.Difficulty = srs.DifficultyKnown
	if res := one(t, submitter(s, now), typedKnown); res.Status != domain.Rejected || res.Reason != domain.ReasonInvalidAnswer {
		t.Errorf("ja sabia nao existe no digitar: %+v", res)
	}
}
