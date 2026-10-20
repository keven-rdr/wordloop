// Package domain guarda os tipos do modulo de aprendizado: cartoes, plano da sessao e respostas.
// O agendamento em si vem da biblioteca pura internal/srs.
package domain

import (
	"time"

	"github.com/google/uuid"

	"github.com/keven-rdr/wordloop/api/internal/srs"
)

// Viewer e quem estuda: so a sessao decide, nunca o pedido.
type Viewer struct {
	UserID uuid.UUID
	Admin  bool
}

// SeesDrafts diz se itens em rascunho entram no estudo: so para administradores (igual ao catalogo).
func (v Viewer) SeesDrafts() bool { return v.Admin }

// Direction e o sentido do cartao: receptivo (ingles -> portugues) ou produtivo (portugues -> ingles).
type Direction string

const (
	Receptive  Direction = "receptive"
	Productive Direction = "productive"
)

// Exercise e o tipo de pergunta: escolher (reconhecimento) ou digitar (producao).
type Exercise string

const (
	Choose Exercise = "choose"
	Type   Exercise = "type"
)

// Mode diz o que a sessao traz: so revisoes, so palavras novas ou as duas coisas.
type Mode string

const (
	ModeReview Mode = "review"
	ModeLearn  Mode = "learn"
	ModeMixed  Mode = "mixed"
)

// Settings sao os ajustes de estudo da pessoa.
type Settings struct {
	TZ         string
	CutoffHour int
	Scheduler  string
}

// Card e o cartao agendado de uma pessoa para um item e uma direcao.
type Card struct {
	ID        uuid.UUID
	ItemID    uuid.UUID
	Direction Direction
	State     srs.CardState
	Scheduler string
}

// Exercise devolve o exercicio do cartao: receptivo escolhe, produtivo digita.
func (c Card) Exercise() Exercise {
	if c.Direction == Productive {
		return Type
	}
	return Choose
}

// Content e o que o cartao mostra, vindo do catalogo (no idioma da pessoa).
type Content struct {
	Term, Gloss, Example, ExampleGloss, ImageKey string
}

// Option e uma alternativa do exercicio de escolher.
type Option struct {
	ItemID uuid.UUID
	Gloss  string
}

// PackCard e um cartao pronto para estudar, com o conteudo (serve ao modo sem rede).
type PackCard struct {
	Card    Card
	Content Content
	Options []Option
	Mastery float64
}

// Pack e a sessao montada: o que estudar agora.
type Pack struct {
	SessionID   uuid.UUID
	Cards       []PackCard
	NewLimit    int
	ReviewLimit int
	ExpiresAt   time.Time
}

// Answer e uma resposta enviada pelo app, ainda sem correcao.
type Answer struct {
	ID               uuid.UUID
	CardID           uuid.UUID
	SessionID        uuid.UUID
	DeviceID         uuid.UUID
	Exercise         Exercise
	Source           string
	ShownAt          time.Time
	AnsweredAt       time.Time
	ResponseMS       int
	Text             string
	ChosenItemID     uuid.UUID
	ClientCorrect    *bool
	HintUsed         bool
	Explained        bool
	ImageBefore      bool
	DifficultyHint   string // dificuldade sugerida pelo app
	Difficulty       srs.Difficulty
	ErrorKind        srs.ErrorKind
	ClientSeq        int64
	HasClientCorrect bool
}

// Status e o destino de cada resposta do lote.
type Status string

const (
	Applied   Status = "applied"
	Duplicate Status = "duplicate"
	Rejected  Status = "rejected"
)

// Reason explica uma rejeicao.
type Reason string

const (
	ReasonCardNotFound  Reason = "card_not_found"
	ReasonInvalidAnswer Reason = "invalid_answer"
	ReasonInvalidTime   Reason = "invalid_time"
)

// Result e o resultado de uma resposta do lote.
type Result struct {
	ID      uuid.UUID
	Status  Status
	Reason  Reason
	Correct bool
	Typo    bool
	Grade   srs.Grade
	Card    *Card
}

// LogRow e a linha gravada no review_log (o que foi decidido no servidor).
type LogRow struct {
	Answer
	UserID                      uuid.UUID
	TZ                          string
	CutoffHour                  int
	LocalDate                   time.Time
	Correct, Typo               bool
	Grade                       srs.Grade
	Before, After               srs.CardState
	RAtReview, ElapsedDays      float64
	ScheduledDays               float64
	Scheduler, SchedulerVersion string
	Production                  bool
}
