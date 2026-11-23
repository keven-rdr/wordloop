package domain

import (
	"errors"
	"time"
)

const (
	reviewsPerMinute = 4 // ~15 s por revisao [PREMISSA]
	newBase          = 6 // palavras novas = 6 + 0,65 x minutos (5 min: 9; 15: 16; 30: 26; 60: 45)
	newPerMinute     = 0.65
	newCap           = 45 // teto de palavras novas numa sessao
	avalancheFactor  = 2  // com a fila de revisoes acima de 2x o limite, as novas caem pela metade
	// MinMinutes e MaxMinutes sao os limites do tempo da sessao (o tempo vem da meta diaria).
	MinMinutes = 5
	MaxMinutes = 240
	sessionTTL = 12 // horas ate o pacote expirar
	// ProductiveFrom e a estabilidade a partir da qual nasce o cartao produtivo (digitar).
	ProductiveFrom = 7.0
	// ProductiveFromSentence vale para itens do tipo frase: montar a frase e o objetivo de estudar uma frase, entao o
	// cartao produtivo nasce mais cedo (reconhecer a frase pronta e facil; sem isso ninguem montaria nas primeiras semanas).
	ProductiveFromSentence = 3.0
)

var (
	ErrInvalidMinutes = errors.New("minutos invalidos")
	ErrInvalidMode    = errors.New("modo invalido")
	ErrInvalidTZ      = errors.New("fuso invalido")
)

// Limits sao os limites da sessao: quantas revisoes, quantas palavras novas e quantos cartoes produtivos (digitar,
// montar) cabem. Os produtivos tem cota propria: nao tiram a vaga de palavras novas.
type Limits struct{ Reviews, New, Productive int }

// PlanLimits calcula os limites da sessao para o tempo do dia (de MinMinutes a MaxMinutes).
func PlanLimits(minutes int, mode Mode, due int) (Limits, error) {
	if minutes < MinMinutes || minutes > MaxMinutes {
		return Limits{}, ErrInvalidMinutes
	}
	if mode != ModeReview && mode != ModeLearn && mode != ModeMixed {
		return Limits{}, ErrInvalidMode
	}
	l := Limits{Reviews: minutes * reviewsPerMinute, New: NewWordsFor(minutes)}
	if due > avalancheFactor*l.Reviews {
		l.New /= 2
	}
	if mode == ModeReview {
		l.New = 0
	}
	if mode == ModeLearn {
		l.Reviews = 0
	}
	l.Productive = (l.New + 1) / 2
	return l, nil
}

// NewWordsFor e quantas palavras novas cabem em minutes minutos: 6 + 0,65 por minuto, ate newCap.
func NewWordsFor(minutes int) int {
	return min(newCap, int(newBase+newPerMinute*float64(minutes)+0.5))
}

// Expiry e o fim da validade do pacote.
func Expiry(now time.Time) time.Time { return now.Add(sessionTTL * time.Hour) }
