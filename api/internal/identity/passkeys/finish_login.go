package passkeys

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/keven-rdr/wordloop/api/internal/identity/app"
	"github.com/keven-rdr/wordloop/api/internal/platform/token"
)

type FinishLogin struct {
	Ceremony   Ceremony
	Store      Store
	Challenges Challenges
	Users      app.Users
	Opener     app.SessionOpener
	Now        func() time.Time
}

type FinishLoginInput struct {
	ChallengeID string
	Credential  []byte
	UserAgent   string
}

func (u FinishLogin) Run(ctx context.Context, in FinishLoginInput) (app.Opened, error) {
	now := nowOf(u.Now)
	ch, err := u.Challenges.Take(ctx, token.Hash(in.ChallengeID), PurposeLogin, now)
	if errors.Is(err, ErrNotFound) {
		return app.Opened{}, ErrUnauthorized
	}
	if err != nil {
		return app.Opened{}, fmt.Errorf("ler o desafio: %w", err)
	}
	lookup := func(id []byte) (Credential, Owner, error) { return u.Store.Find(ctx, id) }
	cred, owner, err := u.Ceremony.FinishLogin(ch.State, in.Credential, lookup)
	if err != nil || !cred.UserVerified {
		return app.Opened{}, ErrUnauthorized
	}
	if cred.CloneWarning {
		slog.Warn("passkey com contador regressivo: possivel clone; login recusado", "user_id", owner.ID, "passkey_id", cred.ID)
		return app.Opened{}, ErrUnauthorized
	}
	if err := u.Store.RecordLogin(ctx, cred, now); err != nil {
		return app.Opened{}, fmt.Errorf("registrar o uso da passkey: %w", err)
	}
	user, err := u.Users.Get(ctx, owner.ID)
	if errors.Is(err, app.ErrNotFound) {
		return app.Opened{}, ErrUnauthorized
	}
	if err != nil {
		return app.Opened{}, fmt.Errorf("ler o perfil: %w", err)
	}
	opened, err := u.Opener.Open(ctx, app.Authentication{User: user, MFA: true, AuthTime: now, UserAgent: in.UserAgent})
	if err != nil {
		return app.Opened{}, fmt.Errorf("abrir a sessao: %w", err)
	}
	return opened, nil
}
