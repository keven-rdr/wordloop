package domain

import (
	"time"

	"github.com/google/uuid"
)

type Identity struct {
	Issuer        string
	Subject       string
	DisplayName   string
	Email         string
	EmailVerified bool
	AuthTime      time.Time
	MFA           bool
}

type User struct {
	ID          uuid.UUID
	DisplayName string
	Role        Role
	MFARequired bool
}

type Principal struct {
	User            User
	SessionID       uuid.UUID
	AuthenticatedAt time.Time
	MFA             bool
}
