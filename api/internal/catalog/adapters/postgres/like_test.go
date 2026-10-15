package postgres

import (
	"testing"

	"github.com/keven-rdr/wordloop/api/internal/catalog/domain"
)

func TestLikeEscapesTheWildcardsTheUserTyped(t *testing.T) {
	for in, want := range map[string]string{
		"":           "",
		"casa":       "%casa%",
		"100%":       `%100\%%`,
		"a_b":        `%a\_b%`,
		`a\b`:        `%a\\b%`,
		"%_\\":       `%\%\_\\%`,
		"by the way": "%by the way%",
	} {
		if got := like(in); got != want {
			t.Errorf("like(%q) = %q, esperava %q", in, got, want)
		}
	}
}

func TestListParamsLeavesEmptyFiltersNull(t *testing.T) {
	p := listParams(domain.Query{Page: 2, Size: 20, Sort: domain.TermAsc})
	if p.RowOffset != 20 || p.RowLimit != 20 || p.ID != nil || p.Unit != nil || p.Cefr.Valid || p.Kind.Valid || p.Pattern.Valid {
		t.Fatalf("filtros vazios devem virar NULL e a pagina 2 pular 20: %+v", p)
	}
}
