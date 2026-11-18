package app

import (
	"context"

	"github.com/keven-rdr/wordloop/api/internal/learning/domain"
)

// SettingsService le e grava as preferencias da pessoa: idioma, tema e fuso.
type SettingsService struct {
	Viewers Viewers
	Repo    Repository
}

// Get devolve as preferencias; na primeira vez cria os padroes (idioma e tema do sistema, fuso acompanhando o aparelho).
func (u SettingsService) Get(ctx context.Context) (domain.Settings, error) {
	v, err := signedIn(ctx, u.Viewers)
	if err != nil {
		return domain.Settings{}, err
	}
	var out domain.Settings
	err = u.Repo.InTx(ctx, func(tx Tx) error {
		var err error
		out, err = tx.Settings(ctx, v.UserID)
		return err
	})
	return out, err
}

// Update grava as preferencias pedidas e devolve os ajustes completos. O corte do dia e o agendador nao mudam aqui.
func (u SettingsService) Update(ctx context.Context, want domain.Settings) (domain.Settings, error) {
	v, err := signedIn(ctx, u.Viewers)
	if err != nil {
		return domain.Settings{}, err
	}
	if !want.ValidPreferences() {
		return domain.Settings{}, domain.ErrInvalidSettings
	}
	var out domain.Settings
	err = u.Repo.InTx(ctx, func(tx Tx) error {
		current, err := tx.Settings(ctx, v.UserID)
		if err != nil {
			return err
		}
		current.TZ, current.UILocale, current.Theme, current.TZAuto = want.TZ, want.UILocale, want.Theme, want.TZAuto
		out = current
		return tx.SavePreferences(ctx, v.UserID, current)
	})
	return out, err
}
