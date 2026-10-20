// Package app tem os casos de uso do aprendizado: montar a sessao (plano) e receber as respostas (revisoes).
package app

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/keven-rdr/wordloop/api/internal/learning/domain"
	"github.com/keven-rdr/wordloop/api/internal/srs"
)

var (
	ErrNoSession    = errors.New("sem sessao valida")
	ErrInvalidBatch = errors.New("lote invalido")
)

// MaxBatch e o maximo de respostas por envio.
const MaxBatch = 200

// Repository abre transacoes; cada caso de uso (ou cada resposta) roda inteiro dentro de uma.
// Declarada aqui, pelo consumidor.
type Repository interface {
	InTx(ctx context.Context, fn func(Tx) error) error
}

// Tx e o que os casos de uso precisam do banco dentro de uma transacao.
type Tx interface {
	// Settings devolve os ajustes da pessoa, criando os padroes na primeira vez.
	Settings(ctx context.Context, user uuid.UUID) (domain.Settings, error)
	SaveTZ(ctx context.Context, user uuid.UUID, tz string) error

	DueCards(ctx context.Context, user uuid.UUID, now time.Time, limit int) ([]domain.Card, error)
	// NewItems lista os proximos itens sem cartao receptivo, na ordem das unidades.
	NewItems(ctx context.Context, user uuid.UUID, drafts bool, limit int) ([]uuid.UUID, error)
	// ProductiveItems lista os itens cujo cartao receptivo ja passou de S minimo (ha um dia ou mais) e que ainda nao tem o produtivo.
	ProductiveItems(ctx context.Context, user uuid.UUID, minStability float64, before time.Time, limit int) ([]uuid.UUID, error)
	CreateCard(ctx context.Context, user uuid.UUID, item uuid.UUID, dir domain.Direction, now time.Time) (domain.Card, error)
	Contents(ctx context.Context, items []uuid.UUID, locale string, drafts bool) (map[uuid.UUID]domain.Content, error)
	Distractors(ctx context.Context, item uuid.UUID, locale string, drafts bool, n int) ([]domain.Option, error)
	CreateSession(ctx context.Context, id, user uuid.UUID, mode domain.Mode, minutes int, now, expires time.Time) error

	// LockCard trava o cartao da pessoa para atualizar (false se nao existe ou e de outra pessoa).
	LockCard(ctx context.Context, user, card uuid.UUID) (domain.Card, bool, error)
	ExpectedTerm(ctx context.Context, item uuid.UUID) (string, error)
	// InsertReview grava a linha do log; false se o UUID ja existia (reenvio).
	InsertReview(ctx context.Context, row domain.LogRow) (bool, error)
	SaveCard(ctx context.Context, card domain.Card) error
	CardLog(ctx context.Context, card uuid.UUID) ([]srs.Answer, error)
}

// Viewers descobre quem estuda a partir da sessao; sem sessao devolve ErrNoSession.
type Viewers interface {
	Viewer(ctx context.Context) (domain.Viewer, error)
}

// Clock e a fonte de tempo (trocavel nos testes).
type Clock func() time.Time

func (c Clock) now() time.Time {
	if c == nil {
		return time.Now().UTC()
	}
	return c().UTC()
}

func signedIn(ctx context.Context, viewers Viewers) (domain.Viewer, error) {
	v, err := viewers.Viewer(ctx)
	if err != nil {
		return domain.Viewer{}, err
	}
	if v.UserID == uuid.Nil {
		return domain.Viewer{}, ErrNoSession
	}
	return v, nil
}
