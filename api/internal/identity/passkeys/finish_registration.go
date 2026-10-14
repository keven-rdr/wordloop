package passkeys

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/keven-rdr/wordloop/api/internal/identity/app"
	"github.com/keven-rdr/wordloop/api/internal/platform/token"
)

type FinishRegistration struct {
	Ceremony   Ceremony
	Store      Store
	Challenges Challenges
	Users      app.Users
	Opener     app.SessionOpener
	Now        func() time.Time
}

type FinishRegistrationInput struct {
	ChallengeID string
	Credential  []byte
	Name        string
	UserAgent   string
}

type Registered struct {
	Credential Credential
	Session    *app.Opened
}

func (u FinishRegistration) Run(ctx context.Context, in FinishRegistrationInput) (Registered, error) {
	now := nowOf(u.Now)
	ch, err := u.Challenges.Take(ctx, token.Hash(in.ChallengeID), PurposeRegister, now)
	if errors.Is(err, ErrNotFound) {
		return Registered{}, ErrInvalid
	}
	if err != nil {
		return Registered{}, fmt.Errorf("ler o desafio: %w", err)
	}
	owner, err := u.Store.Owner(ctx, ch.UserID)
	if err != nil {
		return Registered{}, fmt.Errorf("ler a conta: %w", err)
	}
	cred, err := u.Ceremony.FinishRegistration(owner, ch.State, in.Credential)
	if err != nil || !cred.UserVerified {
		return Registered{}, ErrInvalid
	}
	cred.ID, cred.UserID, cred.CreatedAt, cred.Name = uuid.New(), owner.ID, now, cleanName(in.Name)
	if len(ch.InviteHash) == 0 {
		if err := u.Store.Add(ctx, cred); err != nil {
			return Registered{}, err
		}
		return Registered{Credential: cred}, nil
	}
	if err := u.Store.AddWithInvite(ctx, cred, ch.InviteHash, now); err != nil {
		return Registered{}, err
	}
	opened, err := u.open(ctx, owner.ID, in.UserAgent)
	if err != nil {
		slog.Error("passkey registrada, mas a sessao nao abriu; a pessoa deve entrar com a passkey", "user_id", owner.ID, "err", err)
		return Registered{Credential: cred}, nil
	}
	return Registered{Credential: cred, Session: &opened}, nil
}

func (u FinishRegistration) open(ctx context.Context, userID uuid.UUID, userAgent string) (app.Opened, error) {
	user, err := u.Users.Get(ctx, userID)
	if err != nil {
		return app.Opened{}, fmt.Errorf("ler o perfil: %w", err)
	}
	opened, err := u.Opener.Open(ctx, app.Authentication{User: user, MFA: true, AuthTime: nowOf(u.Now), UserAgent: userAgent})
	if err != nil {
		return app.Opened{}, fmt.Errorf("abrir a sessao: %w", err)
	}
	return opened, nil
}

func cleanName(s string) string {
	s = strings.TrimSpace(strings.Map(func(r rune) rune {
		if r < 0x20 || r == 0x7f {
			return -1
		}
		return r
	}, s))
	if s == "" {
		return DefaultName
	}
	if r := []rune(s); len(r) > MaxNameLength {
		s = string(r[:MaxNameLength])
	}
	return strings.TrimSpace(s)
}
