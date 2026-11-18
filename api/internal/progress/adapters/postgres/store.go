// Package postgres implementa a leitura do progresso com pgx e sqlc. O codigo de db/ vem de queries/*.sql e e
// gerado junto com o dos outros modulos (go generate de identity/adapters/postgres).
package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/keven-rdr/wordloop/api/internal/progress/adapters/postgres/db"
	"github.com/keven-rdr/wordloop/api/internal/progress/app"
	"github.com/keven-rdr/wordloop/api/internal/progress/domain"
)

// Store implementa app.Reader (consultas de leitura) e app.Goals (abrir meta usa uma transacao).
type Store struct {
	q    *db.Queries
	pool *pgxpool.Pool
}

var _ app.Reader = (*Store)(nil)

func NewStore(pool *pgxpool.Pool) *Store { return &Store{q: db.New(pool), pool: pool} }

func date(t time.Time) pgtype.Date { return pgtype.Date{Time: t, Valid: true} }

func (s *Store) Settings(ctx context.Context, user uuid.UUID) (domain.Settings, error) {
	row, err := s.q.GetSettings(ctx, user)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.DefaultSettings, nil
	}
	if err != nil {
		return domain.Settings{}, fmt.Errorf("ler ajustes: %w", err)
	}
	return domain.Settings{TZ: row.Tz, CutoffHour: int(row.DayCutoffHour)}, nil
}

func (s *Store) Days(ctx context.Context, user uuid.UUID, from, to time.Time) ([]domain.Day, error) {
	rows, err := s.q.StudyDays(ctx, db.StudyDaysParams{UserID: user, FromDate: date(from), ToDate: date(to)})
	if err != nil {
		return nil, fmt.Errorf("listar dias de estudo: %w", err)
	}
	out := make([]domain.Day, len(rows))
	for i, r := range rows {
		out[i] = domain.Day{
			Date: r.LocalDate.Time, Reviews: int(r.Reviews), Correct: int(r.Correct), NewItems: int(r.NewItems),
			ActiveMS: r.ActiveMs, KnownWords: int(r.KnownLexemes),
		}
	}
	return out, nil
}

func (s *Store) Forecast(ctx context.Context, user uuid.UUID, st domain.Settings, today time.Time, days int) (map[int]int, error) {
	rows, err := s.q.ForecastCounts(ctx, db.ForecastCountsParams{
		UserID: user, Tz: st.TZ, CutoffHour: int32(st.CutoffHour), Today: date(today), LastOffset: int32(days - 1), //nolint:gosec // 0 a 8 e 6
	})
	if err != nil {
		return nil, fmt.Errorf("prever revisoes: %w", err)
	}
	out := make(map[int]int, len(rows))
	for _, r := range rows {
		out[int(r.DayOffset)] = int(r.Due)
	}
	return out, nil
}

func (s *Store) KnownWords(ctx context.Context, user uuid.UUID) (int, error) {
	n, err := s.q.KnownWords(ctx, user)
	if err != nil {
		return 0, fmt.Errorf("contar palavras conhecidas: %w", err)
	}
	return int(n), nil
}

func (s *Store) MostMissed(ctx context.Context, user uuid.UUID, since time.Time, locale string, limit int) ([]domain.Missed, error) {
	rows, err := s.q.MostMissed(ctx, db.MostMissedParams{
		UserID: user, Since: date(since), Locale: locale, MaxRows: int32(limit), //nolint:gosec // 5
	})
	if err != nil {
		return nil, fmt.Errorf("listar palavras mais erradas: %w", err)
	}
	out := make([]domain.Missed, len(rows))
	for i, r := range rows {
		out[i] = domain.Missed{ItemID: r.ItemID, Term: r.Term, Gloss: r.Gloss, Misses: int(r.Misses)}
	}
	return out, nil
}
