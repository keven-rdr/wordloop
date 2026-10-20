package domain

import (
	"strings"
	"unicode"
)

// minTypoLen: so palavras com 5 letras ou mais toleram um erro de digitacao (docs/fase-1/01-srs.md, secao 4.4).
const minTypoLen = 5

// Verdict e a correcao de uma resposta digitada.
type Verdict struct {
	Correct bool
	Typo    bool // quase: distancia de edicao 1 em palavra longa; vale Hard sem lapso
}

// CorrectTyped corrige o que foi digitado contra o termo esperado, sem diferenciar maiusculas, acentos nem espacos extras.
func CorrectTyped(expected, answer string) Verdict {
	e, a := normalize(expected), normalize(answer)
	switch {
	case a == "":
		return Verdict{}
	case e == a:
		return Verdict{Correct: true}
	case len([]rune(e)) >= minTypoLen && withinOneEdit(e, a):
		return Verdict{Typo: true}
	}
	return Verdict{}
}

// normalize deixa em minusculas, sem acentos e com espacos simples.
func normalize(s string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(strings.TrimSpace(s)) {
		if f, ok := fold[r]; ok {
			r = f
		}
		if unicode.IsSpace(r) {
			r = ' '
		}
		b.WriteRune(r)
	}
	return strings.Join(strings.Fields(b.String()), " ")
}

var fold = func() map[rune]rune {
	m := map[rune]rune{}
	for base, group := range map[rune]string{
		'a': "áàâãäå", 'e': "éèêë", 'i': "íìîï", 'o': "óòôõö", 'u': "úùûü", 'c': "ç", 'n': "ñ",
	} {
		for _, r := range group {
			m[r] = base
		}
	}
	return m
}()

// withinOneEdit diz se a e b diferem por exatamente uma troca, uma insercao, uma remocao ou a inversao de duas
// letras vizinhas (os erros de digitacao mais comuns).
func withinOneEdit(a, b string) bool {
	x, y := []rune(a), []rune(b)
	switch len(x) - len(y) {
	case 0:
		return oneSubstitution(x, y) || oneSwap(x, y)
	case 1:
		return oneDeletion(x, y)
	case -1:
		return oneDeletion(y, x)
	}
	return false
}

// firstDiff e a posicao da primeira diferenca (ou o tamanho do menor, se um e prefixo do outro).
func firstDiff(x, y []rune) int {
	i := 0
	for i < len(x) && i < len(y) && x[i] == y[i] {
		i++
	}
	return i
}

func oneSubstitution(x, y []rune) bool {
	i := firstDiff(x, y)
	return i < len(x) && equal(x[i+1:], y[i+1:])
}

func oneSwap(x, y []rune) bool {
	i := firstDiff(x, y)
	return i+1 < len(x) && x[i] == y[i+1] && x[i+1] == y[i] && equal(x[i+2:], y[i+2:])
}

// oneDeletion: long tem uma letra a mais que short.
func oneDeletion(long, short []rune) bool {
	i := firstDiff(long, short)
	return equal(long[i+1:], short[i:])
}

func equal(a, b []rune) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
