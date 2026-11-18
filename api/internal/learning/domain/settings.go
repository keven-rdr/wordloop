package domain

import (
	"errors"
	"slices"
)

// ErrInvalidSettings e uma preferencia fora das opcoes (fuso desconhecido, idioma ou tema inexistente).
var ErrInvalidSettings = errors.New("preferencias invalidas")

// Opcoes das preferencias de tela; "system" segue o aparelho.
var (
	uiLocales = []string{"system", "pt-BR", "en-US"}
	themes    = []string{"system", "light", "dark"}
)

// ValidPreferences diz se as preferencias de tela e o fuso estao dentro das opcoes.
func (s Settings) ValidPreferences() bool {
	return slices.Contains(uiLocales, s.UILocale) && slices.Contains(themes, s.Theme) && ValidTZ(s.TZ)
}
