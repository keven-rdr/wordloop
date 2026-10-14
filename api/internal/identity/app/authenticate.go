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

type Authenticate struct {
	Sessions Sessions
	Users    Users
	Policy   domain.SessionPolicy
	Now      func() time.Time
}

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
