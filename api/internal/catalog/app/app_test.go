package app_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/keven-rdr/wordloop/api/internal/catalog/app"
	"github.com/keven-rdr/wordloop/api/internal/catalog/catalogtest"
	"github.com/keven-rdr/wordloop/api/internal/catalog/domain"
)

func student() *catalogtest.Viewers {
	return &catalogtest.Viewers{Who: domain.Viewer{Authenticated: true}}
}

func admin() *catalogtest.Viewers {
	return &catalogtest.Viewers{Who: domain.Viewer{Authenticated: true, Admin: true}}
}

func TestEveryReadNeedsASessionExceptTheCredits(t *testing.T) {
	ctx, anon, repo := context.Background(), &catalogtest.Viewers{}, &catalogtest.Repo{}
	if _, err := (app.ListUnits{Viewers: anon, Repo: repo}).Run(ctx); !errors.Is(err, app.ErrNoSession) {
		t.Fatalf("unidades sem sessao: %v", err)
	}
	if _, err := (app.ListItems{Viewers: anon, Repo: repo}).Run(ctx, domain.Query{}); !errors.Is(err, app.ErrNoSession) {
		t.Fatalf("itens sem sessao: %v", err)
	}
	if _, err := (app.GetItem{Viewers: anon, Repo: repo}).Run(ctx, uuid.New(), ""); !errors.Is(err, app.ErrNoSession) {
		t.Fatalf("item sem sessao: %v", err)
	}
	if _, err := (app.ListCredits{Repo: repo}).Run(ctx); err != nil {
		t.Fatalf("os creditos sao publicos: %v", err)
	}
}

func TestOnlyAnAdminSeesDrafts(t *testing.T) {
	ctx, repo := context.Background(), &catalogtest.Repo{}
	for _, c := range []struct {
		who    *catalogtest.Viewers
		drafts bool
	}{{student(), false}, {admin(), true}} {
		if _, err := (app.ListItems{Viewers: c.who, Repo: repo}).Run(ctx, domain.Query{Drafts: !c.drafts}); err != nil {
			t.Fatal(err)
		}
		if repo.LastQuery.Drafts != c.drafts {
			t.Fatalf("rascunho vem da sessao, nunca do pedido: quer %v, veio %v", c.drafts, repo.LastQuery.Drafts)
		}
		if _, err := (app.ListUnits{Viewers: c.who, Repo: repo}).Run(ctx); err != nil || repo.LastDrafts != c.drafts {
			t.Fatalf("unidades: %v %v", err, repo.LastDrafts)
		}
	}
}

func TestListingFillsDefaultsAndTrimsTheSearch(t *testing.T) {
	repo := &catalogtest.Repo{}
	if _, err := (app.ListItems{Viewers: student(), Repo: repo}).Run(context.Background(), domain.Query{Text: "  casa "}); err != nil {
		t.Fatal(err)
	}
	q := repo.LastQuery
	if q.Page != 1 || q.Size != app.DefaultSize || q.Sort != domain.TermAsc || q.Locale != "pt-BR" || q.Text != "casa" {
		t.Fatalf("padroes: %+v", q)
	}
}

func TestInvalidListingsAreRefused(t *testing.T) {
	long := make([]byte, app.MaxText+1)
	for i := range long {
		long[i] = 'a'
	}
	for name, q := range map[string]domain.Query{
		"pagina negativa":    {Page: -1},
		"tamanho enorme":     {Size: app.MaxSize + 1},
		"ordem desconhecida": {Sort: "id,asc"},
		"idioma":             {Locale: "fr-FR"},
		"nivel":              {CEFR: "Z9"},
		"tipo":               {Kind: "dialogue"},
		"busca longa":        {Text: string(long)},
	} {
		if _, err := (app.ListItems{Viewers: student(), Repo: &catalogtest.Repo{}}).Run(context.Background(), q); !errors.Is(err, app.ErrInvalidQuery) {
			t.Errorf("%s: esperava consulta invalida, veio %v", name, err)
		}
	}
}

func TestGetItemChecksTheLocaleAndPassesTheDraftRule(t *testing.T) {
	id := uuid.New()
	repo := &catalogtest.Repo{Detail: domain.Detail{Summary: domain.Summary{ID: id}}}
	use := app.GetItem{Viewers: admin(), Repo: repo}
	if _, err := use.Run(context.Background(), id, "fr-FR"); !errors.Is(err, app.ErrInvalidQuery) {
		t.Fatalf("idioma invalido: %v", err)
	}
	if _, err := use.Run(context.Background(), id, ""); err != nil || repo.LastLocale != "pt-BR" || !repo.LastDrafts {
		t.Fatalf("padrao: %v %q %v", err, repo.LastLocale, repo.LastDrafts)
	}
	if _, err := use.Run(context.Background(), uuid.New(), ""); !errors.Is(err, app.ErrNotFound) {
		t.Fatalf("item inexistente: %v", err)
	}
}
