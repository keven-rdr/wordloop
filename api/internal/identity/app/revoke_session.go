package app

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

// RevokeSession encerra UMA sessao de quem esta logado (por exemplo, o celular perdido).
type RevokeSession struct {
	Auth     Authenticate
	Sessions Sessions
}

// Run devolve ErrNoSession se quem chamou nao esta logado e ErrNotFound se a sessao nao existe OU e de outra pessoa
// (as duas respostas sao iguais de proposito: nao revelar que o id existe). current diz se a sessao encerrada era a atual.
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
