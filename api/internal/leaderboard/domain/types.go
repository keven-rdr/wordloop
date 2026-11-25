// Package domain tem as regras do ranking (ADR 0032): periodo, nivel e apelido. Sem I/O.
package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	// ErrInvalidPeriod e o periodo desconhecido (so semana e geral).
	ErrInvalidPeriod = errors.New("periodo invalido")
	// ErrInvalidNickname e o apelido fora do formato ou com palavra bloqueada.
	ErrInvalidNickname = errors.New("apelido invalido")
	// ErrNicknameTaken e o apelido que outra pessoa ja usa.
	ErrNicknameTaken = errors.New("apelido em uso")
)

// Viewer e quem pergunta.
type Viewer struct{ UserID uuid.UUID }

// Period e o recorte do ranking.
type Period string

const (
	Week Period = "week" // a semana da meta: de segunda a domingo
	All  Period = "all"  // desde sempre
)

// Range sao os dias (inclusive) que o periodo cobre a partir do dia de estudo de hoje.
func (p Period) Range(today time.Time) (time.Time, time.Time, error) {
	switch p {
	case Week:
		start := WeekStart(today)
		return start, start.AddDate(0, 0, 6), nil
	case All:
		return time.Date(1, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(9999, 12, 31, 0, 0, 0, 0, time.UTC), nil
	}
	return time.Time{}, time.Time{}, ErrInvalidPeriod
}

// WeekStart e a segunda-feira da semana do dia dado (datas a meia-noite UTC, como o study_day).
func WeekStart(day time.Time) time.Time {
	back := (int(day.Weekday()) + 6) % 7
	return day.AddDate(0, 0, -back)
}

// Today e o dia de estudo de um instante no fuso e corte da pessoa (a mesma regra do learning, que grava o dia).
func Today(now time.Time, tz string, cutoffHour int) time.Time {
	loc, err := time.LoadLocation(tz)
	if err != nil || tz == "" {
		loc = time.UTC
	}
	local := now.In(loc).Add(-time.Duration(cutoffHour) * time.Hour)
	return time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, time.UTC)
}

// Entry e uma linha do ranking. Nickname vazio = quem pergunta e ainda nao participa.
type Entry struct {
	UserID   uuid.UUID
	Position int
	Nickname string
	Points   int
}

// Board e o ranking do periodo: as primeiras posicoes e a linha de quem pergunta (nil se ainda nao pontuou).
type Board struct {
	Period   Period
	From, To time.Time
	Entries  []Entry
	Me       *Entry
}

// Profile e como a pessoa aparece no ranking.
type Profile struct {
	Nickname string
	Visible  bool
}

// Score e o resumo da pessoa: pontos da semana e de sempre, nivel e dias da meta cumpridos na semana.
type Score struct {
	WeekPoints, TotalPoints, WeekDaysMet int
	Level                                Level
}
