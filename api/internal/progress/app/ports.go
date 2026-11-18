// Package app tem os casos de uso do progresso: o painel e o heatmap. So leitura.
package app

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/keven-rdr/wordloop/api/internal/progress/domain"
)

var (
	ErrNoSession    = errors.New("sem sessao valida")
	ErrInvalidRange = errors.New("intervalo invalido")
)

// Reader e o que os casos de uso precisam do banco. Declarada aqui, pelo consumidor.
type Reader interface {
	// Settings devolve fuso e corte da pessoa; sem ajustes salvos (ainda nao estudou), os padroes.
	Settings(ctx context.Context, user uuid.UUID) (domain.Settings, error)
	// Days lista os dias com estudo entre from e to (inclusive), do mais antigo ao mais novo.
	Days(ctx context.Context, user uuid.UUID, from, to time.Time) ([]domain.Day, error)
	// Forecast conta os cartoes que vencem por posicao a partir de hoje (0 = hoje e atrasadas), ate days-1.
	Forecast(ctx context.Context, user uuid.UUID, s domain.Settings, today time.Time, days int) (map[int]int, error)
	KnownWords(ctx context.Context, user uuid.UUID) (int, error)
	// MostMissed lista as palavras mais erradas desde a data, com a traducao no idioma pedido.
	MostMissed(ctx context.Context, user uuid.UUID, since time.Time, locale string, limit int) ([]domain.Missed, error)
}

// Viewers descobre quem pergunta a partir da sessao; sem sessao devolve ErrNoSession.
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
