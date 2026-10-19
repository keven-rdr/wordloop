package sim_test

import (
	"bytes"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/keven-rdr/wordloop/api/internal/srs"
	"github.com/keven-rdr/wordloop/api/internal/srs/sim"
)

func profile(name string) sim.Profile {
	for _, p := range sim.Profiles() {
		if p.Name == name {
			return p
		}
	}
	panic("perfil inexistente: " + name)
}

func TestRunIsDeterministicAndSeedChangesIt(t *testing.T) {
	cfg := sim.DefaultConfig()
	cfg.Days = 40
	a := sim.Run(cfg, srs.SimpleV1{}, profile("mediano"))
	b := sim.Run(cfg, srs.SimpleV1{}, profile("mediano"))
	if a.Reviews != b.Reviews || a.Correct != b.Correct || a.Lapses != b.Lapses || a.Total != b.Total {
		t.Fatalf("mesma semente, mesmo resultado: %+v %+v", a, b)
	}
	cfg.Seed = 99
	c := sim.Run(cfg, srs.SimpleV1{}, profile("mediano"))
	if c.Reviews == a.Reviews && c.Correct == a.Correct && c.Total == a.Total {
		t.Error("outra semente deveria dar outro caminho")
	}
}

func TestResultInvariants(t *testing.T) {
	cfg := sim.DefaultConfig()
	for _, r := range sim.Compare(cfg, []srs.Scheduler{srs.SimpleV1{}, sim.Ladder{}}) {
		if len(r.Days) != cfg.Days || r.Correct > r.Reviews || r.Retention() < 0 || r.Retention() > 1 {
			t.Fatalf("%s/%s: resultado incoerente: %+v", r.Strategy, r.Profile, r)
		}
		if r.Known30 > r.Total || r.Known90 > r.Total || r.Total > cfg.Days*cfg.NewPerDay {
			t.Errorf("%s/%s: sabidas %d/%d de %d palavras", r.Strategy, r.Profile, r.Known30, r.Known90, r.Total)
		}
		sum, hits := 0, 0
		for _, b := range r.Calibration {
			sum += b.Count
			hits += b.Hits
			if b.Hits > b.Count {
				t.Errorf("%s/%s: acertos acima das revisoes na faixa", r.Strategy, r.Profile)
			}
		}
		if sum != r.Reviews || hits != r.Correct {
			t.Errorf("%s/%s: calibracao %d/%d x revisoes %d/%d", r.Strategy, r.Profile, sum, hits, r.Reviews, r.Correct)
		}
		for i, d := range r.Days {
			if d.Reviews > cfg.Budget || d.Correct > d.Reviews || (!d.Studied && d.Reviews+d.New > 0) {
				t.Fatalf("%s/%s dia %d: %+v", r.Strategy, r.Profile, i, d)
			}
		}
		if r.MaxDue() < 0 || r.PerStudyDay() > float64(cfg.Budget) {
			t.Errorf("%s/%s: carga fora do orcamento", r.Strategy, r.Profile)
		}
	}
}

func TestMoreDiligenceMeansMoreReviews(t *testing.T) {
	cfg := sim.DefaultConfig()
	lazy, keen := profile("mediano"), profile("mediano")
	lazy.Diligence, keen.Diligence = 0.2, 1.0
	if l, k := sim.Run(cfg, srs.SimpleV1{}, lazy), sim.Run(cfg, srs.SimpleV1{}, keen); l.Reviews >= k.Reviews || l.Total >= k.Total {
		t.Errorf("estudar mais dias deveria dar mais revisoes e palavras: %d/%d x %d/%d", l.Reviews, l.Total, k.Reviews, k.Total)
	}
}

func TestAvalancheStopsNewWords(t *testing.T) {
	cfg := sim.DefaultConfig()
	cfg.Budget = 5 // so 5 revisoes por dia: a fila cresce e passa de 2x a meta
	r := sim.Run(cfg, srs.SimpleV1{}, profile("cuidadoso"))
	if r.Total >= cfg.Days*cfg.NewPerDay {
		t.Errorf("com a fila acima de 2x a meta nao entram palavras novas: %d de %d", r.Total, cfg.Days*cfg.NewPerDay)
	}
}

// Guarda de regressao do cenario de referencia (semente fixa): o simple_v1 retem mais que a escada fixa para
// quem erra a autoavaliacao e estuda menos, e a retencao do aluno cuidadoso fica perto do alvo de 90%.
func TestSimpleV1AgainstTheLadder(t *testing.T) {
	cfg := sim.DefaultConfig()
	for _, name := range []string{"mediano", "irregular"} {
		s, l := sim.Run(cfg, srs.SimpleV1{}, profile(name)), sim.Run(cfg, sim.Ladder{}, profile(name))
		if s.Retention() <= l.Retention() {
			t.Errorf("%s: simple_v1 %.3f deveria reter mais que a escada %.3f", name, s.Retention(), l.Retention())
		}
	}
	if r := sim.Run(cfg, srs.SimpleV1{}, profile("cuidadoso")).Retention(); r < 0.88 || r > 0.98 {
		t.Errorf("retencao do aluno cuidadoso fora de 0,88-0,98: %.3f", r)
	}
}

func TestReportListsEveryRunAndItsCalibration(t *testing.T) {
	cfg := sim.DefaultConfig()
	cfg.Days = 30
	var buf bytes.Buffer
	sim.Report(&buf, sim.Compare(cfg, []srs.Scheduler{srs.SimpleV1{}, sim.Ladder{}}))
	out := buf.String()
	for _, want := range []string{"estrategia", "simple_v1", "ladder", "cuidadoso", "irregular", "calibracao simple_v1 / mediano", "observado"} {
		if !strings.Contains(out, want) {
			t.Errorf("relatorio sem %q:\n%s", want, out)
		}
	}
	var empty bytes.Buffer
	sim.Report(&empty, nil)
	if !strings.Contains(empty.String(), "estrategia") {
		t.Error("sem resultados ainda imprime o cabecalho")
	}
	if (sim.Result{}).Retention() != 0 || (sim.Result{}).PerStudyDay() != 0 {
		t.Error("resultado vazio nao divide por zero")
	}
}

func TestLadder(t *testing.T) {
	l := sim.Ladder{}
	now := time.Date(2026, 10, 1, 4, 0, 0, 0, time.UTC)
	ok := srs.Review{Correct: true, Difficulty: srs.DifficultyMedium}
	c := srs.CardState{ID: "x", State: srs.StateNew}
	for _, want := range []float64{1, 3, 7, 14, 30, 60, 120, 120} {
		c, _ = l.Next(c, ok, now)
		if c.Stability != want || c.State != srs.StateReview {
			t.Fatalf("degrau esperado %v, veio %v", want, c.Stability)
		}
		now = now.Add(24 * time.Hour)
	}
	wrong, err := l.Next(c, srs.Review{Correct: false}, now)
	if err != nil || wrong.Stability != 1 || wrong.Lapses != 1 || wrong.SinceLapse != 0 {
		t.Errorf("errar volta ao 1o degrau e conta lapso: %+v %v", wrong, err)
	}
	first, _ := l.Next(srs.CardState{State: srs.StateNew}, srs.Review{Correct: false}, now)
	if first.Lapses != 0 {
		t.Error("o 1o erro nao e lapso")
	}
	if _, err := l.Next(c, srs.Review{Correct: true}, now); !errors.Is(err, srs.ErrInvalidReview) {
		t.Errorf("acerto sem dificuldade: %v", err)
	}
	if l.Retrievability(srs.CardState{}, now) != 0 || l.Retrievability(c, c.LastReviewAt) != 1 {
		t.Error("retencao: novo = 0; no instante da revisao = 1")
	}
	if l.Name() != "ladder" {
		t.Error("nome")
	}
}

type brokenScheduler struct{ sim.Ladder }

func (brokenScheduler) Next(c srs.CardState, _ srs.Review, _ time.Time) (srs.CardState, error) {
	return c, errors.New("quebrou")
}

func TestRunPanicsIfTheSchedulerRejectsAValidAnswer(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Error("o simulador so gera respostas validas: erro do agendador e bug")
		}
	}()
	sim.Run(sim.DefaultConfig(), brokenScheduler{}, profile("cuidadoso"))
}
