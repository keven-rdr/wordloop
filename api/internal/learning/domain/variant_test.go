package domain

import (
	"testing"

	"github.com/google/uuid"

	"github.com/keven-rdr/wordloop/api/internal/srs"
)

func TestBlank(t *testing.T) {
	cases := map[string]struct {
		example, term, want string
		ok                  bool
	}{
		"troca a palavra":          {"This is a big house.", "house", "This is a big ____.", true},
		"ignora maiusculas":        {"House prices are up.", "house", "____ prices are up.", true},
		"so a primeira ocorrencia": {"A house is not a home, a house is a house.", "house", "A ____ is not a home, a house is a house.", true},
		"expressao":                {"Please give up now.", "give up", "Please ____ now.", true},
		"palavra dentro de outra":  {"The housekeeper left.", "house", "", false},
		"flexionada":               {"She runs fast.", "run", "", false},
		"sem exemplo":              {"", "house", "", false},
		"sem termo":                {"This is a house.", "  ", "", false},
		"termo com pontuacao":      {"Use a.m. for morning.", "a.m.", "Use ____ for morning.", true},
	}
	for name, c := range cases {
		got, ok := Blank(c.example, c.term)
		if ok != c.ok || got != c.want {
			t.Errorf("%s: Blank(%q, %q) = %q, %v", name, c.example, c.term, got, ok)
		}
	}
}

func productive(stability float64, reps int) Card {
	id := uuid.New()
	return Card{ID: id, Direction: Productive, State: srs.CardState{State: srs.StateReview, Stability: stability, Reps: reps}}
}

func TestPickExerciseReceptiveAlwaysChooses(t *testing.T) {
	c := Card{ID: uuid.New(), Direction: Receptive}
	full := Content{Term: "house", Example: "This is a house.", ImageKey: "a.png", Theme: FunctionalTheme}
	if got, blank := PickExercise(c, full); got != Choose || blank != "" {
		t.Errorf("receptivo escolhe: %v %q", got, blank)
	}
}

func TestPickExerciseWithoutExampleOrImageIsTheDefaultTyping(t *testing.T) {
	for reps := 0; reps < 12; reps++ {
		if got, _ := PickExercise(productive(10, reps), Content{Term: "house", Gloss: "casa"}); got != Type {
			t.Fatalf("sem exemplo nem imagem: %v (reps %d)", got, reps)
		}
	}
}

func TestPickExerciseFunctionalWordsWithAnExampleAlwaysGetTheCloze(t *testing.T) {
	content := Content{Term: "the", Example: "The cat sleeps.", ImageKey: "a.png", Theme: FunctionalTheme}
	for reps := 0; reps < 12; reps++ {
		got, blank := PickExercise(productive(5, reps), content)
		if got != Cloze || blank != "____ cat sleeps." {
			t.Fatalf("palavra funcional: %v %q (reps %d)", got, blank, reps)
		}
	}
}

func TestPickExerciseVariesWithTheRepetitionAndNeverUsesAMissingVariant(t *testing.T) {
	both := Content{Term: "house", Example: "This is a big house.", ImageKey: "a.png"}
	seen := map[Exercise]bool{}
	for reps := 0; reps < 40; reps++ {
		got, _ := PickExercise(productive(10, reps), both)
		seen[got] = true
	}
	if !seen[Cloze] || !seen[Image] || !seen[Type] {
		t.Errorf("com lacuna e imagem possiveis, as tres variantes aparecem: %v", seen)
	}

	onlyImage := Content{Term: "house", ImageKey: "a.png"}
	for reps := 0; reps < 40; reps++ {
		if got, blank := PickExercise(productive(10, reps), onlyImage); got == Cloze || blank != "" {
			t.Fatalf("sem exemplo nao ha lacuna: %v %q", got, blank)
		}
	}
	onlyCloze := Content{Term: "house", Example: "This is a big house."}
	for reps := 0; reps < 40; reps++ {
		if got, _ := PickExercise(productive(10, reps), onlyCloze); got == Image {
			t.Fatalf("sem imagem nao ha imagem->palavra (reps %d)", reps)
		}
	}
}

func TestPickExerciseImageFadesWhenTheCardIsStable(t *testing.T) {
	onlyImage := Content{Term: "house", ImageKey: "a.png"}
	for reps := 0; reps < 40; reps++ {
		if got, _ := PickExercise(productive(ImageFadesAt, reps), onlyImage); got != Type {
			t.Fatalf("a partir de S = 21 d a imagem some: %v (reps %d)", got, reps)
		}
	}
}

func TestPickExerciseIsDeterministicPerCardAndRepetition(t *testing.T) {
	c := productive(10, 4)
	content := Content{Term: "house", Example: "This is a big house.", ImageKey: "a.png"}
	first, firstBlank := PickExercise(c, content)
	for range 5 {
		if got, blank := PickExercise(c, content); got != first || blank != firstBlank {
			t.Fatal("o mesmo cartao na mesma revisao tem sempre a mesma pergunta")
		}
	}
}

func TestExerciseKinds(t *testing.T) {
	for _, e := range []Exercise{Type, Cloze, Image} {
		if !e.IsTyped() {
			t.Errorf("%s e digitado", e)
		}
		if !(Card{Direction: Productive}).Accepts(e) || (Card{Direction: Receptive}).Accepts(e) {
			t.Errorf("%s serve so ao produtivo", e)
		}
	}
	if Choose.IsTyped() || !(Card{Direction: Receptive}).Accepts(Choose) || (Card{Direction: Productive}).Accepts(Choose) {
		t.Error("escolher serve so ao receptivo")
	}
}
