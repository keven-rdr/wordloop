package domain

import (
	"testing"

	"github.com/google/uuid"
)

func ids(n int) []uuid.UUID {
	out := make([]uuid.UUID, n)
	for i := range out {
		out[i] = uuid.New()
	}
	return out
}

func TestStretchCountIsOneInFive(t *testing.T) {
	for n, want := range map[int]int{0: 0, 4: 0, 5: 1, 9: 1, 16: 3, 45: 9} {
		if got := StretchCount(n); got != want {
			t.Errorf("StretchCount(%d) = %d, esperava %d", n, got, want)
		}
	}
}

func TestPickNewLeavesRoomForStretchAndFillsWithTheLesson(t *testing.T) {
	main, stretch := ids(20), ids(5)
	got := PickNew(main, stretch, 10) // 2 de desafio
	if len(got) != 10 || got[7] != main[7] || got[8] != stretch[0] || got[9] != stretch[1] {
		t.Fatalf("8 da licao e 2 de desafio: %v", got)
	}
	// sem desafio, a licao completa
	if got := PickNew(main, nil, 10); len(got) != 10 || got[9] != main[9] {
		t.Errorf("sem desafio, so a licao: %v", got)
	}
	// licao curta: o desafio completa
	if got := PickNew(main[:3], stretch, 10); len(got) != 8 {
		t.Errorf("3 da licao + 5 de desafio: %v", got)
	}
	// repetido entre as listas entra uma vez so
	if got := PickNew(main[:5], append([]uuid.UUID{main[0]}, stretch[:1]...), 5); len(got) != 5 || got[4] != stretch[0] {
		t.Errorf("sem repetidos: %v", got)
	}
}

func cards(n int) []Card {
	out := make([]Card, n)
	for i := range out {
		out[i] = Card{ID: uuid.New()}
	}
	return out
}

func sameCards(x, y []Card) bool {
	if len(x) != len(y) {
		return false
	}
	for i := range x {
		if x[i].ID != y[i].ID {
			return false
		}
	}
	return true
}

func TestMixKeepsReviewsInOrderAndShufflesNewOnesBySeed(t *testing.T) {
	reviews, fresh := cards(10), cards(8)
	a, b := uuid.New(), uuid.New()
	mixed := Mix(reviews, fresh, a)
	if len(mixed) != 18 || !sameCards(mixed, Mix(reviews, fresh, a)) {
		t.Fatal("a mesma semente da o mesmo pacote, sem perder cartoes")
	}
	if sameCards(mixed, Mix(reviews, fresh, b)) {
		t.Error("outra semente, outro pacote")
	}
	isReview := map[uuid.UUID]bool{}
	for _, c := range reviews {
		isReview[c.ID] = true
	}
	var kept []Card
	seen := map[uuid.UUID]bool{}
	for _, c := range mixed {
		if isReview[c.ID] {
			kept = append(kept, c)
		}
		seen[c.ID] = true
	}
	if !sameCards(kept, reviews) || len(seen) != 18 {
		t.Error("as revisoes mantem a ordem (menor R primeiro) e nada se repete")
	}
	if got := Mix(nil, fresh, a); len(got) != 8 {
		t.Errorf("so novas: %d", len(got))
	}
	if got := Mix(reviews, nil, a); !sameCards(got, reviews) {
		t.Error("so revisoes: a mesma ordem")
	}
}
