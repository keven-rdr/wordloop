package srs

// Grade e a nota (1 a 4) que alimenta o agendador.
type Grade int

const (
	Again Grade = 1
	Hard  Grade = 2
	Good  Grade = 3
	Easy  Grade = 4
)

// Lapse diz se a nota e um lapso (esquecer um cartao que ja estava em revisao).
func (g Grade) Lapse() bool { return g == Again }

// Grade traduz a resposta em nota (docs/fase-1/01-srs.md, secao 4.4):
//   - errou: Again; mas "foi distracao" ou erro de digitacao vale Hard sem lapso;
//   - acertou: a dificuldade dificil/media/facil vira Hard/Good/Easy;
//   - Easy so em exercicio de producao (em reconhecimento conta como Good);
//   - dica, explicacao ou imagem vistas antes de responder limitam a nota a Hard.
func (r Review) Grade() (Grade, error) {
	if !r.Correct {
		return r.wrongGrade(), nil
	}
	g, err := r.rightGrade()
	if err != nil {
		return 0, err
	}
	if r.Hint || r.Explained || r.ImageBefore {
		g = min(g, Hard)
	}
	return g, nil
}

func (r Review) wrongGrade() Grade {
	if r.ErrorKind == ErrorSlip || r.Typo {
		return Hard
	}
	return Again
}

func (r Review) rightGrade() (Grade, error) {
	switch r.Difficulty {
	case DifficultyHard:
		return Hard, nil
	case DifficultyMedium:
		return Good, nil
	case DifficultyEasy:
		if r.Production {
			return Easy, nil
		}
		return Good, nil
	}
	return 0, ErrInvalidReview
}
