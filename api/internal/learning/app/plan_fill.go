package app

import (
	"context"
	"hash/fnv"
	"sort"

	"github.com/google/uuid"

	"github.com/keven-rdr/wordloop/api/internal/learning/domain"
	"github.com/keven-rdr/wordloop/api/internal/srs"
)

// fill junta cada cartao ao conteudo do catalogo; cartoes cujo item nao esta mais visivel ficam de fora.
func (u PlanSession) fill(ctx context.Context, tx Tx, v domain.Viewer, cards []domain.Card, locale string) ([]domain.PackCard, error) {
	ids := make([]uuid.UUID, 0, len(cards))
	for _, c := range cards {
		ids = append(ids, c.ItemID)
	}
	contents, err := tx.Contents(ctx, ids, locale, v.SeesDrafts())
	if err != nil {
		return nil, err
	}
	out := make([]domain.PackCard, 0, len(cards))
	for _, c := range cards {
		content, ok := contents[c.ItemID]
		if !ok {
			continue
		}
		pc := domain.PackCard{Card: c, Content: content, Mastery: srs.Mastery(c.State.Stability)}
		if c.Exercise() == domain.Choose {
			if pc.Options, err = u.options(ctx, tx, v, c, content, locale); err != nil {
				return nil, err
			}
		}
		out = append(out, pc)
	}
	return out, nil
}

// options monta a resposta certa e os distratores, numa ordem fixa por cartao (o mesmo cartao, a mesma ordem).
func (u PlanSession) options(ctx context.Context, tx Tx, v domain.Viewer, c domain.Card, content domain.Content, locale string) ([]domain.Option, error) {
	others, err := tx.Distractors(ctx, c.ItemID, locale, v.SeesDrafts(), optionCount)
	if err != nil {
		return nil, err
	}
	opts := append([]domain.Option{{ItemID: c.ItemID, Gloss: content.Gloss}}, others...)
	sort.Slice(opts, func(i, j int) bool { return orderKey(c.ID, opts[i].ItemID) < orderKey(c.ID, opts[j].ItemID) })
	return opts, nil
}

func orderKey(card, item uuid.UUID) uint32 {
	h := fnv.New32a()
	_, _ = h.Write(card[:])
	_, _ = h.Write(item[:])
	return h.Sum32()
}
