package postgres

import (
	"context"
	"fmt"
	"math"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/keven-rdr/wordloop/api/internal/catalog/adapters/postgres/db"
	"github.com/keven-rdr/wordloop/api/internal/catalog/app"
	"github.com/keven-rdr/wordloop/api/internal/catalog/domain"
)

const fallbackLocale = "pt-BR"

// Catalog implementa app.Repository sobre o Postgres.
type Catalog struct{ q *db.Queries }

var _ app.Repository = (*Catalog)(nil)

// NewCatalog recebe o pool (ou uma transacao).
func NewCatalog(conn db.DBTX) *Catalog { return &Catalog{q: db.New(conn)} }

func (c *Catalog) Units(ctx context.Context, drafts bool) ([]domain.Unit, error) {
	rows, err := c.q.ListUnits(ctx, drafts)
	if err != nil {
		return nil, fmt.Errorf("listar unidades: %w", err)
	}
	out := make([]domain.Unit, 0, len(rows))
	for _, r := range rows {
		out = append(out, domain.Unit{ID: r.ID, Position: int(r.Position), ThemeKey: r.ThemeKey, Items: int(r.ItemCount)})
	}
	return out, nil
}

func (c *Catalog) Items(ctx context.Context, q domain.Query) (domain.Page, error) {
	rows, err := c.q.ListItems(ctx, listParams(q))
	if err != nil {
		return domain.Page{}, fmt.Errorf("listar itens: %w", err)
	}
	page := domain.Page{Items: make([]domain.Summary, 0, len(rows)), Page: q.Page, Size: q.Size}
	for _, r := range rows {
		page.Total = int(r.Total)
		page.Items = append(page.Items, summary(r))
	}
	return page, nil
}

func (c *Catalog) Item(ctx context.Context, id uuid.UUID, locale string, drafts bool) (domain.Detail, error) {
	page, err := c.Items(ctx, domain.Query{ID: id, Page: 1, Size: 1, Sort: domain.TermAsc, Locale: locale, Drafts: drafts})
	if err != nil {
		return domain.Detail{}, err
	}
	if len(page.Items) == 0 {
		return domain.Detail{}, app.ErrNotFound
	}
	detail := domain.Detail{Summary: page.Items[0]}
	if err := c.fill(ctx, &detail, locale, drafts); err != nil {
		return domain.Detail{}, err
	}
	return detail, nil
}

func (c *Catalog) fill(ctx context.Context, d *domain.Detail, locale string, drafts bool) error {
	examples, err := c.q.ListItemExamples(ctx, db.ListItemExamplesParams{ItemID: d.ID, Locale: locale, FallbackLocale: fallbackLocale})
	if err != nil {
		return fmt.Errorf("listar exemplos: %w", err)
	}
	for _, e := range examples {
		d.Examples = append(d.Examples, domain.Example{Text: e.Text, Translation: e.Translation})
	}
	images, err := c.q.ListItemImages(ctx, db.ListItemImagesParams{ItemID: d.ID, IncludeDrafts: drafts})
	if err != nil {
		return fmt.Errorf("listar imagens: %w", err)
	}
	for _, m := range images {
		d.Images = append(d.Images, domain.Image{Key: m.StorageKey, Credit: m.AttributionText})
	}
	related, err := c.q.ListRelatedItems(ctx, db.ListRelatedItemsParams{ItemID: d.ID, IncludeDrafts: drafts, Locale: locale})
	if err != nil {
		return fmt.Errorf("listar relacionados: %w", err)
	}
	for _, r := range related {
		d.Related = append(d.Related, domain.Related{ID: r.ItemID, Term: r.Term, Gloss: r.Gloss, Type: r.Type})
	}
	return nil
}

func (c *Catalog) Credits(ctx context.Context) ([]domain.Credit, error) {
	rows, err := c.q.ListCredits(ctx)
	if err != nil {
		return nil, fmt.Errorf("listar creditos: %w", err)
	}
	out := make([]domain.Credit, 0, len(rows))
	for _, r := range rows {
		out = append(out, domain.Credit{Code: r.Code, Name: r.Name, URL: r.Url, License: r.LicenseSpdx, Notice: r.AttributionNotice})
	}
	return out, nil
}

func listParams(q domain.Query) db.ListItemsParams {
	return db.ListItemsParams{
		IncludeDrafts: q.Drafts, Locale: q.Locale,
		ID: optionalID(q.ID), Unit: optionalID(q.Unit),
		Cefr: optionalText(q.CEFR), Kind: optionalText(q.Kind), Pattern: optionalText(like(q.Text)),
		Sort:      string(q.Sort),
		RowLimit:  clamp32(q.Size),
		RowOffset: clamp32((q.Page - 1) * q.Size),
	}
}

// clamp32 converte para int32 sem estourar; a camada de aplicacao ja limita pagina e tamanho.
func clamp32(n int) int32 {
	if n < 0 || n > math.MaxInt32 {
		return 0
	}
	return int32(n)
}

func summary(r db.ListItemsRow) domain.Summary {
	s := domain.Summary{ID: r.ID, Kind: r.Kind, Term: r.Term, Gloss: r.Gloss, CEFR: r.Cefr.String, Status: r.Status, ImageKey: r.ImageKey}
	if r.UnitID != nil {
		s.UnitID = *r.UnitID
	}
	return s
}

func optionalID(id uuid.UUID) *uuid.UUID {
	if id == uuid.Nil {
		return nil
	}
	return &id
}

func optionalText(s string) pgtype.Text { return pgtype.Text{String: s, Valid: s != ""} }

var likeEscaper = strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)

// like transforma o texto digitado em um padrao ILIKE que acha em qualquer parte, sem tratar % e _ como curingas.
func like(text string) string {
	if text == "" {
		return ""
	}
	return "%" + likeEscaper.Replace(text) + "%"
}
