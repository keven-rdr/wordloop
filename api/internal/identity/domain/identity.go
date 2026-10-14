// Package domain guarda os tipos do modulo identity, sem I/O.
package domain

import (
	"time"

	"github.com/google/uuid"
)

// Identity e o que o provedor OIDC afirma sobre quem entrou (so o necessario; e-mail e senha ficam no provedor).
// AuthTime e MFA descrevem COMO a pessoa se autenticou; sao a unica prova de 2 fatores que aceitamos para admin.
type Identity struct {
	Issuer      string
	Subject     string
	DisplayName string
	AuthTime    time.Time // quando o provedor autenticou a pessoa; zero se o provedor nao informou
	MFA         bool      // o provedor afirma que houve 2 fatores (claim amr)
}

// User e o perfil do app, criado no primeiro login.
type User struct {
	ID          uuid.UUID
	DisplayName string
	Role        Role
}

// Principal e quem esta fazendo o pedido, descoberto SO pela sessao (cookie) e pelo banco, nunca pelo corpo do pedido.
type Principal struct {
	User            User
	SessionID       uuid.UUID // identificador publico da sessao atual (nao e o segredo do cookie)
	AuthenticatedAt time.Time
	MFA             bool
}
