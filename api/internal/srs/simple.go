package srs

import (
	"math"
	"time"
)

const (
	maxStability = 365.0
	lapseFactor  = 0.25 // erro: S' = max(1, 0,25*S)
	lateCredit   = 0.5  // metade do atraso conta
	growthCapN   = 5    // nas 5 primeiras revisoes S' <= 3*S [PREMISSA: Fase 1, secao 6]
	growthCap    = 3.0
)

var (
	firstStability = map[Grade]float64{Hard: 1, Good: 2, Easy: 4}
	multiplier     = map[Grade]float64{Hard: 1.2, Good: 2.0, Easy: 3.0}
)

// SimpleV1 e a regra simples e explicavel do MVP (cada numero cabe numa tabela e pode ser ajustado).
type SimpleV1 struct {
	// NoFuzz desliga o espalhamento dos vencimentos (testes e simulacoes).
	NoFuzz bool
}

var _ Scheduler = SimpleV1{}

func (SimpleV1) Name() string { return "simple_v1" }

// Retrievability usa a curva unica compartilhada por todas as estrategias.
func (SimpleV1) Retrievability(c CardState, now time.Time) float64 {
	if c.LastReviewAt.IsZero() {
		return 0
	}
	return Retrievability(daysBetween(c.LastReviewAt, now), c.Stability)
}

// Next aplica a resposta ao cartao: pura e determinista.
func (s SimpleV1) Next(c CardState, r Review, now time.Time) (CardState, error) {
	g, err := r.Grade()
	if err != nil {
		return c, err
	}
	if !c.LastReviewAt.IsZero() && now.Before(c.LastReviewAt) {
		return c, ErrTimeTravel
	}
	next := c
	next.Reps++
	next.LastReviewAt = now
	switch {
	case c.State == StateNew || c.State == "" || c.State == StateLearning:
		first(&next, c, g)
	case g.Lapse():
		next.Stability = math.Max(1, lapseFactor*c.Stability)
		next.Lapses++
		next.SinceLapse = 0
		next.State = StateRelearning
	default:
		next.Stability = grow(c, g, now)
		next.SinceLapse++
		next.State = StateReview
	}
	days := s.days(next)
	next.DueAt = now.Add(time.Duration(days * 24 * float64(time.Hour)))
	return next, nil
}

// first trata a 1a resposta (ou a graduacao de um cartao em aprendizado): so conta como lapso em review.
func first(next *CardState, c CardState, g Grade) {
	if g.Lapse() {
		next.Stability, next.State = 1, StateLearning
		return
	}
	next.Stability = math.Max(c.Stability, firstStability[g])
	next.State = StateReview
	next.SinceLapse = c.SinceLapse + 1
}

// grow e o acerto em review: S' = (S + 0,5*atraso) * m, com teto de crescimento no comeco.
func grow(c CardState, g Grade, now time.Time) float64 {
	late := 0.0
	if !c.DueAt.IsZero() {
		late = math.Max(0, daysBetween(c.DueAt, now))
	}
	s := (c.Stability + lateCredit*late) * multiplier[g]
	if c.Reps < growthCapN {
		s = math.Min(s, growthCap*c.Stability)
	}
	return math.Min(s, maxStability)
}

// days e o intervalo ate o proximo vencimento: S arredondado (meio para cima), minimo 1 dia, com fuzz.
func (s SimpleV1) days(c CardState) float64 {
	d := math.Max(1, math.Floor(c.Stability+0.5))
	if s.NoFuzz {
		return d
	}
	return fuzzed(d, c.ID, c.Reps)
}
