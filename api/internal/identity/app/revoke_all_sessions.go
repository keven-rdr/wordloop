package app

import (
	"context"
	"fmt"
)

type RevokeAllSessions struct {
	Auth     Authenticate
	Sessions Sessions
}

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
