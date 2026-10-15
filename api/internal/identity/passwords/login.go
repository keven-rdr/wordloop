package passwords

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/keven-rdr/wordloop/api/internal/identity/app"
)

type Login struct {
	Store     Store
	Hasher    Hasher
	Users     app.Users
	Opener    app.SessionOpener
	DummyHash string
	Now       func() time.Time
}

type LoginInput struct {
	Email     string
	Password  string
	UserAgent string
}

func (u Login) Run(ctx context.Context, in LoginInput) (app.Opened, error) {
	now := nowOf(u.Now)
	if len(in.Password) > maxPasswordBytes {
		return app.Opened{}, ErrInvalidCredentials
	}
	acc, exists, err := u.find(ctx, in.Email)
	if err != nil {
		return app.Opened{}, err
	}
	usable, err := u.reserve(ctx, acc, exists, now)
	if err != nil {
		return app.Opened{}, err
	}
	hash := u.DummyHash
	if usable {
		hash = acc.Hash
	}
	ok, err := u.Hasher.Verify(ctx, in.Password, hash)
	if err != nil {
		return app.Opened{}, fmt.Errorf("conferir a senha: %w", err)
	}
	if !usable || !ok {
		return app.Opened{}, ErrInvalidCredentials
	}
	return u.open(ctx, acc, in.UserAgent, now)
}

func (u Login) find(ctx context.Context, email string) (Account, bool, error) {
	acc, err := u.Store.FindByEmail(ctx, strings.ToLower(strings.TrimSpace(email)))
	if errors.Is(err, ErrNotFound) {
		return Account{}, false, nil
	}
	if err != nil {
		return Account{}, false, fmt.Errorf("buscar a conta: %w", err)
	}
	return acc, acc.Hash != "", nil
}

func (u Login) reserve(ctx context.Context, acc Account, exists bool, now time.Time) (bool, error) {
	if !exists || acc.LockedUntil.After(now) {
		return false, nil
	}
	allowed, err := u.Store.BeginAttempt(ctx, acc.ID, now)
	if err != nil {
		return false, fmt.Errorf("contar a tentativa: %w", err)
	}
	return allowed, nil
}

func (u Login) open(ctx context.Context, acc Account, userAgent string, now time.Time) (app.Opened, error) {
	if err := u.Store.RegisterSuccess(ctx, acc.ID); err != nil {
		return app.Opened{}, fmt.Errorf("zerar as tentativas: %w", err)
	}
	user, err := u.Users.Get(ctx, acc.ID)
	if errors.Is(err, app.ErrNotFound) {
		return app.Opened{}, ErrInvalidCredentials
	}
	if err != nil {
		return app.Opened{}, fmt.Errorf("ler o perfil: %w", err)
	}
	opened, err := u.Opener.Open(ctx, app.Authentication{User: user, MFA: false, AuthTime: now, UserAgent: userAgent})
	if err != nil {
		return app.Opened{}, fmt.Errorf("abrir a sessao: %w", err)
	}
	return opened, nil
}
