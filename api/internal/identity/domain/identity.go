package domain

import (
	"time"

	"github.com/google/uuid"
)

type Identity struct {
	Issuer      string
	Subject     string
	DisplayName string
	AuthTime    time.Time
	MFA         bool
}

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
