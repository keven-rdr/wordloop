package app

import (
	"context"
	"fmt"
)

// RevokeAllSessions encerra todas as sessoes de quem esta logado, inclusive a atual ("sair de todos os aparelhos").
type RevokeAllSessions struct {
	Auth     Authenticate
	Sessions Sessions
}

// Run devolve ErrNoSession se quem chamou nao esta logado.
func (u RevokeAllSessions) Run(ctx context.Context, sessionID string) error {
	who, err := u.Auth.Run(ctx, sessionID)
	if err != nil {
		return err
	}
	if err := u.Sessions.DeleteAllByUser(ctx, who.User.ID); err != nil {
		return fmt.Errorf("encerrar todas as sessoes: %w", err)
	}
	return nil
}
