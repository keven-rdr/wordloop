package domain_test

import (
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/keven-rdr/wordloop/api/internal/learning/domain"
	"github.com/keven-rdr/wordloop/api/internal/srs"
)

func cardFor(id string) domain.Card {
	u := uuid.MustParse(id)
	return domain.Card{ID: u, Direction: domain.Productive, State: srs.CardState{Reps: 2}}
}

func TestAssemblePossibleNeedsThreeToNineWords(t *testing.T) {
	cases := map[string]bool{
		"Go.": false, "I go.": false, "I like it.": true, "Tom likes to read books every day.": true,
		"One two three four five six seven eight nine.": true, "One two three four five six seven eight nine ten.": false, "": false,
		"Well, I like it!": true,
	}
	for term, want := range cases {
		if got := domain.AssemblePossible(term); got != want {
			t.Errorf("%q: %v, esperava %v", term, got, want)
		}
	}
}

func TestTilesAreTheWordsWithoutPunctuationNeverInTheRightOrderAndStablePerCard(t *testing.T) {
	term := "I don't want to go there, Tom."
	right := []string{"I", "don't", "want", "to", "go", "there", "Tom"}
	for i := range 40 { // varios cartoes: nenhum pode mostrar a ordem certa
		c := cardFor(uuid.NewSHA1(uuid.NameSpaceURL, []byte{byte(i)}).String())
		tiles := domain.Tiles(c, term)
		if reflect.DeepEqual(tiles, right) {
			t.Fatalf("cartao %d: as pecas vieram na ordem certa", i)
		}
		if got, want := slices.Sorted(slices.Values(tiles)), slices.Sorted(slices.Values(right)); !slices.Equal(got, want) {
			t.Fatalf("as pecas sao as palavras da frase, sem pontuacao: %v", tiles)
		}
		if !slices.Equal(tiles, domain.Tiles(c, term)) {
			t.Fatal("o mesmo cartao mostra sempre a mesma ordem")
		}
	}
	a, b := domain.Tiles(cardFor("00000000-0000-0000-0000-000000000001"), term), domain.Tiles(cardFor("00000000-0000-0000-0000-000000000002"), term)
	if slices.Equal(a, b) {
		t.Log("dois cartoes com a mesma ordem e possivel, mas improvavel; nao e erro")
	}
}

func TestTilesLowercaseTheFirstWordExceptIAndNames(t *testing.T) {
	cases := map[string]string{
		"She likes tea.": "she", "I like tea.": "I", "I'm tired.": "I'm", "Tom likes tea.": "Tom", "Mary's here.": "Mary's", "Where is it?": "where",
	}
	for term, first := range cases {
		tiles := domain.Tiles(cardFor("00000000-0000-0000-0000-000000000009"), term)
		if !slices.Contains(tiles, first) {
			t.Errorf("%q: esperava a peca %q em %v", term, first, tiles)
		}
		if strings.HasSuffix(strings.Join(tiles, " "), ".") || slices.Contains(tiles, "") {
			t.Errorf("%q: pontuacao ou peca vazia em %v", term, tiles)
		}
	}
}

func TestTilesOfIdenticalWordsStillTerminateAndEmptyTermHasNone(t *testing.T) {
	if got := domain.Tiles(cardFor("00000000-0000-0000-0000-000000000003"), "No, no, no."); len(got) != 3 {
		t.Fatalf("pecas iguais: %v", got)
	}
	if got := domain.Tiles(cardFor("00000000-0000-0000-0000-000000000003"), " ... "); got != nil {
		t.Fatalf("sem palavras, sem pecas: %v", got)
	}
}

func TestCorrectAssembledIsExactOnWordsAndOrderButIgnoresCaseAndPunctuation(t *testing.T) {
	const term = "I don't want to go there, Tom."
	right := []string{
		"I don't want to go there Tom", "i DON'T want to go there tom", "I don't want to go there, Tom.", "  I  don't want to go there Tom  ",
	}
	for _, a := range right {
		if v := domain.CorrectAssembled(term, a); !v.Correct || v.Typo {
			t.Errorf("%q deveria estar certo: %+v", a, v)
		}
	}
	wrong := []string{
		"", "I don't want to go Tom there", "I dont want to go there Tom", /* sem tolerancia a erro de digitacao */
		"I don't want to go there", "I don't want to go there Tom Tom", "don't I want to go there Tom",
	}
	for _, a := range wrong {
		if v := domain.CorrectAssembled(term, a); v.Correct || v.Typo {
			t.Errorf("%q deveria estar errado, e nunca 'quase': %+v", a, v)
		}
	}
}

func TestPickExercisePutsSentencesInAssembleAndKeepsTheRestAsBefore(t *testing.T) {
	c := cardFor("00000000-0000-0000-0000-000000000004")
	sentence := domain.Content{Term: "I like red houses.", Gloss: "Eu gosto de casas vermelhas.", Kind: "sentence"}
	if ex, cloze := domain.PickExercise(c, sentence); ex != domain.Assemble || cloze != "" {
		t.Fatalf("frase no cartao produtivo: %v %q", ex, cloze)
	}
	short := domain.Content{Term: "Hello there.", Kind: "sentence"}
	if ex, _ := domain.PickExercise(c, short); ex == domain.Assemble {
		t.Fatal("frase com 2 palavras nao se monta: cai nas outras variantes")
	}
	word := domain.Content{Term: "house", Gloss: "casa", Kind: "word", Example: "I like red houses."}
	if ex, _ := domain.PickExercise(c, word); ex == domain.Assemble {
		t.Fatal("palavra nunca vira montar frase")
	}
	receptive := domain.Card{ID: c.ID, Direction: domain.Receptive}
	if ex, _ := domain.PickExercise(receptive, sentence); ex != domain.Choose {
		t.Fatalf("cartao receptivo sempre escolhe: %v", ex)
	}
}

func TestAssembleIsProductionAndOnlyFitsTheProductiveCard(t *testing.T) {
	if !domain.Assemble.IsTyped() {
		t.Fatal("montar frase e producao (libera o Facil e vale no replay)")
	}
	if !(domain.Card{Direction: domain.Productive}).Accepts(domain.Assemble) || (domain.Card{Direction: domain.Receptive}).Accepts(domain.Assemble) {
		t.Fatal("so o cartao produtivo aceita montar frase")
	}
}
