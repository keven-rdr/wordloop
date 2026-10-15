package app

import (
	"context"

	"github.com/keven-rdr/wordloop/api/internal/catalog/domain"
)

type ListUnits struct {
	Viewers Viewers
	Repo    Repository
}

func (u ListUnits) Run(ctx context.Context) ([]domain.Unit, error) {
	v, err := signedIn(ctx, u.Viewers)
	if err != nil {
		return nil, err
	}
	return u.Repo.Units(ctx, v.SeesDrafts())
}
