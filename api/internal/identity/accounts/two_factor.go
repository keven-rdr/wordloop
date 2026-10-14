package accounts

import (
	"context"

	"github.com/keven-rdr/wordloop/api/internal/identity/domain"
)

// TwoFactor liga ou desliga a exigencia de 2 fatores no login da propria conta. Por ora so administradores
// podem usar. Ligar exige que a sessao atual ja tenha comprovado os 2 fatores, para ninguem se trancar fora.
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
