package srs

import (
	"errors"
	"time"
)

// State e o estado do cartao no ciclo de aprendizado.
type State string

const (
	StateNew        State = "new"
	StateLearning   State = "learning"
	StateReview     State = "review"
	StateRelearning State = "relearning"
)

// Difficulty e a dificuldade que a pessoa informa ao acertar.
type Difficulty string

const (
	DifficultyHard   Difficulty = "hard"
	DifficultyMedium Difficulty = "medium"
	DifficultyEasy   Difficulty = "easy"
)

// ErrorKind e o que a pessoa diz depois de errar ("Como foi?").
type ErrorKind string

const (
	ErrorNone   ErrorKind = ""
	ErrorNoIdea ErrorKind = "no_idea" // Nao sabia
	ErrorAlmost ErrorKind = "almost"  // Quase lembrei
	ErrorSlip   ErrorKind = "slip"    // Foi distracao
)

// CardState e o estado agendado de um cartao (usuario, item, direcao). Stability (S) e em dias.
type CardState struct {
	ID           string // semente do fuzz; o id do cartao
	State        State
	Stability    float64
	Reps         int
	Lapses       int
	SinceLapse   int // revisoes desde o ultimo lapso (para "dominado")
	LastReviewAt time.Time
	DueAt        time.Time
}

// Review e uma resposta ja corrigida pelo servidor.
type Review struct {
	Correct     bool
	Difficulty  Difficulty // so vale quando Correct
	ErrorKind   ErrorKind  // so vale quando !Correct
	Typo        bool       // erro de digitacao com distancia de edicao <= 1 (palavra >= 5 letras)
	Production  bool       // digitar/ouvir->digitar; so a producao aceita Easy
	Hint        bool
	Explained   bool
	ImageBefore bool
}

// Scheduler e a estrategia de agendamento. Next e pura: mesma entrada, mesma saida.
type Scheduler interface {
	Name() string
	Next(c CardState, r Review, now time.Time) (CardState, error)
	Retrievability(c CardState, now time.Time) float64
}

var (
	ErrInvalidReview = errors.New("srs: resposta invalida")
	ErrTimeTravel    = errors.New("srs: resposta anterior a ultima revisao")
)
