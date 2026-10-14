// Package domain guarda os tipos do modulo identity, sem I/O.
package domain

import (
	"time"

	"github.com/google/uuid"
)

// Identity e o que o provedor OIDC afirma sobre quem entrou (so o necessario; e-mail e senha ficam no provedor).
type Identity struct {
	Issuer      string
	Subject     string
	DisplayName string
	AuthTime    time.Time
	MFA         bool
}

// User e o perfil do app, criado no primeiro login.
type User struct {
	ID          uuid.UUID
	DisplayName string
	Role        Role
}

type Principal struct {
	User            User
	SessionID       uuid.UUID
	AuthenticatedAt time.Time
	MFA             bool
}
