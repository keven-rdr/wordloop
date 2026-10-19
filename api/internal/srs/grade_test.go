package srs

import (
	"errors"
	"testing"
)

func TestGradeMapping(t *testing.T) {
	cases := map[string]struct {
		r    Review
		want Grade
	}{
		"errou":                       {Review{Correct: false, ErrorKind: ErrorNoIdea}, Again},
		"errou quase":                 {Review{Correct: false, ErrorKind: ErrorAlmost}, Again},
		"foi distracao":               {Review{Correct: false, ErrorKind: ErrorSlip}, Hard},
		"erro de digitacao":           {Review{Correct: false, Typo: true}, Hard},
		"acertou dificil":             {Review{Correct: true, Difficulty: DifficultyHard}, Hard},
		"acertou medio":               {Review{Correct: true, Difficulty: DifficultyMedium}, Good},
		"facil em reconhecimento":     {Review{Correct: true, Difficulty: DifficultyEasy}, Good},
		"facil em producao":           {Review{Correct: true, Difficulty: DifficultyEasy, Production: true}, Easy},
		"dica limita a Hard":          {Review{Correct: true, Difficulty: DifficultyEasy, Production: true, Hint: true}, Hard},
		"explicacao limita a Hard":    {Review{Correct: true, Difficulty: DifficultyMedium, Explained: true}, Hard},
		"imagem antes limita a Hard":  {Review{Correct: true, Difficulty: DifficultyMedium, ImageBefore: true}, Hard},
		"dificil com dica continua":   {Review{Correct: true, Difficulty: DifficultyHard, Hint: true}, Hard},
		"media em producao e Good":    {Review{Correct: true, Difficulty: DifficultyMedium, Production: true}, Good},
		"errou ignora a dificuldade":  {Review{Correct: false, Difficulty: DifficultyEasy}, Again},
		"digitacao com dica continua": {Review{Correct: false, Typo: true, Hint: true}, Hard},
	}
	for name, c := range cases {
		got, err := c.r.Grade()
		if err != nil || got != c.want {
			t.Errorf("%s: %v %v, esperava %v", name, got, err, c.want)
		}
	}
}

func TestGradeRejectsMissingDifficultyOnCorrect(t *testing.T) {
	for _, d := range []Difficulty{"", "otimo"} {
		if _, err := (Review{Correct: true, Difficulty: d}).Grade(); !errors.Is(err, ErrInvalidReview) {
			t.Errorf("dificuldade %q: esperava ErrInvalidReview, veio %v", d, err)
		}
	}
}

func TestOnlyAgainIsALapse(t *testing.T) {
	if !Again.Lapse() || Hard.Lapse() || Good.Lapse() || Easy.Lapse() {
		t.Error("so Again e lapso")
	}
}
