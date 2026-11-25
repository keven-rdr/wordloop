package app

import (
	"context"

	"github.com/keven-rdr/wordloop/api/internal/leaderboard/domain"
)

// Profile e como quem pergunta aparece no ranking; sem perfil, apelido vazio e fora do ranking.
func (s Service) Profile(ctx context.Context) (domain.Profile, error) {
	v, err := s.who(ctx)
	if err != nil {
		return domain.Profile{}, err
	}
	p, _, err := s.Store.Profile(ctx, v.UserID)
	return p, err
}

// SaveProfile grava o apelido e se a pessoa participa. O apelido e validado (domain.CheckNickname) e e unico.
func (s Service) SaveProfile(ctx context.Context, p domain.Profile) (domain.Profile, error) {
	v, err := s.who(ctx)
	if err != nil {
		return domain.Profile{}, err
	}
	if err := domain.CheckNickname(p.Nickname); err != nil {
		return domain.Profile{}, err
	}
	return s.Store.SaveProfile(ctx, v.UserID, p)
}
