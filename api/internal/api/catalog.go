package api

import (
	"context"

	"github.com/keven-rdr/wordloop/api/internal/api/gen"
)

// Catalog e o modulo de conteudo (somente leitura); nulo = catalogo indisponivel.
type Catalog interface {
	ListUnits(ctx context.Context, req gen.ListUnitsRequestObject) (gen.ListUnitsResponseObject, error)
	ListItems(ctx context.Context, req gen.ListItemsRequestObject) (gen.ListItemsResponseObject, error)
	GetItem(ctx context.Context, req gen.GetItemRequestObject) (gen.GetItemResponseObject, error)
	ListCredits(ctx context.Context, req gen.ListCreditsRequestObject) (gen.ListCreditsResponseObject, error)
}

func (s Server) ListUnits(ctx context.Context, req gen.ListUnitsRequestObject) (gen.ListUnitsResponseObject, error) {
	if s.Catalog == nil {
		return gen.ListUnits401Response{}, nil
	}
	return s.Catalog.ListUnits(ctx, req)
}

func (s Server) ListItems(ctx context.Context, req gen.ListItemsRequestObject) (gen.ListItemsResponseObject, error) {
	if s.Catalog == nil {
		return gen.ListItems401Response{}, nil
	}
	return s.Catalog.ListItems(ctx, req)
}

func (s Server) GetItem(ctx context.Context, req gen.GetItemRequestObject) (gen.GetItemResponseObject, error) {
	if s.Catalog == nil {
		return gen.GetItem401Response{}, nil
	}
	return s.Catalog.GetItem(ctx, req)
}

func (s Server) ListCredits(ctx context.Context, req gen.ListCreditsRequestObject) (gen.ListCreditsResponseObject, error) {
	if s.Catalog == nil {
		return gen.ListCredits200JSONResponse{}, nil
	}
	return s.Catalog.ListCredits(ctx, req)
}
