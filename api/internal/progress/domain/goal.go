package domain

import (
	"errors"
	"slices"
	"sort"
	"time"

	"github.com/google/uuid"
)

// CycleDays e a duracao de um ciclo de meta: a semana, de segunda a domingo (no dia de estudo da pessoa).
const CycleDays = 7

// ErrInvalidGoal e a meta fora dos limites (minutos, dias ou dias da semana).
var ErrInvalidGoal = errors.New("meta invalida")

type Preset string

const (
	Light   Preset = "light"
	Steady  Preset = "steady"
	Intense Preset = "intense"
	Custom  Preset = "custom"
)

func (p Preset) valid() bool { return slices.Contains([]Preset{Light, Steady, Intense, Custom}, p) }

// GoalRequest e o que a pessoa escolhe: o tempo por dia, o mesmo de segunda a domingo. Preset vazio vale custom;
// DaysTarget 0 vale 7 (todos os dias). Nao ha escolha de dias da semana: a meta vale a semana inteira.
type GoalRequest struct {
	Preset       Preset
	DailyMinutes int
	DaysTarget   int
}

// Goal e a meta de um ciclo. Os dias da semana vao de 1 (segunda) a 7 (domingo).
type Goal struct {
	ID                       uuid.UUID
	PeriodStart, PeriodEnd   time.Time
	Preset                   Preset
	DailyMinutes, DaysTarget int
	Weekdays                 []int
}

var allWeekdays = []int{1, 2, 3, 4, 5, 6, 7}

// WeekStart e a segunda-feira da semana do dia dado (datas a meia-noite UTC, como o study_day).
func WeekStart(day time.Time) time.Time {
	back := (int(day.Weekday()) + 6) % 7 // segunda = 0 ... domingo = 6
	return AddDays(day, -back)
}

// NewGoal valida o pedido e monta a meta da semana de today: de segunda a domingo, todos os dias com o mesmo tempo.
// Aberta no meio da semana, vale do mesmo jeito ate domingo; na segunda seguinte o app pergunta de novo.
func NewGoal(today time.Time, req GoalRequest) (Goal, error) {
	start := WeekStart(today)
	g := Goal{
		PeriodStart: start, PeriodEnd: AddDays(start, CycleDays-1),
		Preset: req.Preset, DailyMinutes: req.DailyMinutes, DaysTarget: req.DaysTarget, Weekdays: slices.Clone(allWeekdays),
	}
	if g.Preset == "" {
		g.Preset = Custom
	}
	if g.DaysTarget == 0 {
		g.DaysTarget = CycleDays
	}
	if !g.valid() {
		return Goal{}, ErrInvalidGoal
	}
	return g, nil
}

func (g Goal) valid() bool {
	return g.Preset.valid() && g.DailyMinutes >= 1 && g.DailyMinutes <= maxMinutes && g.DaysTarget >= 1 && g.DaysTarget <= CycleDays
}

// Ended diz se o ciclo ja acabou (hoje e depois do ultimo dia).
func (g Goal) Ended(today time.Time) bool { return today.After(g.PeriodEnd) }

// GoalProgress e o quanto do ciclo foi cumprido. Depois do fim do ciclo e a retrospectiva.
type GoalProgress struct {
	DaysStudied, DaysTarget, MinutesStudied, MinutesGoal, NewWords, DaysLeft int
	Met                                                                      bool
}

// Progress soma os dias do ciclo (so os ate hoje): dias com estudo, tempo ativo e palavras novas.
func (g Goal) Progress(days []Day, today time.Time) GoalProgress {
	last := today
	if g.PeriodEnd.Before(last) {
		last = g.PeriodEnd
	}
	p := GoalProgress{DaysTarget: g.DaysTarget, MinutesGoal: g.DailyMinutes * g.DaysTarget}
	var ms int64
	for _, d := range Window(days, g.PeriodStart, last) {
		if d.Studied() {
			p.DaysStudied++
		}
		ms += d.ActiveMS
		p.NewWords += d.NewItems
	}
	p.MinutesStudied = int((ms + 30_000) / 60_000) // arredonda para o minuto mais proximo
	p.Met = p.DaysStudied >= g.DaysTarget
	if !g.Ended(today) {
		p.DaysLeft = int(g.PeriodEnd.Sub(today).Hours() / 24)
	}
	return p
}

// Suggestion e uma meta sugerida.
type Suggestion struct {
	Preset                   Preset
	DailyMinutes, DaysTarget int
}

const (
	historyDays    = 28
	minHistoryDays = 3 // com menos dias de estudo no periodo, valem os valores iniciais
	stepMinutes    = 5
	maxMinutes     = 240
)

// Suggest propoe tres metas (leve, constante e intensa) a partir do tempo real dos ultimos 28 dias: 0,8x a mediana,
// a mediana e o percentil 75 dos minutos por dia estudado. A meta vale todos os dias da semana (7 dias).
// O segundo retorno diz se partiu do historico; sem ele, valores iniciais fixos (5, 10 e 20 minutos).
func Suggest(days []Day, today time.Time) ([]Suggestion, bool) {
	var minutes []float64
	for _, d := range Window(days, AddDays(today, 1-historyDays), today) {
		if d.Studied() {
			minutes = append(minutes, float64(d.ActiveMS)/60_000)
		}
	}
	if len(minutes) < minHistoryDays {
		return []Suggestion{{Light, 5, CycleDays}, {Steady, 10, CycleDays}, {Intense, 20, CycleDays}}, false
	}
	sort.Float64s(minutes)
	median, p75 := percentile(minutes, 0.5), percentile(minutes, 0.75)
	return []Suggestion{
		{Light, snap(0.8 * median), CycleDays},
		{Steady, snap(median), CycleDays},
		{Intense, snap(p75), CycleDays},
	}, true
}

// percentile usa o metodo do posto mais proximo; a lista ja vem ordenada e nao vazia.
func percentile(sorted []float64, p float64) float64 {
	rank := int(p*float64(len(sorted))+0.999999) - 1
	return sorted[min(len(sorted)-1, max(0, rank))]
}

// snap arredonda para o multiplo de 5 minutos mais proximo, de 5 a 240.
func snap(minutes float64) int {
	n := int(minutes/stepMinutes+0.5) * stepMinutes
	return min(maxMinutes, max(stepMinutes, n))
}
