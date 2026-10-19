package srs

import (
	"math"
	"testing"
)

func near(t *testing.T, name string, got, want, tol float64) {
	t.Helper()
	if math.Abs(got-want) > tol {
		t.Errorf("%s = %v, esperava %v (+-%v)", name, got, want, tol)
	}
}

func TestRetrievabilityIs90PercentAtS(t *testing.T) {
	for _, s := range []float64{0.5, 1, 2, 9, 30, 365} {
		near(t, "R(S,S)", Retrievability(s, s), 0.9, 1e-9)
		near(t, "R(0,S)", Retrievability(0, s), 1, 1e-12)
		near(t, "Interval(0,9;S)", Interval(0.9, s), s, 1e-9)
	}
}

func TestRetrievabilityEdges(t *testing.T) {
	if Retrievability(5, 0) != 0 {
		t.Error("sem estabilidade nao ha retencao")
	}
	if Retrievability(-3, 2) != 1 {
		t.Error("tempo negativo conta como zero")
	}
}

func TestMasteryTable(t *testing.T) {
	// docs/fase-1/01-srs.md, secao 4.1: S = 2 -> 21%; 9 -> 44%; 30 -> 66%; 90 -> 87%
	for s, want := range map[float64]float64{2: 0.21, 9: 0.44, 30: 0.66, 90: 0.87, 0: 0, 180: 1, 5000: 1} {
		near(t, "Mastery", Mastery(s), want, 0.005)
	}
}

func TestMasteredAndLeech(t *testing.T) {
	if !(CardState{State: StateReview, Stability: 90, SinceLapse: 2}).Mastered() {
		t.Error("S >= 90 sem lapso nas 2 ultimas revisoes e dominado")
	}
	for _, c := range []CardState{
		{State: StateReview, Stability: 89, SinceLapse: 5},
		{State: StateReview, Stability: 200, SinceLapse: 1},
		{State: StateRelearning, Stability: 200, SinceLapse: 5},
	} {
		if c.Mastered() {
			t.Errorf("nao deveria estar dominado: %+v", c)
		}
	}
	if (CardState{Lapses: 5}).Leech() || !(CardState{Lapses: 6}).Leech() {
		t.Error("sanguessuga com 6 lapsos")
	}
}
