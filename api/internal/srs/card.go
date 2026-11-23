package srs

import (
	"errors"
	"time"
)

type State string

const (
	StateNew        State = "new"
	StateLearning   State = "learning"
	StateReview     State = "review"
	StateRelearning State = "relearning"
)

type Difficulty string

const (
	DifficultyHard   Difficulty = "hard"
	DifficultyMedium Difficulty = "medium"
	DifficultyEasy   Difficulty = "easy"
	// DifficultyKnown e o "ja sabia": so no primeiro contato, no reconhecimento e com acerto (Review.GradeOn).
	DifficultyKnown Difficulty = "known"
)

type ErrorKind string

const (
	ErrorNone   ErrorKind = ""
	ErrorNoIdea ErrorKind = "no_idea"
	ErrorAlmost ErrorKind = "almost"
	ErrorSlip   ErrorKind = "slip"
)

type CardState struct {
	ID           string
	State        State
	Stability    float64
	Reps         int
	Lapses       int
	SinceLapse   int
	LastReviewAt time.Time
	DueAt        time.Time
}

type Review struct {
	Correct     bool
	Difficulty  Difficulty
	ErrorKind   ErrorKind
	Typo        bool
	Production  bool
	Hint        bool
	Explained   bool
	ImageBefore bool
}

type Scheduler interface {
	Name() string
	Next(c CardState, r Review, now time.Time) (CardState, error)
	Retrievability(c CardState, now time.Time) float64
}

var (
	ErrInvalidReview = errors.New("srs: resposta invalida")
	ErrTimeTravel    = errors.New("srs: resposta anterior a ultima revisao")
)

// IsNew diz se o cartao ainda nao teve resposta (o primeiro contato com a palavra).
func (c CardState) IsNew() bool { return c.State == StateNew || c.State == "" }
