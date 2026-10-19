package srs

import (
	"testing"
	"time"
)

var epoch = time.Date(2026, 10, 1, 4, 0, 0, 0, time.UTC)

func day(n float64) time.Time { return epoch.Add(time.Duration(n * 24 * float64(time.Hour))) }

type step struct {
	day    float64
	review Review
	s      float64
	due    float64 // dias de intervalo ate o vencimento
	state  State
	lapses int
}

var (
	med      = Review{Correct: true, Difficulty: DifficultyMedium}
	easy     = Review{Correct: true, Difficulty: DifficultyEasy, Production: true}
	hard     = Review{Correct: true, Difficulty: DifficultyHard}
	wrong    = Review{Correct: false, ErrorKind: ErrorNoIdea}
	noFuzzV1 = SimpleV1{NoFuzz: true}
)

// Golden vectors do cenario da Fase 1 (docs/fase-1/srs_reference.py, simple_run; secao 6 do doc de SRS).
// Diferenca conhecida: aqui o intervalo arredonda "meio para cima" (4,5 -> 5) e o script usa o arredondamento
// do Python (4,5 -> 4); por isso a 6a resposta vem um dia depois do vencimento do script e o S coincide.
func TestSimpleV1GoldenScenario(t *testing.T) {
	steps := []step{
		{0, med, 2, 2, StateReview, 0},
		{2, easy, 6, 6, StateReview, 0},
		{11, hard, 9, 9, StateReview, 0}, // 3 dias de atraso: (6 + 0,5*3) * 1,2
		{20, wrong, 2.25, 2, StateRelearning, 1},
		{22, med, 4.5, 5, StateReview, 1},
		{27, med, 9, 9, StateReview, 1},
	}
	c := CardState{ID: "card-1", State: StateNew}
	for i, st := range steps {
		next, err := noFuzzV1.Next(c, st.review, day(st.day))
		if err != nil {
			t.Fatalf("passo %d: %v", i+1, err)
		}
		near(t, "S", next.Stability, st.s, 1e-9)
		near(t, "intervalo", daysBetween(day(st.day), next.DueAt), st.due, 1e-9)
		if next.State != st.state || next.Lapses != st.lapses {
			t.Errorf("passo %d: estado %v/%d lapsos, esperava %v/%d", i+1, next.State, next.Lapses, st.state, st.lapses)
		}
		c = next
	}
}

func TestSimpleV1FirstAnswers(t *testing.T) {
	cases := map[string]struct {
		r Review
		s float64
	}{"dificil": {hard, 1}, "medio": {med, 2}, "facil": {easy, 4}}
	for name, c := range cases {
		next, err := noFuzzV1.Next(CardState{ID: "x", State: StateNew}, c.r, day(0))
		if err != nil || next.Stability != c.s || next.State != StateReview || next.Lapses != 0 {
			t.Errorf("1o acerto %s: %+v %v", name, next, err)
		}
	}
	c, _ := noFuzzV1.Next(CardState{ID: "x", State: StateNew}, wrong, day(0))
	if c.State != StateLearning || c.Stability != 1 || c.Lapses != 0 {
		t.Errorf("erro no 1o contato nao e lapso: %+v", c)
	}
	c, _ = noFuzzV1.Next(c, med, day(1))
	if c.State != StateReview || c.Stability != 2 {
		t.Errorf("acerto em aprendizado gradua: %+v", c)
	}
}

func TestSimpleV1GrowthCapAndCeiling(t *testing.T) {
	c := CardState{ID: "x", State: StateReview, Stability: 10, Reps: 2, LastReviewAt: day(0), DueAt: day(10)}
	next, _ := noFuzzV1.Next(c, easy, day(100)) // muito atrasado: (10 + 45) * 3 = 165, mas o teto e 3x
	if next.Stability != 30 {
		t.Errorf("teto de crescimento nas 5 primeiras revisoes: S = %v", next.Stability)
	}
	c.Reps = 9
	c.Stability = 300
	next, _ = noFuzzV1.Next(c, easy, day(100))
	if next.Stability != 365 {
		t.Errorf("S nunca passa de 365: %v", next.Stability)
	}
}

func TestSimpleV1ErrorFloorAndSlipIsNotALapse(t *testing.T) {
	c := CardState{ID: "x", State: StateReview, Stability: 2, Reps: 3, LastReviewAt: day(0), DueAt: day(2)}
	next, _ := noFuzzV1.Next(c, wrong, day(2))
	if next.Stability != 1 || next.Lapses != 1 {
		t.Errorf("erro: S = max(1, 0,25*S): %+v", next)
	}
	slip := Review{Correct: false, ErrorKind: ErrorSlip}
	next, _ = noFuzzV1.Next(c, slip, day(2))
	near(t, "S da distracao", next.Stability, 2.4, 1e-9)
	if next.Lapses != 0 || next.State != StateReview {
		t.Errorf("distracao vale Hard sem lapso: %+v", next)
	}
}

func TestSimpleV1RejectsBadInput(t *testing.T) {
	c := CardState{ID: "x", State: StateReview, Stability: 2, Reps: 1, LastReviewAt: day(5), DueAt: day(7)}
	if _, err := noFuzzV1.Next(c, med, day(4)); err != ErrTimeTravel {
		t.Errorf("resposta antes da ultima revisao: %v", err)
	}
	if _, err := noFuzzV1.Next(c, Review{Correct: true}, day(6)); err == nil {
		t.Error("acerto sem dificuldade e invalido")
	}
	if noFuzzV1.Name() != "simple_v1" {
		t.Error("nome da estrategia")
	}
}

func TestSimpleV1Retrievability(t *testing.T) {
	if noFuzzV1.Retrievability(CardState{}, day(3)) != 0 {
		t.Error("cartao novo nao tem retencao")
	}
	c, _ := noFuzzV1.Next(CardState{ID: "x", State: StateNew}, med, day(0))
	near(t, "R no vencimento", noFuzzV1.Retrievability(c, day(2)), 0.9, 1e-9)
}

func TestFuzzIsDeterministicAndBounded(t *testing.T) {
	for _, days := range []float64{1, 2, 2.4} {
		if fuzzed(days, "x", 3) != days {
			t.Errorf("sem fuzz abaixo de 2,5 dias: %v", days)
		}
	}
	for n := 1; n <= 200; n++ {
		for _, days := range []float64{3, 6, 10, 19, 20, 60, 365} {
			a, b := fuzzed(days, "carta", n), fuzzed(days, "carta", n)
			pct := fuzzPct(days)
			if a != b || a < 1 || a < days*(1-pct)-0.5 || a > days*(1+pct)+0.5 {
				t.Fatalf("fuzz(%v, %d) = %v (%v)", days, n, a, b)
			}
		}
	}
	same := 0
	for n := 1; n <= 20; n++ {
		if fuzzed(60, "a", n) == fuzzed(60, "b", n) {
			same++
		}
	}
	if same == 20 {
		t.Error("cartoes diferentes deveriam espalhar os vencimentos")
	}
}
