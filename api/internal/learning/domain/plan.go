package domain

import (
	"errors"
	"time"
)

const (
	reviewsPerMinute = 4  // ~15 s por revisao [PREMISSA]
	minutesPerNew    = 5  // palavras novas = minutos / 5
	newCap           = 8  // teto inicial (nivel A1) [PREMISSA: Fase 1, secao 10]
	avalancheFactor  = 2  // sem novas se a fila de revisoes passa de 2x o limite
	sessionTTL       = 12 // horas ate o pacote expirar
	// ProductiveFrom e a estabilidade a partir da qual nasce o cartao produtivo (digitar).
	ProductiveFrom = 7.0
)

var (
	ErrInvalidMinutes = errors.New("minutos invalidos")
	ErrInvalidMode    = errors.New("modo invalido")
	ErrInvalidTZ      = errors.New("fuso invalido")
)

// Limits sao os limites da sessao: quantas revisoes e quantas palavras novas cabem.
type Limits struct{ Reviews, New int }

// PlanLimits calcula os limites de "quanto tempo hoje?": 5, 15, 30 ou 60 minutos.
func PlanLimits(minutes int, mode Mode, due int) (Limits, error) {
	switch minutes {
	case 5, 15, 30, 60:
	default:
		return Limits{}, ErrInvalidMinutes
	}
	if mode != ModeReview && mode != ModeLearn && mode != ModeMixed {
		return Limits{}, ErrInvalidMode
	}
	l := Limits{Reviews: minutes * reviewsPerMinute, New: min(minutes/minutesPerNew, newCap)}
	if mode == ModeReview || due > avalancheFactor*l.Reviews {
		l.New = 0
	}
	if mode == ModeLearn {
		l.Reviews = 0
	}
	return l, nil
}

// Expiry e o fim da validade do pacote.
func Expiry(now time.Time) time.Time { return now.Add(sessionTTL * time.Hour) }
