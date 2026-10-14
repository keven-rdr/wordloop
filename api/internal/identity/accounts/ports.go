package accounts

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/keven-rdr/wordloop/api/internal/identity/domain"
)

var (
	ErrNotFound      = errors.New("conta nao encontrada")
	ErrEmailTaken    = errors.New("e-mail ja cadastrado")
	ErrInvalidEmail  = errors.New("e-mail invalido")
	ErrAccountExists = errors.New("ja existe uma conta com esse e-mail que nao pode receber convite")
	ErrInvalidURL    = errors.New("APP_PUBLIC_URL invalida: use http(s)://host")
	ErrForbidden     = errors.New("somente administradores podem mudar esta opcao")
	ErrWeakSession   = errors.New("ligar a verificacao em duas etapas exige uma sessao que ja a comprovou")
)

type Account struct {
	ID             uuid.UUID
	Role           domain.Role
	HasCredentials bool
}

type MFAStore interface {
	SetMFARequired(ctx context.Context, userID uuid.UUID, required bool) error
}

type Store interface {
	MFAStore
	FindByEmail(ctx context.Context, email string) (Account, error)
	CreateAdmin(ctx context.Context, email, name string, verifiedAt time.Time) (uuid.UUID, error)
	ReplaceInvite(ctx context.Context, userID uuid.UUID, idHash []byte, createdAt, expiresAt time.Time) error
}
