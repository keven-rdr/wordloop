// Package learningtest tem dubles em memoria do aprendizado; so para testes.
package learningtest

import (
	"context"
	"errors"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/keven-rdr/wordloop/api/internal/learning/app"
	"github.com/keven-rdr/wordloop/api/internal/learning/domain"
	"github.com/keven-rdr/wordloop/api/internal/srs"
)

// Viewers devolve sempre o mesmo Viewer (ou o erro configurado).
type Viewers struct {
	Who domain.Viewer
	Err error
}

var _ app.Viewers = (*Viewers)(nil)

func (v *Viewers) Viewer(context.Context) (domain.Viewer, error) {
	if v.Err != nil {
		return domain.Viewer{}, v.Err
	}
	if v.Who.UserID == uuid.Nil {
		return domain.Viewer{}, app.ErrNoSession
	}
	return v.Who, nil
}

// Item e um item do catalogo de mentira.
type Item struct {
	ID      uuid.UUID
	Content domain.Content
	Draft   bool
}

// Store e um repositorio em memoria: dados do catalogo, cartoes, log e sessoes de uma pessoa.
type Store struct {
	Items    []Item // na ordem de estudo
	Settings domain.Settings
	Cards    map[uuid.UUID]*domain.Card
	Log      []domain.LogRow
	Sessions []uuid.UUID
	FailOn   string // nome do metodo que deve falhar (para testar o erro)
	Related  map[uuid.UUID]uuid.UUID
}

var _ app.Repository = (*Store)(nil)

// New cria um repositorio com os itens, os ajustes padrao e nenhum cartao.
func New(items ...Item) *Store {
	return &Store{
		Items:    items,
		Settings: domain.Settings{TZ: "America/Sao_Paulo", CutoffHour: 4, Scheduler: "simple_v1"},
		Cards:    map[uuid.UUID]*domain.Card{},
		Related:  map[uuid.UUID]uuid.UUID{},
	}
}

var ErrBoom = errors.New("falha simulada")

// InTx roda fn; se falhar, desfaz o que ela mudou (copia e restaura).
func (s *Store) InTx(_ context.Context, fn func(app.Tx) error) error {
	cards := map[uuid.UUID]domain.Card{}
	for id, c := range s.Cards {
		cards[id] = *c
	}
	logLen, sessions, settings := len(s.Log), len(s.Sessions), s.Settings
	if err := fn(&tx{s: s}); err != nil {
		s.Cards = map[uuid.UUID]*domain.Card{}
		for id, c := range cards {
			c := c
			s.Cards[id] = &c
		}
		s.Log, s.Sessions, s.Settings = s.Log[:logLen], s.Sessions[:sessions], settings
		return err
	}
	return nil
}

type tx struct{ s *Store }

var _ app.Tx = (*tx)(nil)

func (t *tx) fail(method string) error {
	if t.s.FailOn == method {
		return ErrBoom
	}
	return nil
}

func (t *tx) Settings(context.Context, uuid.UUID) (domain.Settings, error) {
	return t.s.Settings, t.fail("Settings")
}

func (t *tx) SaveTZ(_ context.Context, _ uuid.UUID, tz string) error {
	t.s.Settings.TZ = tz
	return t.fail("SaveTZ")
}

func (t *tx) DueCards(_ context.Context, _ uuid.UUID, now time.Time, limit int) ([]domain.Card, error) {
	var out []domain.Card
	for _, c := range t.s.Cards {
		if c.State.State != srs.StateNew && !c.State.DueAt.After(now) {
			out = append(out, *c)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID.String() < out[j].ID.String() })
	return out[:min(len(out), limit)], t.fail("DueCards")
}

func (t *tx) visible(it Item, drafts bool) bool { return !it.Draft || drafts }

func (t *tx) hasCard(item uuid.UUID, dir domain.Direction) (*domain.Card, bool) {
	for _, c := range t.s.Cards {
		if c.ItemID == item && c.Direction == dir {
			return c, true
		}
	}
	return nil, false
}

func (t *tx) NewItems(_ context.Context, _ uuid.UUID, drafts bool, limit int) ([]uuid.UUID, error) {
	var out []uuid.UUID
	for _, it := range t.s.Items {
		c, has := t.hasCard(it.ID, domain.Receptive)
		if t.visible(it, drafts) && (!has || c.State.State == srs.StateNew) && len(out) < limit {
			out = append(out, it.ID)
		}
	}
	return out, t.fail("NewItems")
}

func (t *tx) ProductiveItems(_ context.Context, _ uuid.UUID, minS float64, before time.Time, limit int) ([]uuid.UUID, error) {
	var out []uuid.UUID
	for _, c := range t.s.Cards {
		if c.Direction != domain.Receptive || c.State.Stability < minS || c.State.LastReviewAt.After(before) || c.State.State == srs.StateNew {
			continue
		}
		if p, has := t.hasCard(c.ItemID, domain.Productive); (!has || p.State.State == srs.StateNew) && len(out) < limit {
			out = append(out, c.ItemID)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].String() < out[j].String() })
	return out, t.fail("ProductiveItems")
}

func (t *tx) CreateCard(_ context.Context, _ uuid.UUID, item uuid.UUID, dir domain.Direction, _ time.Time) (domain.Card, error) {
	if c, ok := t.hasCard(item, dir); ok {
		return *c, nil
	}
	id := uuid.New()
	c := &domain.Card{ID: id, ItemID: item, Direction: dir, Scheduler: "simple_v1", State: srs.CardState{ID: id.String(), State: srs.StateNew}}
	t.s.Cards[id] = c
	return *c, t.fail("CreateCard")
}

func (t *tx) Contents(_ context.Context, items []uuid.UUID, _ string, drafts bool) (map[uuid.UUID]domain.Content, error) {
	out := map[uuid.UUID]domain.Content{}
	for _, id := range items {
		for _, it := range t.s.Items {
			if it.ID == id && t.visible(it, drafts) {
				out[id] = it.Content
			}
		}
	}
	return out, t.fail("Contents")
}

func (t *tx) Distractors(_ context.Context, item uuid.UUID, _ string, drafts bool, n int) ([]domain.Option, error) {
	var out []domain.Option
	for _, it := range t.s.Items {
		if it.ID != item && t.visible(it, drafts) && len(out) < n && t.s.Related[item] != it.ID {
			out = append(out, domain.Option{ItemID: it.ID, Gloss: it.Content.Gloss})
		}
	}
	return out, t.fail("Distractors")
}

func (t *tx) CreateSession(_ context.Context, id, _ uuid.UUID, _ domain.Mode, _ int, _, _ time.Time) error {
	t.s.Sessions = append(t.s.Sessions, id)
	return t.fail("CreateSession")
}

func (t *tx) LockCard(_ context.Context, _ uuid.UUID, id uuid.UUID) (domain.Card, bool, error) {
	c, ok := t.s.Cards[id]
	if !ok {
		return domain.Card{}, false, t.fail("LockCard")
	}
	return *c, true, t.fail("LockCard")
}

func (t *tx) ExpectedTerm(_ context.Context, item uuid.UUID) (string, error) {
	for _, it := range t.s.Items {
		if it.ID == item {
			return it.Content.Term, t.fail("ExpectedTerm")
		}
	}
	return "", t.fail("ExpectedTerm")
}

func (t *tx) InsertReview(_ context.Context, row domain.LogRow) (bool, error) {
	if err := t.fail("InsertReview"); err != nil {
		return false, err
	}
	for _, r := range t.s.Log {
		if r.ID == row.ID {
			return false, nil
		}
	}
	t.s.Log = append(t.s.Log, row)
	return true, nil
}

func (t *tx) SaveCard(_ context.Context, c domain.Card) error {
	cp := c
	t.s.Cards[c.ID] = &cp
	return t.fail("SaveCard")
}

func (t *tx) CardLog(_ context.Context, card uuid.UUID) ([]srs.Answer, error) {
	var out []srs.Answer
	for _, r := range t.s.Log {
		if r.CardID == card {
			out = append(out, srs.Answer{At: r.AnsweredAt, Review: srs.Review{
				Correct: r.Correct, Typo: r.Typo, Production: r.Production, Hint: r.HintUsed, Explained: r.Explained,
				ImageBefore: r.ImageBefore, Difficulty: r.Difficulty, ErrorKind: r.ErrorKind,
			}})
		}
	}
	return out, t.fail("CardLog")
}

// Lines devolve os ids dos itens na ordem, para os testes montarem pedidos.
func (s *Store) Lines() string {
	ids := make([]string, 0, len(s.Items))
	for _, it := range s.Items {
		ids = append(ids, it.Content.Term)
	}
	return strings.Join(ids, ",")
}
