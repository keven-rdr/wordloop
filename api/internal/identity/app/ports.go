// Package app tem os casos de uso do modulo identity (um por arquivo). Depende so do dominio e das interfaces daqui.
package app

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/keven-rdr/wordloop/api/internal/identity/domain"
)

// Tokens e o resultado de trocar o code no provedor: quem e a pessoa e o refresh token (para renovar depois).
type Tokens struct {
	Identity     domain.Identity
	RefreshToken string
}

// Provider e o provedor OIDC (Keycloak em dev, IdP externo em tst). A troca ja valida o ID token e o nonce.
type Provider interface {
	AuthURL(ctx context.Context, state, nonce, verifier string) (string, error)
	Exchange(ctx context.Context, code, verifier, nonce string) (Tokens, error)
}

// Users guarda o perfil do app.
type Users interface {
	Upsert(ctx context.Context, id domain.Identity) (domain.User, error)
	Get(ctx context.Context, id uuid.UUID) (domain.User, error) // ErrNotFound se nao existe
}

// NewSession e a sessao a gravar. So o HASH do id vai ao banco, nunca o id.
type NewSession struct {
	IDHash          []byte
	UserID          uuid.UUID
	RefreshTokenEnc string
	CreatedAt       time.Time
	ExpiresAt       time.Time
	AuthenticatedAt time.Time
	MFA             bool
	UserAgent       string
}

type SessionRecord struct {
	PublicID        uuid.UUID
	UserID          uuid.UUID
	CreatedAt       time.Time
	LastUsedAt      time.Time
	ExpiresAt       time.Time
	AuthenticatedAt time.Time
	MFA             bool
	UserAgent       string
}

// Sessions guarda as sessoes do BFF.
type Sessions interface {
	Create(ctx context.Context, s NewSession) error
	Get(ctx context.Context, idHash []byte) (SessionRecord, error) // ErrNotFound se nao existe
	Touch(ctx context.Context, idHash []byte, at time.Time) error
	Delete(ctx context.Context, idHash []byte) error
	ListByUser(ctx context.Context, userID uuid.UUID) ([]SessionRecord, error)
	DeleteByPublicID(ctx context.Context, userID, publicID uuid.UUID) error
	DeleteAllByUser(ctx context.Context, userID uuid.UUID) error
}

// Sealer cifra e autentica valores que vao ao cookie ou ao banco.
type Sealer interface {
	Seal(plain []byte) (string, error)
	Open(sealed string) ([]byte, error)
}
