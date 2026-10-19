package sim

import (
	"time"

	"github.com/keven-rdr/wordloop/api/internal/srs"
)

// Ladder e a estrategia de comparacao: intervalos fixos (1, 3, 7, 14, 30, 60, 120 dias), sem usar a nota
// de dificuldade; errar volta ao primeiro degrau. Serve de referencia para ver o que o simple_v1 ganha.
type Ladder struct{}

var _ srs.Scheduler = Ladder{}

var rungs = []float64{1, 3, 7, 14, 30, 60, 120}

func (Ladder) Name() string { return "ladder" }

func (Ladder) Retrievability(c srs.CardState, now time.Time) float64 {
	if c.LastReviewAt.IsZero() {
		return 0
	}
	return srs.Retrievability(now.Sub(c.LastReviewAt).Hours()/24, c.Stability)
}

func (Ladder) Next(c srs.CardState, r srs.Review, now time.Time) (srs.CardState, error) {
	if _, err := r.Grade(); err != nil {
		return c, err
	}
	next := c
	next.Reps++
	next.LastReviewAt = now
	next.State = srs.StateReview
	if r.Correct {
		next.Stability = rungAfter(c.Stability)
		next.SinceLapse++
	} else {
		next.Stability = rungs[0]
		if c.State == srs.StateReview || c.State == srs.StateRelearning {
			next.Lapses++
		}
		next.SinceLapse = 0
	}
	next.DueAt = now.Add(time.Duration(next.Stability * 24 * float64(time.Hour)))
	return next, nil
}

// rungAfter e o proximo degrau acima do intervalo atual (o ultimo se ja esta no topo).
func rungAfter(current float64) float64 {
	for _, r := range rungs {
		if r > current {
			return r
		}
	}
	return rungs[len(rungs)-1]
}

// Profiles sao os alunos virtuais do relatorio: cuidadoso, mediano e irregular (estuda pouco e erra a autoavaliacao).
func Profiles() []Profile {
	return []Profile{
		{Name: "cuidadoso", Base: 2.5, Gain: 2.6, Diligence: 0.95, Noise: 0.05},
		{Name: "mediano", Base: 1.8, Gain: 2.2, Diligence: 0.75, Noise: 0.12},
		{Name: "irregular", Base: 1.2, Gain: 1.8, Diligence: 0.45, Noise: 0.25},
	}
}

// Compare roda cada estrategia com cada aluno virtual, com a mesma semente (mesmo acaso de partida).
func Compare(cfg Config, scheds []srs.Scheduler) []Result {
	var out []Result
	for _, p := range Profiles() {
		for _, s := range scheds {
			out = append(out, Run(cfg, s, p))
		}
	}
	return out
}
