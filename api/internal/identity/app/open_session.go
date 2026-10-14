package app

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/keven-rdr/wordloop/api/internal/identity/domain"
)

type SessionOpener struct {
	Sessions Sessions
	Sealer   Sealer
	Rand     io.Reader
	Now      func() time.Time
	Policy   domain.SessionPolicy
}

type Authentication struct {
	User         domain.User
	RefreshToken string
	AuthTime     time.Time
	MFA          bool
	UserAgent    string
}

type Opened struct {
	SessionID string
	ExpiresAt time.Time
}

func (o SessionOpener) Open(ctx context.Context, a Authentication) (Opened, error) {
	if a.User.Role == domain.RoleAdmin && !a.MFA {
		return Opened{}, ErrMFARequired
	}
	id, err := randomToken(o.Rand, 32)
	if err != nil {
		return Opened{}, err
	}
	refreshEnc, err := o.sealRefresh(a.RefreshToken)
	if err != nil {
		return Opened{}, err
	}
	now := o.now()
	expires := now.Add(o.Policy.OrDefault().For(a.User.Role).Absolute)
	err = o.Sessions.Create(ctx, NewSession{
		IDHash: hashSession(id), UserID: a.User.ID, RefreshTokenEnc: refreshEnc, CreatedAt: now, ExpiresAt: expires,
		AuthenticatedAt: authenticatedAt(a.AuthTime, now), MFA: a.MFA, UserAgent: a.UserAgent,
	})
	if err != nil {
		return Opened{}, fmt.Errorf("gravar a sessao: %w", err)
	}
	return Opened{SessionID: id, ExpiresAt: expires}, nil
}

func (o SessionOpener) sealRefresh(refresh string) (string, error) {
	if refresh == "" {
		return "", nil
	}
	sealed, err := o.Sealer.Seal([]byte(refresh))
	if err != nil {
		return "", fmt.Errorf("selar o refresh token: %w", err)
	}
	return sealed, nil
}

func (o SessionOpener) now() time.Time {
	if o.Now != nil {
		return o.Now()
	}
	return time.Now()
}

func authenticatedAt(authTime, now time.Time) time.Time {
	if authTime.IsZero() || authTime.After(now) {
		return now
	}
	return authTime
}

func (u FinishLogin) openSession(ctx context.Context, user domain.User, tokens Tokens, userAgent, returnTo string) (LoggedIn, error) {
	opener := SessionOpener{Sessions: u.Sessions, Sealer: u.Sealer, Rand: u.Rand, Now: u.Now, Policy: u.Policy}
	opened, err := opener.Open(ctx, Authentication{
		User: user, RefreshToken: tokens.RefreshToken, AuthTime: tokens.Identity.AuthTime, MFA: tokens.Identity.MFA, UserAgent: userAgent,
	})
	if err != nil {
		return LoggedIn{}, err
	}
	return LoggedIn{SessionID: opened.SessionID, ReturnTo: returnTo, ExpiresAt: opened.ExpiresAt}, nil
}
