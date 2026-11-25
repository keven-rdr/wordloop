// Package postgres le os pontos (study_day, goal_cycle, pela visao user_day_score) e grava o perfil do ranking.
package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/keven-rdr/wordloop/api/internal/leaderboard/adapters/postgres/db"
	"github.com/keven-rdr/wordloop/api/internal/leaderboard/app"
	"github.com/keven-rdr/wordloop/api/internal/leaderboard/domain"
)

// uniqueViolation e o codigo do Postgres para chave unica repetida (aqui, o apelido).
const uniqueViolation = "23505"

// Store implementa app.Store.
type Store struct{ q *db.Queries }

var _ app.Store = (*Store)(nil)

func NewStore(pool *pgxpool.Pool) *Store { return &Store{q: db.New(pool)} }

func date(t time.Time) pgtype.Date { return pgtype.Date{Time: t, Valid: true} }

// Clock e o fuso e o corte do dia da pessoa; sem ajustes ainda, UTC sem corte.
func (s *Store) Clock(ctx context.Context, user uuid.UUID) (string, int, error) {
	r, err := s.q.GetSettings(ctx, user)
	if errors.Is(err, pgx.ErrNoRows) {
		return "UTC", 0, nil
	}
	if err != nil {
		return "", 0, fmt.Errorf("ler fuso: %w", err)
	}
	return r.Tz, int(r.DayCutoffHour), nil
}

func (s *Store) Ranking(ctx context.Context, me uuid.UUID, from, to time.Time, limit int) ([]domain.Entry, error) {
	rows, err := s.q.Ranking(ctx, db.RankingParams{Me: me, FromDate: date(from), ToDate: date(to), RowLimit: int32(limit)}) //nolint:gosec // limite de 0 a 100
	if err != nil {
		return nil, fmt.Errorf("ler ranking: %w", err)
	}
	out := make([]domain.Entry, 0, len(rows))
	for _, r := range rows {
		out = append(out, domain.Entry{UserID: r.UserID, Position: int(r.Position), Nickname: r.Nickname, Points: int(r.Points)})
	}
	return out, nil
}

func (s *Store) DaysMet(ctx context.Context, user uuid.UUID, from, to time.Time) (int, error) {
	n, err := s.q.DaysMet(ctx, db.DaysMetParams{UserID: user, FromDate: date(from), ToDate: date(to)})
	if err != nil {
		return 0, fmt.Errorf("contar dias da meta: %w", err)
	}
	return int(n), nil
}

func (s *Store) Profile(ctx context.Context, user uuid.UUID) (domain.Profile, bool, error) {
	r, err := s.q.GetProfile(ctx, user)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Profile{}, false, nil
	}
	if err != nil {
		return domain.Profile{}, false, fmt.Errorf("ler perfil do ranking: %w", err)
	}
	return domain.Profile{Nickname: r.Nickname, Visible: r.Visible}, true, nil
}

func (s *Store) SaveProfile(ctx context.Context, user uuid.UUID, p domain.Profile) (domain.Profile, error) {
	r, err := s.q.SaveProfile(ctx, db.SaveProfileParams{UserID: user, Nickname: p.Nickname, Visible: p.Visible})
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == uniqueViolation {
		return domain.Profile{}, domain.ErrNicknameTaken
	}
	if err != nil {
		return domain.Profile{}, fmt.Errorf("gravar perfil do ranking: %w", err)
	}
	return domain.Profile{Nickname: r.Nickname, Visible: r.Visible}, nil
}
