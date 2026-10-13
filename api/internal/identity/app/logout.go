package app

import (
	"context"
	"fmt"
)

// Logout encerra a sessao. Sem sessao (cookie ausente) nao e erro: sair duas vezes e inofensivo.
type Logout struct{ Sessions Sessions }

// Run apaga a sessao do banco.
func (u Logout) Run(ctx context.Context, sessionID string) error {
	if sessionID == "" {
		return nil
	}
	if err := u.Sessions.Delete(ctx, hashSession(sessionID)); err != nil {
		return fmt.Errorf("apagar a sessao: %w", err)
	}
	return nil
}
