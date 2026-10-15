package app

import (
	"context"
	"slices"
	"strings"

	"github.com/keven-rdr/wordloop/api/internal/catalog/domain"
)

const (
	DefaultSize = 20
	MaxSize     = 100
	MaxText     = 100
	DefaultLang = "pt-BR"
)

type ListItems struct {
	Viewers Viewers
	Repo    Repository
}

func (u ListItems) Run(ctx context.Context, q domain.Query) (domain.Page, error) {
	v, err := signedIn(ctx, u.Viewers)
	if err != nil {
		return domain.Page{}, err
	}
	q, err = normalize(q)
	if err != nil {
		return domain.Page{}, err
	}
	q.Drafts = v.SeesDrafts()
	return u.Repo.Items(ctx, q)
}

func normalize(q domain.Query) (domain.Query, error) {
	q = withDefaults(q)
	if !valid(q) {
		return domain.Query{}, ErrInvalidQuery
	}
	return q, nil
}

func withDefaults(q domain.Query) domain.Query {
	if q.Page == 0 {
		q.Page = 1
	}
	if q.Size == 0 {
		q.Size = DefaultSize
	}
	if q.Sort == "" {
		q.Sort = domain.TermAsc
	}
	if q.Locale == "" {
		q.Locale = DefaultLang
	}
	q.Text = strings.TrimSpace(q.Text)
	return q
}

func valid(q domain.Query) bool {
	inRange := q.Page >= 1 && q.Size >= 1 && q.Size <= MaxSize && len(q.Text) <= MaxText
	return inRange && q.Sort.Valid() && ValidLocale(q.Locale) && oneOf(q.CEFR, cefrs) && oneOf(q.Kind, kinds)
}

var (
	cefrs   = []string{"A1", "A2", "B1", "B2", "C1", "C2"}
	kinds   = []string{"word", "expression", "phrasal_verb", "sentence"}
	locales = []string{"pt-BR", "en-US"}
)

// ValidLocale diz se o idioma da traducao e um dos que o conteudo tem.
func ValidLocale(l string) bool { return slices.Contains(locales, l) }

// oneOf aceita o valor vazio (sem filtro) ou um dos permitidos.
func oneOf(v string, allowed []string) bool { return v == "" || slices.Contains(allowed, v) }
