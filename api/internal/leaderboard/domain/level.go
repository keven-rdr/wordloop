package domain

import (
	"regexp"
	"strings"
)

const (
	firstLevelPoints = 100 // pontos para sair do nivel 1
	levelGrowth      = 50  // cada nivel pede 50 pontos a mais que o anterior
)

// Level e o nivel pelo total de pontos: o numero, os pontos em que ele comecou e os do proximo nivel.
type Level struct{ Number, Floor, Next int }

// LevelOf sobe de nivel enquanto o total alcanca o proximo alvo: do nivel n para o n+1 pede 100 + n x 50 pontos
// (1 -> 2 aos 100; 2 -> 3 aos 250; 3 -> 4 aos 450).
func LevelOf(total int) Level {
	l := Level{Number: 1, Floor: 0, Next: firstLevelPoints}
	for total >= l.Next {
		l.Floor = l.Next
		l.Next += firstLevelPoints + l.Number*levelGrowth
		l.Number++
	}
	return l
}

var nicknameFormat = regexp.MustCompile(`^[A-Za-z0-9_]{3,20}$`)

// blocked sao trechos que nao podem aparecer num apelido publico (lista curta, em minusculas).
var blocked = []string{"admin", "wordloop", "porra", "caralho", "merda", "puta", "fuck", "shit", "bitch", "nazi"}

// CheckNickname valida o apelido: 3 a 20 letras sem acento, numeros ou _, sem trecho bloqueado.
func CheckNickname(nickname string) error {
	if !nicknameFormat.MatchString(nickname) {
		return ErrInvalidNickname
	}
	lower := strings.ToLower(nickname)
	for _, word := range blocked {
		if strings.Contains(lower, word) {
			return ErrInvalidNickname
		}
	}
	return nil
}
