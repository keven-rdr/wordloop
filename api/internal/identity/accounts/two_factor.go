package accounts

import (
	"context"

	"github.com/keven-rdr/wordloop/api/internal/identity/domain"
)

type TwoFactor struct{ Store MFAStore }

func (u TwoFactor) Set(ctx context.Context, who domain.Principal, enabled bool) error {
	if who.User.Role != domain.RoleAdmin {
		return ErrForbidden
	}
	if enabled && !who.MFA {
		return ErrWeakSession
	}
	return u.Store.SetMFARequired(ctx, who.User.ID, enabled)
}
