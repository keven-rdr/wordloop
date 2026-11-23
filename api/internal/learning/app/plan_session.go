package app

import (
	"context"
	"sort"
	"time"

	"github.com/google/uuid"

	"github.com/keven-rdr/wordloop/api/internal/learning/domain"
	"github.com/keven-rdr/wordloop/api/internal/srs"
)

const (
	dueFetchLimit = 500 // quantas vencidas ler para ordenar por R
	optionCount   = 3   // distratores por pergunta de escolher
	defaultLocale = "pt-BR"
)

// PlanRequest e o pedido de "quanto tempo hoje?".
type PlanRequest struct {
	Minutes int
	Mode    domain.Mode
	TZ      string
	Locale  string
}

// PlanSession monta a sessao: revisoes vencidas (menor R primeiro) e palavras novas, com o conteudo.
type PlanSession struct {
	Viewers   Viewers
	Repo      Repository
	Scheduler srs.Scheduler
	Now       Clock
	NewID     func() uuid.UUID
}

func (u PlanSession) Run(ctx context.Context, req PlanRequest) (domain.Pack, error) {
	v, err := signedIn(ctx, u.Viewers)
	if err != nil {
		return domain.Pack{}, err
	}
	if req.Mode == "" {
		req.Mode = domain.ModeMixed
	}
	if req.Locale == "" {
		req.Locale = defaultLocale
	}
	var pack domain.Pack
	err = u.Repo.InTx(ctx, func(tx Tx) error {
		var err error
		pack, err = u.plan(ctx, tx, v, req)
		return err
	})
	return pack, err
}

func (u PlanSession) plan(ctx context.Context, tx Tx, v domain.Viewer, req PlanRequest) (domain.Pack, error) {
	now := u.Now.now()
	if err := u.rememberTZ(ctx, tx, v.UserID, req.TZ); err != nil {
		return domain.Pack{}, err
	}
	due, err := tx.DueCards(ctx, v.UserID, now, dueFetchLimit)
	if err != nil {
		return domain.Pack{}, err
	}
	limits, err := domain.PlanLimits(req.Minutes, req.Mode, len(due))
	if err != nil {
		return domain.Pack{}, err
	}
	id := u.NewID() // tambem e a semente da sessao: ordem, palavras novas e distratores mudam a cada sessao
	reviews := u.byLowestR(due, now)[:min(len(due), limits.Reviews)]
	fresh, err := u.newCards(ctx, tx, v, limits, now, id)
	if err != nil {
		return domain.Pack{}, err
	}
	cards := domain.Mix(reviews, fresh, id)
	expires := domain.Expiry(now)
	if err := tx.CreateSession(ctx, id, v.UserID, req.Mode, req.Minutes, now, expires); err != nil {
		return domain.Pack{}, err
	}
	packed, err := u.fill(ctx, tx, v, cards, req.Locale, id)
	if err != nil {
		return domain.Pack{}, err
	}
	return domain.Pack{SessionID: id, Cards: packed, NewLimit: limits.New, ReviewLimit: limits.Reviews, ExpiresAt: expires}, nil
}

// rememberTZ guarda o fuso informado pelo aparelho quando e valido, diferente do guardado e a pessoa deixou o fuso
// acompanhar o aparelho (tz_auto); fuso escolhido nas configuracoes nao e mexido.
func (u PlanSession) rememberTZ(ctx context.Context, tx Tx, user uuid.UUID, tz string) error {
	s, err := tx.Settings(ctx, user)
	if err != nil {
		return err
	}
	if s.TZAuto && tz != "" && tz != s.TZ && domain.ValidTZ(tz) {
		return tx.SaveTZ(ctx, user, tz)
	}
	return nil
}

func (u PlanSession) byLowestR(cards []domain.Card, now time.Time) []domain.Card {
	r := func(c domain.Card) float64 { return u.Scheduler.Retrievability(c.State, now) }
	sort.SliceStable(cards, func(i, j int) bool {
		if ri, rj := r(cards[i]), r(cards[j]); ri != rj {
			return ri < rj
		}
		return cards[i].ID.String() < cards[j].ID.String()
	})
	return cards
}

// newCards cria os cartoes novos da sessao: os produtivos que ja podem nascer (cota propria) e as palavras novas
// (da licao atual e algumas de desafio, de licoes adiante).
func (u PlanSession) newCards(ctx context.Context, tx Tx, v domain.Viewer, l domain.Limits, now time.Time, seed uuid.UUID) ([]domain.Card, error) {
	var productive []uuid.UUID
	if l.Productive > 0 {
		var err error
		productive, err = tx.ProductiveItems(ctx, v.UserID, domain.ProductiveFrom, domain.ProductiveFromSentence, now.Add(-24*time.Hour), l.Productive)
		if err != nil {
			return nil, err
		}
	}
	fresh, err := u.newItems(ctx, tx, v, l.New, seed)
	if err != nil {
		return nil, err
	}
	out, err := u.create(ctx, tx, v.UserID, productive, domain.Productive, now)
	if err != nil {
		return nil, err
	}
	receptive, err := u.create(ctx, tx, v.UserID, fresh, domain.Receptive, now)
	return append(out, receptive...), err
}

// newItems escolhe as n palavras novas: as da licao atual em ordem sorteada e 1 em 5 de desafio.
func (u PlanSession) newItems(ctx context.Context, tx Tx, v domain.Viewer, n int, seed uuid.UUID) ([]uuid.UUID, error) {
	if n == 0 {
		return nil, nil
	}
	main, err := tx.NewItems(ctx, v.UserID, v.SeesDrafts(), seed, n)
	if err != nil {
		return nil, err
	}
	var stretch []uuid.UUID
	if k := domain.StretchCount(n); k > 0 {
		if stretch, err = tx.StretchItems(ctx, v.UserID, v.SeesDrafts(), seed, k); err != nil {
			return nil, err
		}
	}
	return domain.PickNew(main, stretch, n), nil
}

func (u PlanSession) create(ctx context.Context, tx Tx, user uuid.UUID, items []uuid.UUID, dir domain.Direction, now time.Time) ([]domain.Card, error) {
	out := make([]domain.Card, 0, len(items))
	for _, item := range items {
		c, err := tx.CreateCard(ctx, user, item, dir, now)
		if err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, nil
}
