package sim

import (
	"sort"
	"strconv"
	"time"

	"github.com/keven-rdr/wordloop/api/internal/srs"
)

// Config define o cenario: quantos dias, palavras novas por dia, revisoes por dia (o tempo do aluno) e a semente.
type Config struct {
	Days      int
	NewPerDay int
	Budget    int // maximo de revisoes (sem contar as novas) num dia de estudo
	Seed      uint64
}

// DefaultConfig e o cenario de referencia: 90 dias, 8 novas por dia (teto inicial A1), ate 60 revisoes por dia.
func DefaultConfig() Config { return Config{Days: 90, NewPerDay: 8, Budget: 60, Seed: 1} }

var epoch = time.Date(2026, 1, 1, 10, 0, 0, 0, time.UTC)

func at(day int) time.Time { return epoch.AddDate(0, 0, day) }

// world guarda o estado de uma simulacao: o aluno, a estrategia e os cartoes (o estado do modelo).
type world struct {
	cfg   Config
	sched srs.Scheduler
	st    *student
	cards map[string]srs.CardState
	order []string
	res   *Result
}

// Run simula um aluno estudando com uma estrategia e devolve as metricas.
func Run(cfg Config, sched srs.Scheduler, p Profile) Result {
	w := &world{cfg: cfg, sched: sched, st: newStudent(p, cfg.Seed), cards: map[string]srs.CardState{}}
	w.res = &Result{Strategy: sched.Name(), Profile: p.Name, Days: make([]DayStat, cfg.Days)}
	for d := 0; d < cfg.Days; d++ {
		w.day(d)
		w.res.snapshot(w, d)
	}
	w.res.Total = len(w.order)
	return *w.res
}

// day simula um dia: se o aluno estuda, faz as revisoes vencidas (menor R primeiro, ate o orcamento) e
// aprende palavras novas, salvo se a fila de revisoes passar de 2x a meta (protecao contra avalanche).
func (w *world) day(d int) {
	if !w.st.studies() {
		return
	}
	now := at(d)
	due := w.due(now)
	stat := &w.res.Days[d]
	stat.Studied = true
	for _, id := range due[:min(len(due), w.cfg.Budget)] {
		w.review(id, now, stat)
	}
	if len(due) > 2*w.cfg.Budget {
		return
	}
	for i := 0; i < w.cfg.NewPerDay; i++ {
		id := "item-" + strconv.Itoa(len(w.order))
		w.order = append(w.order, id)
		w.cards[id] = srs.CardState{ID: id, State: srs.StateNew}
		w.st.learn(id, now)
		w.review(id, now, stat)
		stat.New++
	}
}

// due lista os cartoes vencidos do menor R para o maior (e o id, para ficar determinista).
func (w *world) due(now time.Time) []string {
	var ids []string
	for _, id := range w.order {
		if c := w.cards[id]; c.State != srs.StateNew && !c.DueAt.After(now) {
			ids = append(ids, id)
		}
	}
	sort.Slice(ids, func(i, j int) bool {
		ri, rj := w.sched.Retrievability(w.cards[ids[i]], now), w.sched.Retrievability(w.cards[ids[j]], now)
		if ri != rj {
			return ri < rj
		}
		return ids[i] < ids[j]
	})
	return ids
}

func (w *world) review(id string, now time.Time, stat *DayStat) {
	c := w.cards[id]
	predicted := w.sched.Retrievability(c, now)
	r, _ := w.st.answer(id, c, now)
	next, err := w.sched.Next(c, r, now)
	if err != nil {
		panic("sim: " + err.Error()) // o simulador so gera respostas validas
	}
	w.cards[id] = next
	w.res.record(c.State, predicted, r, next.Lapses > c.Lapses, stat)
}
