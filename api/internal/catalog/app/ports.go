// Package app tem os casos de uso do catalogo (consulta, so leitura).
package app

import (
	"context"
	"errors"

	"github.com/google/uuid"

	"github.com/keven-rdr/wordloop/api/internal/catalog/domain"
)

var (
	ErrNoSession    = errors.New("sem sessao valida")
	ErrNotFound     = errors.New("item nao encontrado")
	ErrInvalidQuery = errors.New("consulta invalida")
)

// Repository le o conteudo. Declarada aqui, pelo consumidor.
type Repository interface {
	Units(ctx context.Context, drafts bool) ([]domain.Unit, error)
	Items(ctx context.Context, q domain.Query) (domain.Page, error)
	Item(ctx context.Context, id uuid.UUID, locale string, drafts bool) (domain.Detail, error)
	Credits(ctx context.Context) ([]domain.Credit, error)
}

// Viewers descobre quem esta perguntando a partir da sessao; sem sessao devolve o Viewer vazio.
type Viewers interface {
	Viewer(ctx context.Context) (domain.Viewer, error)
}

func signedIn(ctx context.Context, viewers Viewers) (domain.Viewer, error) {
	v, err := viewers.Viewer(ctx)
	if err != nil {
		return domain.Viewer{}, err
	}
	if !v.Authenticated {
		return domain.Viewer{}, ErrNoSession
	}
	return v, nil
}
