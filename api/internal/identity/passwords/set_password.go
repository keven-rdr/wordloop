package passwords

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/keven-rdr/wordloop/api/internal/identity/app"
	"github.com/keven-rdr/wordloop/api/internal/identity/domain"
	"github.com/keven-rdr/wordloop/api/internal/platform/token"
)

type SetPassword struct {
	Store  Store
	Hasher Hasher
	Users  app.Users
	Opener app.SessionOpener
	Now    func() time.Time
}

type SetInput struct {
	Token     string
	Password  string
	UserAgent string
}

func (u SetPassword) Run(ctx context.Context, in SetInput) (*app.Opened, error) {
	now := nowOf(u.Now)
	tokenHash := token.Hash(in.Token)
	owner, err := u.Store.PeekToken(ctx, tokenHash, now)
	if errors.Is(err, ErrNotFound) {
		return nil, ErrUnauthorized
	}
	if err != nil {
		return nil, fmt.Errorf("ler o token: %w", err)
	}
	if owner.Purpose == domain.TokenRecovery && owner.Role == domain.RoleAdmin {
		return nil, ErrUnauthorized
	}
	if err := CheckPolicy(in.Password, owner.Email); err != nil {
		return nil, err
	}
	hash, err := u.Hasher.Hash(ctx, in.Password)
	if err != nil {
		return nil, fmt.Errorf("gerar o hash: %w", err)
	}
	if err := u.Store.SetWithToken(ctx, tokenHash, owner.ID, hash, now); err != nil {
		return nil, err
	}
	return u.open(ctx, owner, in.UserAgent, now), nil
}

func (u SetPassword) open(ctx context.Context, owner Owner, userAgent string, now time.Time) *app.Opened {
	user, err := u.Users.Get(ctx, owner.ID)
	if err != nil {
		slog.Error("senha gravada, mas o perfil nao foi lido; a pessoa deve entrar", "user_id", owner.ID, "err", err)
		return nil
	}
	opened, err := u.Opener.Open(ctx, app.Authentication{User: user, MFA: false, AuthTime: now, UserAgent: userAgent})
	if err != nil {
		if !errors.Is(err, app.ErrMFARequired) {
			slog.Error("senha gravada, mas a sessao nao abriu; a pessoa deve entrar", "user_id", owner.ID, "err", err)
		}
		return nil
	}
	return &opened
}
