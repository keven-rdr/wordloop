package app

import (
	"context"

	"github.com/google/uuid"

	"github.com/keven-rdr/wordloop/api/internal/learning/domain"
	"github.com/keven-rdr/wordloop/api/internal/srs"
)

// SchedulerVersion identifica a versao das regras gravada em cada linha do log.
const SchedulerVersion = "1"

// SubmitReviews recebe um lote de respostas: corrige, da a nota, atualiza o cartao e grava no log imutavel.
// Cada resposta roda na propria transacao, entao uma rejeitada nao derruba as outras; reenviar o mesmo UUID nao muda nada.
type SubmitReviews struct {
	Viewers   Viewers
	Repo      Repository
	Scheduler srs.Scheduler
	Now       Clock
}

func (u SubmitReviews) Run(ctx context.Context, answers []domain.Answer) ([]domain.Result, error) {
	v, err := signedIn(ctx, u.Viewers)
	if err != nil {
		return nil, err
	}
	if len(answers) == 0 || len(answers) > MaxBatch {
		return nil, ErrInvalidBatch
	}
	results := make([]domain.Result, 0, len(answers))
	for _, a := range answers {
		var res domain.Result
		err := u.Repo.InTx(ctx, func(tx Tx) error {
			var err error
			res, err = u.apply(ctx, tx, v, a)
			return err
		})
		if err != nil {
			return nil, err
		}
		results = append(results, res)
	}
	return results, nil
}

func rejected(id uuid.UUID, why domain.Reason) domain.Result {
	return domain.Result{ID: id, Status: domain.Rejected, Reason: why}
}
