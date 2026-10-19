package srs

import (
	"math"
	"time"
)

const (
	decay     = 0.1542 // w20 do FSRS-6
	masteryAt = 180.0  // S que da 100% de dominio [PREMISSA: calibrar]
	masteredS = 90.0   // "dominado": S >= 90 dias e sem lapso nas 2 ultimas revisoes
	leechAt   = 6      // sanguessuga: 6 lapsos
)

// factor e o F de R(t,S) = (1 + F*t/S)^-d, com F = 0,9^(-1/d) - 1: assim R(S,S) = 90%.
var factor = math.Pow(0.9, -1/decay) - 1

// Retrievability e a probabilidade estimada de lembrar apos t dias para a estabilidade s.
func Retrievability(t, s float64) float64 {
	if s <= 0 {
		return 0
	}
	return math.Pow(1+factor*math.Max(t, 0)/s, -decay)
}

// Interval devolve em dias o tempo ate a retencao cair para r (com r = 0,9 vale S).
func Interval(r, s float64) float64 {
	return s / factor * (math.Pow(r, -1/decay) - 1)
}

// Mastery e o "dominio" mostrado por item: ln(1+S)/ln(1+180), entre 0 e 1.
func Mastery(s float64) float64 {
	return math.Min(1, math.Max(0, math.Log1p(s)/math.Log1p(masteryAt)))
}

// Mastered diz se o cartao esta dominado (continua voltando, com intervalos longos).
func (c CardState) Mastered() bool {
	return c.State == StateReview && c.Stability >= masteredS && c.SinceLapse >= 2
}

// Leech diz se o cartao e uma sanguessuga: a interface oferece trocar imagem ou frase, nunca suspende em silencio.
func (c CardState) Leech() bool { return c.Lapses >= leechAt }

func daysBetween(from, to time.Time) float64 {
	return to.Sub(from).Hours() / 24
}
