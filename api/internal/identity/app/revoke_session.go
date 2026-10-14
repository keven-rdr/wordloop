package app

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

type RevokeSession struct {
	Auth     Authenticate
	Sessions Sessions
}

func (u RevokeSession) Run(ctx context.Context, sessionID string, target uuid.UUID) (current bool, err error) {
	who, err := u.Auth.Run(ctx, sessionID)
	if err != nil {
		return false, err
	}
	if err := u.Sessions.DeleteByPublicID(ctx, who.User.ID, target); err != nil {
		return false, fmt.Errorf("encerrar a sessao: %w", err)
	}
	return target == who.SessionID, nil
}
