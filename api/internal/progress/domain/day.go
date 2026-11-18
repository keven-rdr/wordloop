package domain

import (
	"errors"
	"time"
)

// ErrInvalidTZ e o fuso desconhecido.
var ErrInvalidTZ = errors.New("fuso invalido")

// Today e o "dia de estudo" de um instante: o dia local descontado o corte (04:00 por padrao), para o estudo
// da madrugada contar no dia anterior. Devolve meia-noite UTC do dia logico, como o study_day grava.
// Mesma regra do modulo learning, que grava o dia; os modulos nao se importam, por isso a copia.
func Today(now time.Time, tz string, cutoffHour int) (time.Time, error) {
	loc, err := time.LoadLocation(tz)
	if err != nil || tz == "" {
		return time.Time{}, ErrInvalidTZ
	}
	local := now.In(loc).Add(-time.Duration(cutoffHour) * time.Hour)
	return time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, time.UTC), nil
}

// AddDays soma dias a uma data (meia-noite UTC).
func AddDays(date time.Time, n int) time.Time { return date.AddDate(0, 0, n) }

// Streak conta os dias seguidos com estudo que terminam hoje. Se hoje ainda nao estudou, a sequencia continua
// valendo ate ontem (a pessoa ainda pode estudar hoje); so quebra quando passa um dia inteiro sem estudo.
func Streak(days []Day, today time.Time) int {
	studied := map[time.Time]bool{}
	for _, d := range days {
		if d.Studied() {
			studied[d.Date] = true
		}
	}
	cursor := today
	if !studied[cursor] {
		cursor = AddDays(cursor, -1)
	}
	n := 0
	for studied[cursor] {
		n++
		cursor = AddDays(cursor, -1)
	}
	return n
}

// Accuracy e a taxa de acerto dos dias dados; nil se nao houve nenhuma resposta.
func Accuracy(days []Day) *float64 {
	reviews, correct := 0, 0
	for _, d := range days {
		reviews += d.Reviews
		correct += d.Correct
	}
	if reviews == 0 {
		return nil
	}
	rate := float64(correct) / float64(reviews)
	return &rate
}

// Window devolve so os dias entre from e to (inclusive).
func Window(days []Day, from, to time.Time) []Day {
	var out []Day
	for _, d := range days {
		if !d.Date.Before(from) && !d.Date.After(to) {
			out = append(out, d)
		}
	}
	return out
}

// Forecast monta os n dias a partir de hoje a partir da contagem por posicao (0 = hoje, com as atrasadas).
func Forecast(today time.Time, due map[int]int, n int) []ForecastDay {
	out := make([]ForecastDay, n)
	for i := range out {
		out[i] = ForecastDay{Date: AddDays(today, i), Due: due[i]}
	}
	return out
}
