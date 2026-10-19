package sim

import (
	"fmt"
	"io"
	"strings"
	"text/tabwriter"

	"github.com/keven-rdr/wordloop/api/internal/srs"
)

// Bins e o numero de faixas da calibracao (R previsto de 0,0 a 1,0 em passos de 0,1).
const Bins = 10

// DayStat resume um dia da simulacao.
type DayStat struct {
	Studied bool
	Reviews int // revisoes de cartoes ja conhecidos
	Correct int
	New     int // palavras novas apresentadas
	Due     int // vencidas ao fim do dia (a fila de amanha)
}

// Bin e uma faixa da calibracao: quantas revisoes tiveram R previsto nela e quantas foram acertos.
type Bin struct{ Count, Hits int }

// Result reune as metricas de uma simulacao.
type Result struct {
	Strategy, Profile string
	Days              []DayStat
	Reviews, Correct  int // so revisoes de cartoes em review/relearning (retencao real)
	Lapses            int
	Known30, Known90  int // palavras que o aluno lembraria agora com >= 90% (verdade), no dia 30 e 90
	Total             int // palavras apresentadas
	Calibration       [Bins]Bin
}

func (r *Result) record(before srs.State, predicted float64, rev srs.Review, lapse bool, stat *DayStat) {
	if before == srs.StateNew {
		return
	}
	stat.Reviews++
	r.Reviews++
	b := min(int(predicted*Bins), Bins-1)
	r.Calibration[b].Count++
	if rev.Correct {
		stat.Correct++
		r.Correct++
		r.Calibration[b].Hits++
	}
	if lapse {
		r.Lapses++
	}
}

// snapshot guarda o fim do dia: a fila de amanha e, nos dias 30 e 90, quantas palavras estao "sabidas".
func (r *Result) snapshot(w *world, d int) {
	now := at(d + 1)
	r.Days[d].Due = len(w.due(now))
	switch d + 1 {
	case 30:
		r.Known30 = w.st.known(w.order, now)
	case 90:
		r.Known90 = w.st.known(w.order, now)
	}
}

// Retention e a taxa de acerto real nas revisoes (o alvo do modelo e 0,90).
func (r Result) Retention() float64 {
	if r.Reviews == 0 {
		return 0
	}
	return float64(r.Correct) / float64(r.Reviews)
}

// PerStudyDay e a media de revisoes (sem contar as novas) nos dias em que o aluno estudou.
func (r Result) PerStudyDay() float64 {
	days, n := 0, 0
	for _, d := range r.Days {
		if d.Studied {
			days++
			n += d.Reviews
		}
	}
	if days == 0 {
		return 0
	}
	return float64(n) / float64(days)
}

// MaxDue e a maior fila de revisoes vencidas ao fim de um dia.
func (r Result) MaxDue() int {
	m := 0
	for _, d := range r.Days {
		m = max(m, d.Due)
	}
	return m
}

// Report escreve a tabela comparativa (uma linha por estrategia e aluno) e a calibracao de cada uma.
func Report(w io.Writer, results []Result) {
	out := func(format string, a ...any) { _, _ = fmt.Fprintf(w, format, a...) }
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprintln(tw, "estrategia\taluno\trevisoes/dia\tretencao\tlapsos\tsabidas d30\tsabidas d90\tpalavras\tfila max")
	for _, r := range results {
		_, _ = fmt.Fprintf(tw, "%s\t%s\t%.1f\t%.3f\t%d\t%d\t%d\t%d\t%d\n",
			r.Strategy, r.Profile, r.PerStudyDay(), r.Retention(), r.Lapses, r.Known30, r.Known90, r.Total, r.MaxDue())
	}
	_ = tw.Flush()
	for _, r := range results {
		out("\ncalibracao %s / %s (R previsto -> acerto observado)\n", r.Strategy, r.Profile)
		out("%s\n", r.calibrationLines())
	}
}

func (r Result) calibrationLines() string {
	var b strings.Builder
	for i, c := range r.Calibration {
		if c.Count == 0 {
			continue
		}
		_, _ = fmt.Fprintf(&b, "  %.1f-%.1f  n=%-5d observado %.3f\n", float64(i)/Bins, float64(i+1)/Bins, c.Count, float64(c.Hits)/float64(c.Count))
	}
	return strings.TrimRight(b.String(), "\n")
}
