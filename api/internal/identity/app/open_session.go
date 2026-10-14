package app

import (
	"context"
	"fmt"
	"time"

	"github.com/keven-rdr/wordloop/api/internal/identity/domain"
)

func (u FinishLogin) openSession(ctx context.Context, user domain.User, tokens Tokens, userAgent, returnTo string) (LoggedIn, error) {
	id, err := randomToken(u.Rand, 32)
	if err != nil {
		return LoggedIn{}, err
	}
	refreshEnc, err := u.Sealer.Seal([]byte(tokens.RefreshToken))
	if err != nil {
		return LoggedIn{}, fmt.Errorf("selar o refresh token: %w", err)
	}
	now := u.now()
	expires := now.Add(u.policy().For(user.Role).Absolute)
	err = u.Sessions.Create(ctx, NewSession{
		IDHash: hashSession(id), UserID: user.ID, RefreshTokenEnc: refreshEnc, CreatedAt: now, ExpiresAt: expires,
		AuthenticatedAt: authenticatedAt(tokens.Identity, now), MFA: tokens.Identity.MFA, UserAgent: userAgent,
	})
	if err != nil {
		return LoggedIn{}, fmt.Errorf("gravar a sessao: %w", err)
	}
	return LoggedIn{SessionID: id, ReturnTo: returnTo, ExpiresAt: expires}, nil
}

func authenticatedAt(id domain.Identity, now time.Time) time.Time {
	if id.AuthTime.IsZero() || id.AuthTime.After(now) {
		return now
	}
	return id.AuthTime
}

func (u FinishLogin) policy() domain.SessionPolicy { return u.Policy.OrDefault() }

func (u FinishLogin) now() time.Time {
	if u.Now != nil {
		return u.Now()
	}
	return time.Now()
}
