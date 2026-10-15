package httpapi

import (
	"github.com/google/uuid"

	"github.com/keven-rdr/wordloop/api/internal/api/gen"
	"github.com/keven-rdr/wordloop/api/internal/catalog/domain"
)

const mediaPath = "/media/"

func query(p gen.ListItemsParams) domain.Query {
	q := domain.Query{Page: value(p.Page), Size: value(p.Size), Text: value(p.Q)}
	if p.Sort != nil {
		q.Sort = domain.Sort(*p.Sort)
	}
	if p.Unit != nil {
		q.Unit = *p.Unit
	}
	if p.Cefr != nil {
		q.CEFR = string(*p.Cefr)
	}
	if p.Kind != nil {
		q.Kind = string(*p.Kind)
	}
	if p.Locale != nil {
		q.Locale = string(*p.Locale)
	}
	return q
}

func locale(l *gen.GetItemParamsLocale) string {
	if l == nil {
		return ""
	}
	return string(*l)
}

func value[T any](p *T) T {
	var zero T
	if p == nil {
		return zero
	}
	return *p
}

func optional(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func optionalID(id uuid.UUID) *uuid.UUID {
	if id == uuid.Nil {
		return nil
	}
	return &id
}

func mediaURL(key string) *string {
	if key == "" {
		return nil
	}
	url := mediaPath + key
	return &url
}

func itemSummary(s domain.Summary) gen.ItemSummary {
	out := gen.ItemSummary{
		Id: s.ID, Kind: gen.ItemSummaryKind(s.Kind), Term: s.Term, Status: gen.ItemSummaryStatus(s.Status),
		Gloss: optional(s.Gloss), UnitId: optionalID(s.UnitID), ImageUrl: mediaURL(s.ImageKey),
	}
	if s.CEFR != "" {
		c := gen.ItemSummaryCefr(s.CEFR)
		out.Cefr = &c
	}
	return out
}

func itemDetail(d domain.Detail) gen.ItemDetail {
	out := gen.ItemDetail{
		Id: d.ID, Kind: gen.ItemDetailKind(d.Kind), Term: d.Term, Status: gen.ItemDetailStatus(d.Status),
		Gloss: optional(d.Gloss), UnitId: optionalID(d.UnitID),
		Examples: make([]gen.ItemExample, 0, len(d.Examples)),
		Images:   make([]gen.ItemImage, 0, len(d.Images)),
		Related:  make([]gen.RelatedItem, 0, len(d.Related)),
	}
	if d.CEFR != "" {
		c := gen.ItemDetailCefr(d.CEFR)
		out.Cefr = &c
	}
	for _, e := range d.Examples {
		out.Examples = append(out.Examples, gen.ItemExample{Text: e.Text, Translation: optional(e.Translation)})
	}
	for _, m := range d.Images {
		out.Images = append(out.Images, gen.ItemImage{Url: mediaPath + m.Key, Credit: m.Credit})
	}
	for _, r := range d.Related {
		out.Related = append(out.Related, gen.RelatedItem{Id: r.ID, Term: r.Term, Gloss: optional(r.Gloss), Type: gen.RelatedItemType(r.Type)})
	}
	return out
}
