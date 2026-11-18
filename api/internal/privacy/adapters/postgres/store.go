// Package postgres implementa a exportacao e a exclusao de dados com pgx e sqlc. O codigo de db/ vem de queries/*.sql
// e e gerado junto com o dos outros modulos (go generate de identity/adapters/postgres).
package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/keven-rdr/wordloop/api/internal/privacy/adapters/postgres/db"
	"github.com/keven-rdr/wordloop/api/internal/privacy/app"
	"github.com/keven-rdr/wordloop/api/internal/privacy/domain"
)

// Store implementa app.Store. Le e apaga nas tabelas de varios modulos: o direito da pessoa sobre os dados atravessa
// o sistema inteiro, por isso este modulo tem consultas proprias.
type Store struct{ pool *pgxpool.Pool }

var _ app.Store = (*Store)(nil)

func NewStore(pool *pgxpool.Pool) *Store { return &Store{pool: pool} }

func (s *Store) Email(ctx context.Context, user uuid.UUID) (string, error) {
	email, err := db.New(s.pool).AccountEmail(ctx, user)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", domain.ErrNotFound
	}
	if err != nil {
		return "", fmt.Errorf("ler e-mail da conta: %w", err)
	}
	return email, nil
}

// Export le tudo numa consulta so (uma instantanea consistente), com as colunas sempre listadas.
func (s *Store) Export(ctx context.Context, user uuid.UUID) (domain.Export, error) {
	data, err := db.New(s.pool).ExportData(ctx, user)
	if err != nil {
		return domain.Export{}, fmt.Errorf("exportar dados: %w", err)
	}
	return domain.Export{Data: data}, nil
}

// Erase apaga o log (com a excecao estreita do gatilho, so para esta pessoa e so nesta transacao), as correcoes
// dele, a conta (o resto sai em cascata) e registra a exclusao sem dado pessoal. Tudo ou nada.
func (s *Store) Erase(ctx context.Context, user uuid.UUID, _ time.Time) (domain.ErasureReport, error) {
	var report domain.ErasureReport
	err := pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		q := db.New(tx)
		if _, err := q.SetErasingUser(ctx, user.String()); err != nil {
			return err
		}
		if err := q.EraseReviewVoids(ctx, user); err != nil {
			return err
		}
		reviews, err := q.EraseReviews(ctx, user)
		if err != nil {
			return err
		}
		gone, err := q.EraseAccount(ctx, user)
		if err != nil {
			return err
		}
		if gone == 0 {
			return domain.ErrNotFound
		}
		report.Reviews = int(reviews)
		return q.RecordErasure(ctx, int32(reviews)) //nolint:gosec // contagem de respostas de uma pessoa
	})
	if errors.Is(err, domain.ErrNotFound) {
		return domain.ErasureReport{}, err
	}
	if err != nil {
		return domain.ErasureReport{}, fmt.Errorf("apagar conta: %w", err)
	}
	return report, nil
}
