package domain

import (
	"hash/fnv"
	"regexp"
	"strconv"
	"strings"
)

const (
	// BlankMark e o que ocupa o lugar da palavra na frase da lacuna.
	BlankMark = "____"
	// FunctionalTheme e a unidade das palavras funcionais (the, of, would): sempre pedem contexto (lacuna).
	FunctionalTheme = "little-words"
	// ImageFadesAt e a estabilidade (dias) a partir da qual a imagem some do exercicio (fading, Fase 1 secao 9).
	ImageFadesAt = 21.0
	variantSlots = 3
)

// Blank troca a palavra pela lacuna na frase de exemplo. O segundo retorno e false se a frase nao tem a palavra
// (por exemplo, flexionada), e a lacuna nao serve.
func Blank(example, term string) (string, bool) {
	if strings.TrimSpace(example) == "" || strings.TrimSpace(term) == "" {
		return "", false
	}
	// A palavra inteira: sem letra ou numero colado antes ou depois (\b nao serve a termos como "a.m.").
	re := regexp.MustCompile(`(?i)(^|[^\p{L}\p{N}])(` + regexp.QuoteMeta(term) + `)([^\p{L}\p{N}]|$)`)
	at := re.FindStringSubmatchIndex(example)
	if at == nil {
		return "", false
	}
	return example[:at[4]] + BlankMark + example[at[5]:], true
}

// variantSlot reparte os cartoes em fatias fixas (0 a 2): o mesmo cartao na mesma revisao cai sempre na mesma
// fatia, e a cada revisao muda (varia o exercicio sem sorteio).
func variantSlot(c Card) int {
	h := fnv.New32a()
	_, _ = h.Write(c.ID[:])
	_, _ = h.Write([]byte(strconv.Itoa(c.State.Reps)))
	return int(h.Sum32() % variantSlots)
}

// SentenceKind e o tipo de item que se monta com pecas.
const SentenceKind = "sentence"

// PickExercise escolhe a pergunta do cartao nesta vez. Receptivo sempre escolhe. Produtivo: frases com 3 a 9 pecas
// viram montar frase (Assemble); palavras funcionais com exemplo viram lacuna; nos demais, a 1a fatia e lacuna (se couber, senao imagem), a 2a e imagem (se houver e a
// estabilidade for menor que ImageFadesAt) e a 3a e digitar a partir da traducao. O segundo retorno e a frase da lacuna.
func PickExercise(c Card, content Content) (Exercise, string) {
	if c.Exercise() == Choose {
		return Choose, ""
	}
	if content.Kind == SentenceKind && AssemblePossible(content.Term) {
		return Assemble, ""
	}
	blank, canCloze := Blank(content.Example, content.Term)
	canImage := content.ImageKey != "" && c.State.Stability < ImageFadesAt
	slot := variantSlot(c)
	switch {
	case canCloze && (content.Theme == FunctionalTheme || slot == 0):
		return Cloze, blank
	case canImage && slot <= 1:
		return Image, ""
	}
	return Type, ""
}
