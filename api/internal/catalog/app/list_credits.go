package app

import (
	"context"

	"github.com/keven-rdr/wordloop/api/internal/catalog/domain"
)

// ListCredits e publico: a licenca exige mostrar o credito a quem usa o app, com ou sem conta.
type ListCredits struct{ Repo Repository }

func (u ListCredits) Run(ctx context.Context) ([]domain.Credit, error) {
	return u.Repo.Credits(ctx)
}
