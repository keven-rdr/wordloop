// Package catalogtest tem dublês em memória do catálogo; só para testes.
package catalogtest

import (
	"context"

	"github.com/google/uuid"

	"github.com/keven-rdr/wordloop/api/internal/catalog/app"
	"github.com/keven-rdr/wordloop/api/internal/catalog/domain"
)

// Viewers devolve sempre o mesmo Viewer (ou o erro configurado).
type Viewers struct {
	Who domain.Viewer
	Err error
}

var _ app.Viewers = (*Viewers)(nil)

func (v *Viewers) Viewer(context.Context) (domain.Viewer, error) { return v.Who, v.Err }

// Repo guarda o último pedido recebido e devolve o conteúdo configurado.
type Repo struct {
	Page    domain.Page
	Detail  domain.Detail
	UnitsOf []domain.Unit
	Sources []domain.Credit
	Err     error

	LastQuery  domain.Query
	LastDrafts bool
	LastLocale string
}

var _ app.Repository = (*Repo)(nil)

func (r *Repo) Units(_ context.Context, drafts bool) ([]domain.Unit, error) {
	r.LastDrafts = drafts
	return r.UnitsOf, r.Err
}

func (r *Repo) Items(_ context.Context, q domain.Query) (domain.Page, error) {
	r.LastQuery = q
	return r.Page, r.Err
}

func (r *Repo) Item(_ context.Context, id uuid.UUID, locale string, drafts bool) (domain.Detail, error) {
	r.LastDrafts, r.LastLocale = drafts, locale
	if r.Err != nil {
		return domain.Detail{}, r.Err
	}
	if r.Detail.ID != id {
		return domain.Detail{}, app.ErrNotFound
	}
	return r.Detail, nil
}

func (r *Repo) Credits(context.Context) ([]domain.Credit, error) { return r.Sources, r.Err }
