package domain

import (
	"cmp"
	"hash/fnv"
	"slices"

	"github.com/google/uuid"
)

// stretchShare e a fatia das palavras novas que vem de licoes adiante (o desafio): 1 em cada 5.
const stretchShare = 5

// StretchCount e quantas das n palavras novas da sessao sao desafio (de licoes adiante).
func StretchCount(n int) int { return n / stretchShare }

// SeedKey e uma chave de ordenacao que depende da semente da sessao: a mesma sessao da sempre a mesma ordem, e outra
// sessao da outra ordem.
func SeedKey(seed uuid.UUID, parts ...uuid.UUID) uint32 {
	h := fnv.New32a()
	_, _ = h.Write(seed[:])
	for _, p := range parts {
		_, _ = h.Write(p[:])
	}
	return h.Sum32()
}

// Mix intercala as palavras novas entre as revisoes, em lugares sorteados pela semente: as revisoes mantem a ordem
// (menor retencao primeiro, quem para no meio revisou as mais urgentes) e as novas mudam de lugar e de ordem a cada
// sessao. A mesma semente da sempre o mesmo pacote.
func Mix(reviews, fresh []Card, seed uuid.UUID) []Card {
	news := slices.Clone(fresh)
	slices.SortStableFunc(news, func(a, b Card) int { return cmp.Compare(SeedKey(seed, a.ID), SeedKey(seed, b.ID)) })
	slots := make([]uint32, len(news)) // posicao relativa (0 a 2^32) de cada nova no meio das revisoes
	for i, c := range news {
		slots[i] = SeedKey(seed, c.ID, seed)
	}
	slices.Sort(slots)
	out := make([]Card, 0, len(reviews)+len(news))
	r := 0
	for i, c := range news {
		at := int(uint64(slots[i]) * uint64(len(reviews)+1) >> 32) // quantas revisoes vem antes desta nova
		for ; r < at && r < len(reviews); r++ {
			out = append(out, reviews[r])
		}
		out = append(out, c)
	}
	return append(out, reviews[r:]...)
}

// PickNew junta as palavras novas da sessao: primeiro as da licao atual (main, na ordem das licoes), deixando lugar
// para as de desafio (stretch); se faltar desafio, a licao completa. Sem repetidos e no maximo n.
func PickNew(main, stretch []uuid.UUID, n int) []uuid.UUID {
	seen := make(map[uuid.UUID]bool, n)
	out := make([]uuid.UUID, 0, n)
	add := func(ids []uuid.UUID, upTo int) {
		for _, id := range ids {
			if len(out) >= upTo {
				return
			}
			if !seen[id] {
				seen[id] = true
				out = append(out, id)
			}
		}
	}
	add(main, n-min(len(stretch), StretchCount(n)))
	add(stretch, n)
	add(main, n)
	return out
}
