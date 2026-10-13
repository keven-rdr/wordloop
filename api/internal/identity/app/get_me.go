package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/keven-rdr/wordloop/api/internal/identity/domain"
)

// GetMe devolve quem esta logado a partir do id da sessao (valor do cookie).
type GetMe struct {
	Sessions Sessions
	Users    Users
	Now      func() time.Time // injetado nos testes
}

// Run devolve ErrNoSession quando nao ha sessao valida.
func (u GetMe) Run(ctx context.Context, sessionID string) (domain.User, error) {
	if sessionID == "" {
		return domain.User{}, ErrNoSession
	}
	hash := hashSession(sessionID)
	rec, err := u.Sessions.Get(ctx, hash)
	if errors.Is(err, ErrNotFound) {
		return domain.User{}, ErrNoSession
	}
	if err != nil {
		return domain.User{}, fmt.Errorf("ler a sessao: %w", err)
	}
	now := time.Now
	if u.Now != nil {
		now = u.Now
	}
	if !rec.ExpiresAt.After(now()) {
		_ = u.Sessions.Delete(ctx, hash)
		return domain.User{}, ErrNoSession
	}
	user, err := u.Users.Get(ctx, rec.UserID)
	if errors.Is(err, ErrNotFound) {
		return domain.User{}, ErrNoSession
	}
	if err != nil {
		return domain.User{}, fmt.Errorf("ler o perfil: %w", err)
	}
	if err := u.Sessions.Touch(ctx, hash); err != nil {
		slog.Warn("nao foi possivel atualizar o ultimo uso da sessao", "err", err)
	}
	return user, nil
}
