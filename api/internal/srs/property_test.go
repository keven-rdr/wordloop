package srs

import (
	"math"
	"math/rand/v2"
	"testing"
)

func randomLog(rng *rand.Rand, n int) []Answer {
	diffs := []Difficulty{DifficultyHard, DifficultyMedium, DifficultyEasy}
	kinds := []ErrorKind{ErrorNoIdea, ErrorAlmost, ErrorSlip}
	log := make([]Answer, n)
	at := 0.0
	for i := range log {
		at += math.Floor(rng.Float64() * 120 * rng.Float64()) // 0 a ~120 dias, mais perto de 0
		r := Review{
			Correct: rng.IntN(100) < 75, Difficulty: diffs[rng.IntN(3)], ErrorKind: kinds[rng.IntN(3)],
			Typo: rng.IntN(10) == 0, Production: rng.IntN(2) == 0,
			Hint: rng.IntN(8) == 0, Explained: rng.IntN(10) == 0, ImageBefore: rng.IntN(10) == 0,
		}
		log[i] = Answer{Review: r, At: day(at)}
	}
	return log
}

// Propriedades: valem para qualquer sequencia de respostas, com fuzz ligado.
func TestSimpleV1Properties(t *testing.T) {
	rng := rand.New(rand.NewPCG(7, 42))
	s := SimpleV1{}
	for run := 0; run < 2000; run++ {
		log := randomLog(rng, 1+rng.IntN(40))
		c := CardState{ID: "carta", State: StateNew}
		for i, a := range log {
			prev := c
			next, err := s.Next(c, a.Review, a.At)
			if err != nil {
				t.Fatalf("run %d passo %d: %v", run, i, err)
			}
			checkInvariants(t, prev, next, a)
			c = next
		}
	}
}

func checkInvariants(t *testing.T, prev, next CardState, a Answer) {
	t.Helper()
	g, _ := a.Grade()
	switch {
	case math.IsNaN(next.Stability) || math.IsInf(next.Stability, 0) || next.Stability <= 0 || next.Stability > 365:
		t.Fatalf("S invalido: %v", next.Stability)
	case !next.DueAt.After(a.At) || daysBetween(a.At, next.DueAt) < 1:
		t.Fatalf("o vencimento tem de ser pelo menos 1 dia depois: %v", next.DueAt)
	case next.Reps != prev.Reps+1:
		t.Fatalf("reps: %d -> %d", prev.Reps, next.Reps)
	case next.Lapses < prev.Lapses:
		t.Fatalf("lapsos nao diminuem: %d -> %d", prev.Lapses, next.Lapses)
	case next.Lapses > prev.Lapses && !lapseAllowed(prev.State, g):
		t.Fatalf("lapso so em erro durante review/relearning: %v %v", prev.State, g)
	case next.State == StateNew:
		t.Fatal("depois de responder o cartao nao e mais novo")
	}
	if r := (SimpleV1{}).Retrievability(next, a.At); r <= 0 || r > 1 {
		t.Fatalf("R fora de (0,1]: %v", r)
	}
}

// lapseAllowed: lapso so em erro, e so em review/relearning (o 1o erro nao e lapso).
func lapseAllowed(from State, g Grade) bool {
	return g.Lapse() && (from == StateReview || from == StateRelearning)
}

// Mais facil nunca cresce menos: com o mesmo estado, Easy >= Good >= Hard >= erro.
func TestSimpleV1GradeOrdering(t *testing.T) {
	rng := rand.New(rand.NewPCG(3, 9))
	s := SimpleV1{NoFuzz: true}
	for i := 0; i < 2000; i++ {
		st := 1 + rng.Float64()*300
		c := CardState{ID: "x", State: StateReview, Stability: st, Reps: rng.IntN(12), LastReviewAt: day(0), DueAt: day(st)}
		now := day(st + rng.Float64()*50)
		var out [4]float64
		for j, r := range []Review{wrong, hard, med, easy} {
			n, _ := s.Next(c, r, now)
			out[j] = n.Stability
		}
		if !(out[0] <= out[1] && out[1] <= out[2] && out[2] <= out[3]) {
			t.Fatalf("ordem quebrada (S=%v): %v", st, out)
		}
	}
}

func TestSimpleV1IsDeterministic(t *testing.T) {
	rng := rand.New(rand.NewPCG(1, 1))
	log := randomLog(rng, 30)
	a, errA := Replay(SimpleV1{}, "c", log)
	b, errB := Replay(SimpleV1{}, "c", log)
	if errA != nil || errB != nil || a != b {
		t.Fatalf("mesma entrada, mesma saida: %+v %+v %v %v", a, b, errA, errB)
	}
}

// Replay do log == estado incremental (o estado guardado e so um cache do log).
func TestReplayEqualsIncrementalState(t *testing.T) {
	rng := rand.New(rand.NewPCG(5, 5))
	s := SimpleV1{}
	for run := 0; run < 300; run++ {
		log := randomLog(rng, 1+rng.IntN(30))
		incremental := CardState{ID: "c", State: StateNew}
		for _, a := range log {
			incremental, _ = s.Next(incremental, a.Review, a.At)
		}
		replayed, err := Replay(s, "c", log)
		if err != nil || replayed != incremental {
			t.Fatalf("replay %+v != incremental %+v (%v)", replayed, incremental, err)
		}
	}
}

func TestReplayStopsOnABadAnswer(t *testing.T) {
	log := []Answer{{Review: med, At: day(5)}, {Review: med, At: day(2)}}
	if _, err := Replay(SimpleV1{}, "c", log); err != ErrTimeTravel {
		t.Errorf("log fora de ordem: %v", err)
	}
	if c, err := Replay(SimpleV1{}, "c", nil); err != nil || c.State != StateNew {
		t.Errorf("log vazio: %+v %v", c, err)
	}
}
