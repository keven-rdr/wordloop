//go:build integration

package postgres_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/keven-rdr/wordloop/api/internal/catalog/adapters/postgres"
	"github.com/keven-rdr/wordloop/api/internal/catalog/app"
	"github.com/keven-rdr/wordloop/api/internal/catalog/domain"
	"github.com/keven-rdr/wordloop/api/internal/platform/db"
	"github.com/keven-rdr/wordloop/api/internal/platform/db/dbtest"
	"github.com/keven-rdr/wordloop/api/migrations"
)

// setup sobe um Postgres real com as migracoes e o seed v0 (tudo em rascunho).
func setup(t *testing.T) (*postgres.Catalog, *pgxpool.Pool) {
	t.Helper()
	ctx := context.Background()
	url, _ := dbtest.Postgres(t)
	pool, err := db.Open(ctx, url)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	if err := db.Migrate(ctx, pool, migrations.FS); err != nil {
		t.Fatal(err)
	}
	return postgres.NewCatalog(pool), pool
}

func publish(t *testing.T, pool *pgxpool.Pool, naturalKey string) {
	t.Helper()
	if _, err := pool.Exec(context.Background(), `UPDATE item SET status = 'published' WHERE natural_key = $1`, naturalKey); err != nil {
		t.Fatal(err)
	}
}

// pendingMedia volta as imagens a "pending", como antes da migracao que as aprova.
func pendingMedia(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	if _, err := pool.Exec(context.Background(), `UPDATE media SET status = 'pending'`); err != nil {
		t.Fatal(err)
	}
}

func list(t *testing.T, c *postgres.Catalog, q domain.Query) domain.Page {
	t.Helper()
	if q.Page == 0 {
		q.Page, q.Size, q.Sort, q.Locale = 1, 200, domain.TermAsc, "pt-BR"
	}
	page, err := c.Items(context.Background(), q)
	if err != nil {
		t.Fatal(err)
	}
	return page
}

func TestTheSeedStartsAsDraftAndOnlyAdminsSeeIt(t *testing.T) {
	c, pool := setup(t)
	if got := list(t, c, domain.Query{}); got.Total != 0 || len(got.Items) != 0 {
		t.Fatalf("o aluno nao ve rascunho: %+v", got)
	}
	admin := list(t, c, domain.Query{Drafts: true})
	if admin.Total != 125 || len(admin.Items) != 125 {
		t.Fatalf("o admin ve o seed inteiro (125 itens): %d/%d", len(admin.Items), admin.Total)
	}
	publish(t, pool, "word|house|noun|1")
	got := list(t, c, domain.Query{})
	if got.Total != 1 || got.Items[0].Term != "house" || got.Items[0].Gloss != "casa" || got.Items[0].CEFR != "A1" {
		t.Fatalf("depois de publicar: %+v", got)
	}
	if got.Items[0].ImageKey != "openmoji/1F3E0.svg" {
		t.Fatalf("a imagem aprovada aparece para o aluno: %q", got.Items[0].ImageKey)
	}
	pendingMedia(t, pool)
	got = list(t, c, domain.Query{})
	if got.Items[0].ImageKey != "" {
		t.Fatalf("imagem pendente nao aparece para o aluno: %q", got.Items[0].ImageKey)
	}
	if admin := list(t, c, domain.Query{Drafts: true, Text: "house"}); admin.Items[0].ImageKey != "openmoji/1F3E0.svg" {
		t.Fatalf("o admin ve a imagem pendente: %+v", admin.Items)
	}
}

func TestSearchFiltersAndSorting(t *testing.T) {
	c, _ := setup(t)
	found := false
	for _, it := range list(t, c, domain.Query{Drafts: true, Text: "CASA"}).Items {
		found = found || it.Term == "house"
	}
	if !found {
		t.Fatal("busca por traducao sem diferenciar maiusculas deveria achar house")
	}
	if got := list(t, c, domain.Query{Drafts: true, Text: "%"}); got.Total != 0 {
		t.Fatalf("o %% digitado nao e curinga: %d", got.Total)
	}
	if got := list(t, c, domain.Query{Drafts: true, Kind: "phrasal_verb"}); got.Total != 3 {
		t.Fatalf("phrasal verbs: %d", got.Total)
	}
	asc := list(t, c, domain.Query{Drafts: true, Kind: "word", Sort: domain.TermAsc, Page: 1, Size: 3, Locale: "pt-BR"})
	desc := list(t, c, domain.Query{Drafts: true, Kind: "word", Sort: domain.TermDesc, Page: 1, Size: 3, Locale: "pt-BR"})
	if asc.Items[0].Term >= desc.Items[0].Term {
		t.Fatalf("ordem crescente e decrescente iguais: %v / %v", asc.Items[0].Term, desc.Items[0].Term)
	}
	if asc.Total != 105 || len(asc.Items) != 3 {
		t.Fatalf("paginacao: %d itens de %d", len(asc.Items), asc.Total)
	}
	second := list(t, c, domain.Query{Drafts: true, Kind: "word", Sort: domain.TermAsc, Page: 2, Size: 3, Locale: "pt-BR"})
	if second.Items[0].ID == asc.Items[0].ID {
		t.Fatal("a segunda pagina repetiu a primeira")
	}
}

func TestEnglishHelpFallsBackToThePortugueseGloss(t *testing.T) {
	c, _ := setup(t)
	got := list(t, c, domain.Query{Drafts: true, Text: "house", Page: 1, Size: 5, Sort: domain.TermAsc, Locale: "en-US"})
	if len(got.Items) == 0 || got.Items[0].Gloss != "casa" {
		t.Fatalf("sem traducao em ingles, cai para a do curso: %+v", got.Items)
	}
}

func TestItemDetailHasExamplesRelatedAndCredit(t *testing.T) {
	c, pool := setup(t)
	ctx := context.Background()
	var id uuid.UUID
	if err := pool.QueryRow(ctx, `SELECT id FROM item WHERE natural_key = 'word|hot|adj|1'`).Scan(&id); err != nil {
		t.Fatal(err)
	}
	if _, err := c.Item(ctx, id, "pt-BR", false); !errors.Is(err, app.ErrNotFound) {
		t.Fatalf("rascunho para o aluno e 404: %v", err)
	}
	d, err := c.Item(ctx, id, "pt-BR", true)
	if err != nil || d.Term != "hot" || len(d.Examples) != 1 || d.Examples[0].Translation == "" || len(d.Images) != 1 || d.Images[0].Credit == "" {
		t.Fatalf("detalhe: %+v %v", d, err)
	}
	if len(d.Related) == 0 {
		t.Fatalf("hot tem antonimo e colocacao: %+v", d.Related)
	}
	for _, r := range d.Related {
		if r.ID == id || r.Term == "" {
			t.Fatalf("relacionado invalido: %+v", r)
		}
	}
	publish(t, pool, "word|hot|adj|1")
	pendingMedia(t, pool)
	if d, err = c.Item(ctx, id, "pt-BR", false); err != nil || len(d.Related) != 0 || len(d.Images) != 0 {
		t.Fatalf("relacionado em rascunho e imagem pendente nao vazam para o aluno: %+v %v", d, err)
	}
}

func TestUnitsCountOnlyWhatTheViewerSees(t *testing.T) {
	c, pool := setup(t)
	ctx := context.Background()
	student, err := c.Units(ctx, false)
	if err != nil || len(student) != 6 || student[0].Items != 0 {
		t.Fatalf("aluno: %+v %v", student, err)
	}
	admin, _ := c.Units(ctx, true)
	if admin[0].Items == 0 || admin[0].ThemeKey != "home-and-family" {
		t.Fatalf("admin: %+v", admin)
	}
	publish(t, pool, "word|house|noun|1")
	if student, _ = c.Units(ctx, false); student[0].Items != 1 {
		t.Fatalf("depois de publicar: %+v", student)
	}
}

func TestCreditsListEverySource(t *testing.T) {
	c, _ := setup(t)
	credits, err := c.Credits(context.Background())
	if err != nil || len(credits) != 2 {
		t.Fatalf("creditos: %+v %v", credits, err)
	}
	for _, cr := range credits {
		if cr.License == "" || cr.Notice == "" {
			t.Fatalf("credito sem licenca ou texto: %+v", cr)
		}
	}
}
