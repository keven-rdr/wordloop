package domain

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

type Identity struct {
	Issuer        string
	Subject       string
	DisplayName   string
	PictureURL    string
	Email         string
	EmailVerified bool
	AuthTime      time.Time
	MFA           bool
}

type User struct {
	ID          uuid.UUID
	DisplayName string
	PictureURL  string
	Role        Role
	MFARequired bool
}

type Principal struct {
	User            User
	SessionID       uuid.UUID
	AuthenticatedAt time.Time
	MFA             bool
}

// CleanPictureURL so aceita https: a foto vem de terceiros e vai para um <img> da interface.
func CleanPictureURL(raw string) string {
	raw = strings.TrimSpace(raw)
	if len(raw) > 2048 || !strings.HasPrefix(raw, "https://") {
		return ""
	}
	return raw
}
