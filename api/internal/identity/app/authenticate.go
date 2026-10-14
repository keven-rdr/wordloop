package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/keven-rdr/wordloop/api/internal/identity/domain"
)

const touchEvery = time.Minute

// Authenticate descobre quem esta fazendo o pedido a partir do id da sessao (valor do cookie).
// O papel e lido do banco a cada chamada e define os prazos: nada vem do corpo do pedido.
type Authenticate struct {
	Sessions Sessions
	Users    Users
	Policy   domain.SessionPolicy // campo zero = padrao (OrDefault)
	Now      func() time.Time     // injetado nos testes
}

// Run devolve ErrNoSession quando nao ha sessao valida (ausente, inexistente, expirada, ociosa ou de usuario removido).
func (u Authenticate) Run(ctx context.Context, sessionID string) (domain.Principal, error) {
	if sessionID == "" {
		return domain.Principal{}, ErrNoSession
	}
	hash := hashSession(sessionID)
	rec, err := u.Sessions.Get(ctx, hash)
	if errors.Is(err, ErrNotFound) {
		return domain.Principal{}, ErrNoSession
	}
	if err != nil {
		return domain.Principal{}, fmt.Errorf("ler a sessao: %w", err)
	}
	user, err := u.Users.Get(ctx, rec.UserID)
	if err != nil && !errors.Is(err, ErrNotFound) {
		return domain.Principal{}, fmt.Errorf("ler o perfil: %w", err)
	}
	now := u.now()
	if errors.Is(err, ErrNotFound) || !u.alive(rec, user.Role, now) {
		_ = u.Sessions.Delete(ctx, hash)
		return domain.Principal{}, ErrNoSession
	}
	if now.Sub(rec.LastUsedAt) >= touchEvery {
		if err := u.Sessions.Touch(ctx, hash, now); err != nil {
			slog.Warn("nao foi possivel atualizar o ultimo uso da sessao", "err", err)
		}
	}
	return domain.Principal{User: user, SessionID: rec.PublicID, AuthenticatedAt: rec.AuthenticatedAt, MFA: rec.MFA}, nil
}

// alive aplica o teto gravado na sessao e os prazos do papel ATUAL (quem virou admin passa a ter o prazo curto na hora).
func (u Authenticate) alive(rec SessionRecord, role domain.Role, now time.Time) bool {
	if !now.Before(rec.ExpiresAt) {
		return false
	}
	return !u.policy().For(role).Expired(rec.CreatedAt, rec.LastUsedAt, now)
}

func (u Authenticate) policy() domain.SessionPolicy { return u.Policy.OrDefault() }

func (u Authenticate) now() time.Time {
	if u.Now != nil {
		return u.Now()
	}
	return time.Now()
}
