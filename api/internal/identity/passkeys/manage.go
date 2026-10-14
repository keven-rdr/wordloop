package passkeys

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/keven-rdr/wordloop/api/internal/identity/domain"
)

type View struct {
	ID             uuid.UUID
	Name           string
	CreatedAt      time.Time
	LastUsedAt     time.Time
	BackupEligible bool
}

type List struct{ Store Store }

func (u List) Run(ctx context.Context, userID uuid.UUID) ([]View, error) {
	creds, err := u.Store.List(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("listar as passkeys: %w", err)
	}
	out := make([]View, 0, len(creds))
	for _, c := range creds {
		out = append(out, View{ID: c.ID, Name: c.Name, CreatedAt: c.CreatedAt, LastUsedAt: c.LastUsedAt, BackupEligible: c.BackupEligible})
	}
	return out, nil
}

type Delete struct{ Store Store }

func (u Delete) Run(ctx context.Context, who domain.Principal, id uuid.UUID) error {
	allow := func(passkeys, identities int) bool { return canRemove(who.User.Role, passkeys, identities) }
	if err := u.Store.Delete(ctx, who.User.ID, id, allow); err != nil {
		return fmt.Errorf("apagar a passkey: %w", err)
	}
	return nil
}

func canRemove(role domain.Role, passkeys, identities int) bool {
	if passkeys > 1 {
		return true
	}
	return passkeys == 1 && identities > 0 && role != domain.RoleAdmin
}
