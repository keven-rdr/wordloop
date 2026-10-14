package passkeys

import (
	"context"
	"fmt"
	"io"
	"time"
)

type BeginLogin struct {
	Ceremony   Ceremony
	Challenges Challenges
	Rand       io.Reader
	Now        func() time.Time
}

func (u BeginLogin) Run(ctx context.Context) (Started, error) {
	options, state, err := u.Ceremony.BeginLogin()
	if err != nil {
		return Started{}, fmt.Errorf("iniciar o login: %w", err)
	}
	return saveChallenge(ctx, u.Challenges, u.Rand, nowOf(u.Now), Challenge{Purpose: PurposeLogin, State: state}, options)
}
