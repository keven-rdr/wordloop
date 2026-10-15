package registration

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/keven-rdr/wordloop/api/internal/identity/domain"
)

var (
	ErrInvalidEmail = errors.New("e-mail invalido")
	ErrNotFound     = errors.New("conta nao encontrada")
	ErrEmailTaken   = errors.New("e-mail ja cadastrado")
)

const (
	VerifyTTL     = 24 * time.Hour
	RecoveryTTL   = 30 * time.Minute
	DefaultLocale = "pt-BR"
	cooldown      = time.Minute
	hourlyLimit   = 5
	sendTimeout   = 15 * time.Second
)

type Account struct {
	ID       uuid.UUID
	Role     domain.Role
	Verified bool
}

type Store interface {
	FindByEmail(ctx context.Context, email string) (Account, error)
	CreatePending(ctx context.Context, email, displayName string) (uuid.UUID, error)
	RecentTokens(ctx context.Context, userID uuid.UUID, since time.Time) (int, error)
	ReplaceToken(ctx context.Context, userID uuid.UUID, purpose domain.TokenPurpose, hash []byte, createdAt, expiresAt time.Time) error
}

type Input struct {
	Email  string
	Locale string
}
