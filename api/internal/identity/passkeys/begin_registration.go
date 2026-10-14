package passkeys

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/google/uuid"

	"github.com/keven-rdr/wordloop/api/internal/platform/token"
)

type BeginRegistration struct {
	Ceremony   Ceremony
	Store      Store
	Challenges Challenges
	Rand       io.Reader
	Now        func() time.Time
}

type Subject struct {
	SessionUserID uuid.UUID
	InviteToken   string
}

type Started struct {
	Options     []byte
	ChallengeID string
	ExpiresAt   time.Time
}

func (u BeginRegistration) Run(ctx context.Context, s Subject) (Started, error) {
	now := nowOf(u.Now)
	userID, inviteHash, err := u.subject(ctx, s, now)
	if err != nil {
		return Started{}, err
	}
	owner, err := u.Store.Owner(ctx, userID)
	if err != nil {
		return Started{}, fmt.Errorf("ler a conta: %w", err)
	}
	existing, err := u.Store.List(ctx, userID)
	if err != nil {
		return Started{}, fmt.Errorf("listar as passkeys: %w", err)
	}
	if len(existing) >= MaxPerAccount {
		return Started{}, ErrTooMany
	}
	options, state, err := u.Ceremony.BeginRegistration(owner, existing)
	if err != nil {
		return Started{}, fmt.Errorf("iniciar o registro: %w", err)
	}
	return saveChallenge(ctx, u.Challenges, u.Rand, now, Challenge{
		Purpose: PurposeRegister, UserID: userID, InviteHash: inviteHash, State: state,
	}, options)
}

func (u BeginRegistration) subject(ctx context.Context, s Subject, now time.Time) (uuid.UUID, []byte, error) {
	if s.SessionUserID != uuid.Nil {
		return s.SessionUserID, nil, nil
	}
	if s.InviteToken == "" {
		return uuid.Nil, nil, ErrUnauthorized
	}
	hash := token.Hash(s.InviteToken)
	id, err := u.Store.PeekInvite(ctx, hash, now)
	if err != nil {
		return uuid.Nil, nil, ErrUnauthorized
	}
	return id, hash, nil
}

func saveChallenge(ctx context.Context, store Challenges, rnd io.Reader, now time.Time, c Challenge, options []byte) (Started, error) {
	id, hash, err := token.New(rnd)
	if err != nil {
		return Started{}, err
	}
	c.IDHash, c.CreatedAt, c.ExpiresAt = hash, now, now.Add(ChallengeTTL)
	if err := store.Save(ctx, c); err != nil {
		return Started{}, fmt.Errorf("guardar o desafio: %w", err)
	}
	return Started{Options: options, ChallengeID: id, ExpiresAt: c.ExpiresAt}, nil
}

func nowOf(f func() time.Time) time.Time {
	if f != nil {
		return f()
	}
	return time.Now()
}
