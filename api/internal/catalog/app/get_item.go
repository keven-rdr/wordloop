package app

import (
	"context"

	"github.com/google/uuid"

	"github.com/keven-rdr/wordloop/api/internal/catalog/domain"
)

type GetItem struct {
	Viewers Viewers
	Repo    Repository
}

func (u GetItem) Run(ctx context.Context, id uuid.UUID, locale string) (domain.Detail, error) {
	v, err := signedIn(ctx, u.Viewers)
	if err != nil {
		return domain.Detail{}, err
	}
	if locale == "" {
		locale = DefaultLang
	}
	if !ValidLocale(locale) {
		return domain.Detail{}, ErrInvalidQuery
	}
	return u.Repo.Item(ctx, id, locale, v.SeesDrafts())
}
