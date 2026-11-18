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
	ErrNoSession    = errors.New("progresso: pedido sem sessao")
	ErrInvalidRange = errors.New("progresso: intervalo fora do permitido")
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

// Identity diz quem esta perguntando, pela sessao; sem sessao devolve ErrNoSession.
type Identity interface {
	Who(ctx context.Context) (domain.Viewer, error)
}

// Clock e a fonte de tempo dos casos de uso (trocavel nos testes); nulo usa o relogio do sistema.
type Clock func() time.Time

func (c Clock) now() time.Time {
	read := time.Now
	if c != nil {
		read = c
	}
	return read().UTC()
}

// asker resolve quem pergunta; uma identidade sem usuario tambem e "sem sessao".
func asker(ctx context.Context, id Identity) (domain.Viewer, error) {
	who, err := id.Who(ctx)
	if err == nil && who.UserID == uuid.Nil {
		err = ErrNoSession
	}
	return who, err
}
