package domain

import "math"

// Pontos de cada resposta (ADR 0032). A mesma tabela recalcula os dias antigos na migracao 00025.
const (
	pointsRecognition = 10   // acerto no reconhecimento (escolher)
	pointsProduction  = 15   // acerto na producao (digitar, lacuna, imagem, montar)
	pointsEffort      = 2    // erro: o esforco tambem conta, pouco
	DailyPointsCap    = 3000 // teto de pontos por dia de estudo
)

// levelFactor e o multiplicador do acerto pelo nivel CEFR do item; sem nivel conta como A1, B2 ou mais como 2.
func levelFactor(cefr string) float64 {
	switch cefr {
	case "", "A1":
		return 1
	case "A2":
		return 1.2
	case "B1":
		return 1.5
	}
	return 2
}

// ReviewPoints sao os pontos de uma resposta: acerto pelo tipo de exercicio e pelo nivel da palavra; erro so o esforco.
func ReviewPoints(correct, production bool, cefr string) int {
	if !correct {
		return pointsEffort
	}
	base := pointsRecognition
	if production {
		base = pointsProduction
	}
	return int(math.Round(float64(base) * levelFactor(cefr)))
}
