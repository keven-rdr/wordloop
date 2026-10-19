package srs

import (
	"hash/fnv"
	"math"
	"strconv"
)

const fuzzFrom = 2.5 // abaixo disso nao ha fuzz

// fuzzPct e a variacao maxima do intervalo, por faixa (como no Anki [VERIFICAR no codigo rslib]).
func fuzzPct(days float64) float64 {
	switch {
	case days < fuzzFrom:
		return 0
	case days < 7:
		return 0.15
	case days < 20:
		return 0.10
	default:
		return 0.05
	}
}

// unit sorteia de forma determinista um valor em [-1, 1] a partir do cartao e do numero da revisao.
func unit(id string, reps int) float64 {
	h := fnv.New32a()
	_, _ = h.Write([]byte(id + ":" + strconv.Itoa(reps)))
	return float64(h.Sum32())/float64(math.MaxUint32)*2 - 1
}

// fuzzed espalha vencimentos para evitar pilhas no mesmo dia; nunca devolve menos de 1 dia.
func fuzzed(days float64, id string, reps int) float64 {
	pct := fuzzPct(days)
	if pct == 0 {
		return days
	}
	return math.Max(1, math.Floor(days*(1+pct*unit(id, reps))+0.5))
}
