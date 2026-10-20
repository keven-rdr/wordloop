// Package postgres implementa o repositorio do aprendizado com pgx e sqlc.
// O codigo de db/ vem de queries/*.sql e e gerado junto com o do identity (go generate de identity/adapters/postgres).
package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/keven-rdr/wordloop/api/internal/learning/adapters/postgres/db"
	"github.com/keven-rdr/wordloop/api/internal/learning/app"
	"github.com/keven-rdr/wordloop/api/internal/learning/domain"
	"github.com/keven-rdr/wordloop/api/internal/srs"
)

// Store implementa app.Repository: cada InTx e uma transacao do Postgres.
type Store struct{ pool *pgxpool.Pool }

var _ app.Repository = (*Store)(nil)

func NewStore(pool *pgxpool.Pool) *Store { return &Store{pool: pool} }

func (s *Store) InTx(ctx context.Context, fn func(app.Tx) error) error {
	return pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error { return fn(&txn{q: db.New(tx)}) })
}

type txn struct{ q *db.Queries }

var _ app.Tx = (*txn)(nil)

func ts(t time.Time) pgtype.Timestamptz { return pgtype.Timestamptz{Time: t, Valid: !t.IsZero()} }

func when(t pgtype.Timestamptz) time.Time {
	if !t.Valid {
		return time.Time{}
	}
	return t.Time
}

func (t *txn) Settings(ctx context.Context, user uuid.UUID) (domain.Settings, error) {
	if err := t.q.CreateSettings(ctx, user); err != nil {
		return domain.Settings{}, fmt.Errorf("criar ajustes: %w", err)
	}
	s, err := t.q.GetSettings(ctx, user)
	if err != nil {
		return domain.Settings{}, fmt.Errorf("ler ajustes: %w", err)
	}
	return domain.Settings{TZ: s.Tz, CutoffHour: int(s.DayCutoffHour), Scheduler: s.Scheduler}, nil
}

func (t *txn) SaveTZ(ctx context.Context, user uuid.UUID, tz string) error {
	return t.q.SaveTZ(ctx, db.SaveTZParams{UserID: user, Tz: tz})
}

func (t *txn) DueCards(ctx context.Context, user uuid.UUID, now time.Time, limit int) ([]domain.Card, error) {
	rows, err := t.q.DueCards(ctx, db.DueCardsParams{UserID: user, Now: ts(now), RowLimit: int32(limit)}) //nolint:gosec // limite pequeno e fixo
	if err != nil {
		return nil, fmt.Errorf("listar cartoes vencidos: %w", err)
	}
	out := make([]domain.Card, 0, len(rows))
	for _, r := range rows {
		out = append(out, card(r))
	}
	return out, nil
}

func (t *txn) NewItems(ctx context.Context, user uuid.UUID, drafts bool, limit int) ([]uuid.UUID, error) {
	return t.q.NewItems(ctx, db.NewItemsParams{UserID: user, IncludeDrafts: drafts, RowLimit: int32(limit)}) //nolint:gosec // limite pequeno e fixo
}

func (t *txn) ProductiveItems(ctx context.Context, user uuid.UUID, minS float64, before time.Time, limit int) ([]uuid.UUID, error) {
	return t.q.ProductiveItems(ctx, db.ProductiveItemsParams{ //nolint:gosec // limite pequeno e fixo
		UserID: user, MinStability: minS, Before: ts(before), RowLimit: int32(limit), //nolint:gosec // limite pequeno e fixo
	})
}

func (t *txn) CreateCard(ctx context.Context, user, item uuid.UUID, dir domain.Direction, now time.Time) (domain.Card, error) {
	r, err := t.q.CreateCard(ctx, db.CreateCardParams{UserID: user, ItemID: item, Direction: string(dir), Now: ts(now)})
	if err != nil {
		return domain.Card{}, fmt.Errorf("criar cartao: %w", err)
	}
	return card(db.DueCardsRow(r)), nil
}

func (t *txn) Contents(ctx context.Context, items []uuid.UUID, locale string, drafts bool) (map[uuid.UUID]domain.Content, error) {
	rows, err := t.q.CardContents(ctx, db.CardContentsParams{Ids: items, Locale: locale, IncludeDrafts: drafts})
	if err != nil {
		return nil, fmt.Errorf("ler conteudo: %w", err)
	}
	out := make(map[uuid.UUID]domain.Content, len(rows))
	for _, r := range rows {
		out[r.ItemID] = domain.Content{Term: r.Term, Gloss: r.Gloss, Example: r.Example, ExampleGloss: r.ExampleGloss, ImageKey: r.ImageKey}
	}
	return out, nil
}

func (t *txn) Distractors(ctx context.Context, item uuid.UUID, locale string, drafts bool, n int) ([]domain.Option, error) {
	rows, err := t.q.Distractors(ctx, db.DistractorsParams{ItemID: item, Locale: locale, IncludeDrafts: drafts, RowLimit: int32(n)}) //nolint:gosec // n pequeno e fixo
	if err != nil {
		return nil, fmt.Errorf("ler distratores: %w", err)
	}
	out := make([]domain.Option, 0, len(rows))
	for _, r := range rows {
		out = append(out, domain.Option{ItemID: r.ItemID, Gloss: r.Gloss})
	}
	return out, nil
}

func (t *txn) CreateSession(ctx context.Context, id, user uuid.UUID, mode domain.Mode, minutes int, now, expires time.Time) error {
	return t.q.CreateSession(ctx, db.CreateSessionParams{
		ID: id, UserID: user, Mode: string(mode), Minutes: int16(minutes), Now: ts(now), Expires: ts(expires), //nolint:gosec // minutos validados (5 a 60)
	})
}

func (t *txn) LockCard(ctx context.Context, user, id uuid.UUID) (domain.Card, bool, error) {
	r, err := t.q.LockCard(ctx, db.LockCardParams{ID: id, UserID: user})
	if err == pgx.ErrNoRows {
		return domain.Card{}, false, nil
	}
	if err != nil {
		return domain.Card{}, false, fmt.Errorf("travar cartao: %w", err)
	}
	return card(db.DueCardsRow(r)), true, nil
}

func (t *txn) ExpectedTerm(ctx context.Context, item uuid.UUID) (string, error) {
	return t.q.ExpectedTerm(ctx, item)
}

func (t *txn) SaveCard(ctx context.Context, c domain.Card) error {
	s := c.State
	return t.q.SaveCard(ctx, db.SaveCardParams{
		ID: c.ID, State: string(s.State), Stability: s.Stability, DueAt: ts(s.DueAt), LastReviewAt: ts(s.LastReviewAt),
		Reps: int32(s.Reps), Lapses: int32(s.Lapses), SinceLapse: int32(s.SinceLapse), Scheduler: c.Scheduler, //nolint:gosec // contadores pequenos
	})
}

func (t *txn) CardLog(ctx context.Context, id uuid.UUID) ([]srs.Answer, error) {
	rows, err := t.q.CardLog(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("ler o log do cartao: %w", err)
	}
	out := make([]srs.Answer, 0, len(rows))
	for _, r := range rows {
		out = append(out, srs.Answer{At: r.AnsweredAt.Time, Review: srs.Review{
			Correct: r.IsCorrect, Typo: r.TypoFlag, Production: r.ExerciseType == string(domain.Type),
			Hint: r.HintUsed, Explained: r.ExplanationViewed, ImageBefore: r.ImageBeforeAnswer,
			Difficulty: srs.Difficulty(r.DifficultyChosen.String), ErrorKind: srs.ErrorKind(r.ErrorKind.String),
		}})
	}
	return out, nil
}

func card(r db.DueCardsRow) domain.Card {
	return domain.Card{
		ID: r.ID, ItemID: r.ItemID, Direction: domain.Direction(r.Direction), Scheduler: r.Scheduler,
		State: srs.CardState{
			ID: r.ID.String(), State: srs.State(r.State), Stability: r.Stability, Reps: int(r.Reps),
			Lapses: int(r.Lapses), SinceLapse: int(r.SinceLapse), LastReviewAt: when(r.LastReviewAt), DueAt: when(r.DueAt),
		},
	}
}
