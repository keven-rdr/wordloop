package passwords

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/keven-rdr/wordloop/api/internal/identity/domain"
)

var (
	ErrInvalidCredentials = errors.New("e-mail ou senha incorretos")
	ErrWeak               = errors.New("senha fraca")
	ErrUnauthorized       = errors.New("convite invalido ou vencido")
	ErrNotFound           = errors.New("conta nao encontrada")
)

const (
	MinLength        = 12
	MaxLength        = 128
	maxPasswordBytes = 1024
)

type Hasher interface {
	Hash(ctx context.Context, plain string) (string, error)
	Verify(ctx context.Context, plain, encoded string) (bool, error)
}

type Account struct {
	ID          uuid.UUID
	Hash        string
	LockedUntil time.Time
}

type Owner struct {
	ID      uuid.UUID
	Email   string
	Role    domain.Role
	Purpose domain.TokenPurpose
}

type Store interface {
	FindByEmail(ctx context.Context, email string) (Account, error)
	BeginAttempt(ctx context.Context, id uuid.UUID, at time.Time) (allowed bool, err error)
	RegisterSuccess(ctx context.Context, id uuid.UUID) error
	PeekToken(ctx context.Context, tokenHash []byte, now time.Time) (Owner, error)
	SetWithToken(ctx context.Context, tokenHash []byte, userID uuid.UUID, hash string, now time.Time) error
}

func nowOf(f func() time.Time) time.Time {
	if f != nil {
		return f()
	}
	return time.Now()
}
