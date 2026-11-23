package app

import (
	"context"
	"sort"

	"github.com/google/uuid"

	"github.com/keven-rdr/wordloop/api/internal/learning/domain"
	"github.com/keven-rdr/wordloop/api/internal/srs"
)

// fill junta cada cartao ao conteudo do catalogo; cartoes cujo item nao esta mais visivel ficam de fora.
func (u PlanSession) fill(ctx context.Context, tx Tx, v domain.Viewer, cards []domain.Card, locale string, seed uuid.UUID) ([]domain.PackCard, error) {
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
		pc.Exercise, pc.ClozeText = domain.PickExercise(c, content)
		if pc.Exercise == domain.Assemble {
			pc.Tiles = domain.Tiles(c, content.Term)
		}
		if pc.Exercise == domain.Choose {
			if pc.Options, err = u.options(ctx, tx, v, c, content, locale, seed); err != nil {
				return nil, err
			}
		}
		out = append(out, pc)
	}
	return out, nil
}

// options monta a resposta certa e os distratores, numa ordem que muda a cada sessao (a resposta nao fica sempre no
// mesmo lugar) e e a mesma dentro da sessao.
func (u PlanSession) options(ctx context.Context, tx Tx, v domain.Viewer, c domain.Card, content domain.Content, locale string, seed uuid.UUID) ([]domain.Option, error) {
	others, err := tx.Distractors(ctx, c.ItemID, locale, v.SeesDrafts(), seed, optionCount)
	if err != nil {
		return nil, err
	}
	opts := append([]domain.Option{{ItemID: c.ItemID, Gloss: content.Gloss}}, others...)
	key := func(o domain.Option) uint32 { return domain.SeedKey(seed, c.ID, o.ItemID) }
	sort.Slice(opts, func(i, j int) bool { return key(opts[i]) < key(opts[j]) })
	return opts, nil
}
