// Package httpapi expoe o catalogo pelo contrato gerado (strict server).
package httpapi

import (
	"context"
	"errors"

	"github.com/keven-rdr/wordloop/api/internal/api/gen"
	"github.com/keven-rdr/wordloop/api/internal/catalog/app"
	"github.com/keven-rdr/wordloop/api/internal/catalog/domain"
)

// Handler junta os casos de uso de leitura do catalogo.
type Handler struct {
	Units   app.ListUnits
	Items   app.ListItems
	Item    app.GetItem
	Credits app.ListCredits
}

func (h Handler) ListUnits(ctx context.Context, _ gen.ListUnitsRequestObject) (gen.ListUnitsResponseObject, error) {
	units, err := h.Units.Run(ctx)
	if errors.Is(err, app.ErrNoSession) {
		return gen.ListUnits401Response{}, nil
	}
	if err != nil {
		return nil, err
	}
	out := make(gen.ListUnits200JSONResponse, 0, len(units))
	for _, u := range units {
		out = append(out, gen.CatalogUnit{Id: u.ID, Position: u.Position, ThemeKey: u.ThemeKey, ItemCount: u.Items})
	}
	return out, nil
}

func (h Handler) ListItems(ctx context.Context, req gen.ListItemsRequestObject) (gen.ListItemsResponseObject, error) {
	page, err := h.Items.Run(ctx, query(req.Params))
	if errors.Is(err, app.ErrNoSession) {
		return gen.ListItems401Response{}, nil
	}
	if errors.Is(err, app.ErrInvalidQuery) {
		return gen.ListItems400Response{}, nil
	}
	if err != nil {
		return nil, err
	}
	items := make([]gen.ItemSummary, 0, len(page.Items))
	for _, s := range page.Items {
		items = append(items, itemSummary(s))
	}
	return gen.ListItems200JSONResponse{Items: items, Page: page.Page, Size: page.Size, Total: page.Total}, nil
}

func (h Handler) GetItem(ctx context.Context, req gen.GetItemRequestObject) (gen.GetItemResponseObject, error) {
	d, err := h.Item.Run(ctx, req.Id, locale(req.Params.Locale))
	switch {
	case errors.Is(err, app.ErrNoSession):
		return gen.GetItem401Response{}, nil
	case errors.Is(err, app.ErrNotFound):
		return gen.GetItem404Response{}, nil
	case errors.Is(err, app.ErrInvalidQuery):
		return gen.GetItem400Response{}, nil
	case err != nil:
		return nil, err
	}
	return gen.GetItem200JSONResponse(itemDetail(d)), nil
}

func (h Handler) ListCredits(ctx context.Context, _ gen.ListCreditsRequestObject) (gen.ListCreditsResponseObject, error) {
	credits, err := h.Credits.Run(ctx)
	if err != nil {
		return nil, err
	}
	out := make(gen.ListCredits200JSONResponse, 0, len(credits))
	for _, c := range credits {
		out = append(out, creditOf(c))
	}
	return out, nil
}

func creditOf(c domain.Credit) gen.Credit {
	return gen.Credit{Code: c.Code, Name: c.Name, Url: optional(c.URL), License: c.License, Notice: c.Notice}
}
