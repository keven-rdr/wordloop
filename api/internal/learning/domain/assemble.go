package domain

import (
	"hash/fnv"
	"sort"
	"strconv"
	"strings"
)

const (
	// MinTiles e MaxTiles limitam a frase que se monta com pecas: menos de 3 nao ha o que ordenar; mais de 9 nao cabe no celular.
	MinTiles = 3
	MaxTiles = 9
)

// trimPunct tira a pontuacao das pontas de uma palavra (as pecas nao mostram virgula nem ponto final).
const trimPunct = ".,?!;:\"()"

// properNames sao as palavras que mantem a maiuscula mesmo no comeco da frase (o resto vai para minuscula, para a
// peca nao entregar qual palavra abre a frase).
var properNames = map[string]bool{"Tom": true, "Mary": true}

// answerWords separa a frase em palavras sem pontuacao, na ordem original.
func answerWords(s string) []string {
	var out []string
	for _, w := range strings.Fields(s) {
		if w = strings.Trim(w, trimPunct); w != "" {
			out = append(out, w)
		}
	}
	return out
}

// AssemblePossible diz se a frase serve para montar com pecas (de MinTiles a MaxTiles palavras).
func AssemblePossible(term string) bool {
	n := len(answerWords(term))
	return n >= MinTiles && n <= MaxTiles
}

// lowerFirst tira a maiuscula da primeira palavra, menos "I" (e contracoes dele) e nomes proprios conhecidos.
func lowerFirst(w string) string {
	if w == "I" || strings.HasPrefix(w, "I'") || properNames[w] || strings.HasPrefix(w, "Tom'") || strings.HasPrefix(w, "Mary'") {
		return w
	}
	return strings.ToLower(w[:1]) + w[1:]
}

// Tiles devolve as pecas da frase: as palavras, embaralhadas de forma estavel por cartao (o mesmo cartao mostra sempre
// a mesma ordem) e nunca na ordem certa, para a primeira vista ja ser um exercicio.
func Tiles(c Card, term string) []string {
	words := answerWords(term)
	if len(words) == 0 {
		return nil
	}
	words[0] = lowerFirst(words[0])
	tiles := append([]string(nil), words...)
	shuffle(tiles, c, 0)
	for salt := 1; sameOrder(tiles, words) && salt < 8; salt++ { // reembaralha ate sair da ordem certa
		shuffle(tiles, c, salt)
	}
	if sameOrder(tiles, words) && len(tiles) > 1 { // todas iguais entre si ("no no no"): nao ha ordem a trocar
		tiles[0], tiles[len(tiles)-1] = tiles[len(tiles)-1], tiles[0]
	}
	return tiles
}

func shuffle(tiles []string, c Card, salt int) {
	key := func(i int) uint32 {
		h := fnv.New32a()
		_, _ = h.Write(c.ID[:])
		_, _ = h.Write([]byte(strconv.Itoa(salt) + ":" + strconv.Itoa(i) + ":"))
		_, _ = h.Write([]byte(tiles[i]))
		return h.Sum32()
	}
	keys := make([]uint32, len(tiles))
	for i := range tiles {
		keys[i] = key(i)
	}
	order := make([]int, len(tiles))
	for i := range order {
		order[i] = i
	}
	sort.SliceStable(order, func(a, b int) bool { return keys[order[a]] < keys[order[b]] })
	out := make([]string, len(tiles))
	for i, from := range order {
		out[i] = tiles[from]
	}
	copy(tiles, out)
}

func sameOrder(a, b []string) bool {
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

// CorrectAssembled corrige a frase montada: as mesmas palavras na mesma ordem, sem diferenciar maiusculas, acentos
// nem pontuacao. Nunca ha "quase": uma peca fora do lugar e erro (diferente de CorrectTyped, que perdoa um erro de
// digitacao em palavra longa, o que numa frase inteira perdoaria uma palavra errada).
func CorrectAssembled(expected, answer string) Verdict {
	e, a := answerWords(expected), answerWords(answer)
	if len(a) == 0 || len(e) != len(a) {
		return Verdict{}
	}
	for i := range e {
		if normalize(e[i]) != normalize(a[i]) {
			return Verdict{}
		}
	}
	return Verdict{Correct: true}
}
