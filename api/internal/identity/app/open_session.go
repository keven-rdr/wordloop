package app

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

func (u FinishLogin) openSession(ctx context.Context, userID uuid.UUID, refreshToken, userAgent, returnTo string) (LoggedIn, error) {
	id, err := randomToken(u.Rand, 32)
	if err != nil {
		return LoggedIn{}, err
	}
	refreshEnc, err := u.Sealer.Seal([]byte(refreshToken))
	if err != nil {
		return LoggedIn{}, fmt.Errorf("selar o refresh token: %w", err)
	}
	ttl := u.SessionTTL
	if ttl == 0 {
		ttl = DefaultSessionTTL
	}
	expires := u.now().Add(ttl)
	err = u.Sessions.Create(ctx, NewSession{
		IDHash: hashSession(id), UserID: userID, RefreshTokenEnc: refreshEnc, ExpiresAt: expires, UserAgent: userAgent,
	})
	if err != nil {
		return LoggedIn{}, fmt.Errorf("gravar a sessao: %w", err)
	}
	return LoggedIn{SessionID: id, ReturnTo: returnTo, ExpiresAt: expires}, nil
}

func (u FinishLogin) now() time.Time {
	if u.Now != nil {
		return u.Now()
	}
	return time.Now()
}
