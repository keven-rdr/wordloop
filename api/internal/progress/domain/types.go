// Package domain guarda os tipos do progresso: dias de estudo, sequencia, previsao e o que a pessoa mais erra.
// Nao faz I/O; so le o que o modulo learning gravou (study_day, cartoes e log).
package domain

import (
	"time"

	"github.com/google/uuid"
)

// Viewer e quem pergunta: so a sessao decide, nunca o pedido.
type Viewer struct{ UserID uuid.UUID }

// Settings sao os ajustes que definem o dia de estudo da pessoa.
type Settings struct {
	TZ         string
	CutoffHour int
}

// DefaultSettings valem para quem ainda nao estudou (os mesmos padroes do modulo learning).
var DefaultSettings = Settings{TZ: "America/Sao_Paulo", CutoffHour: 4}

// Day e a agregacao de um dia de estudo (uma linha de study_day). Date e meia-noite UTC do dia logico.
type Day struct {
	Date       time.Time
	Reviews    int
	Correct    int
	NewItems   int
	ActiveMS   int64
	KnownWords int
}

// Studied diz se a pessoa estudou no dia: o que conta e entrar e responder, nao os minutos.
func (d Day) Studied() bool { return d.Reviews > 0 }

// ForecastDay e quantos cartoes vencem num dia.
type ForecastDay struct {
	Date time.Time
	Due  int
}

// Missed e uma palavra que a pessoa erra, com quantos erros no periodo.
type Missed struct {
	ItemID uuid.UUID
	Term   string
	Gloss  string
	Misses int
}

// Dashboard e o painel da pessoa.
type Dashboard struct {
	Today        time.Time
	Streak       int
	StudiedToday bool
	Accuracy7d   *float64 // nil se nao houve resposta nos ultimos 7 dias
	TodayStats   Day
	Forecast     []ForecastDay
	KnownWords   int
	MostMissed   []Missed
}

// Heatmap e o intervalo pedido com so os dias em que houve estudo.
type Heatmap struct {
	From, To time.Time
	Days     []Day
}
