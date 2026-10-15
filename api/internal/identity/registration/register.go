package registration

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/keven-rdr/wordloop/api/internal/identity/accounts"
	"github.com/keven-rdr/wordloop/api/internal/identity/domain"
)

type Register struct{ Issuer }

func (u Register) Run(ctx context.Context, in Input) error {
	addr, err := accounts.NormalizeEmail(in.Email)
	if err != nil {
		return ErrInvalidEmail
	}
	acc, err := u.Store.FindByEmail(ctx, addr)
	if errors.Is(err, ErrNotFound) {
		return u.createPending(ctx, addr, in.Locale)
	}
	if err != nil {
		return fmt.Errorf("buscar a conta: %w", err)
	}
	return u.handle(ctx, acc, addr, in.Locale)
}

func (u Register) handle(ctx context.Context, acc Account, addr, locale string) error {
	switch {
	case !acc.Verified:
		return u.issue(ctx, acc.ID, addr, kindVerify, locale)
	case acc.Role == domain.RoleUser:
		return u.issue(ctx, acc.ID, addr, kindExists, locale)
	}
	return nil
}

func (u Register) createPending(ctx context.Context, addr, locale string) error {
	name, _, _ := strings.Cut(addr, "@")
	id, err := u.Store.CreatePending(ctx, addr, name)
	if errors.Is(err, ErrEmailTaken) {
		return u.afterTaken(ctx, addr, locale)
	}
	if err != nil {
		return fmt.Errorf("criar a conta pendente: %w", err)
	}
	return u.issue(ctx, id, addr, kindVerify, locale)
}

func (u Register) afterTaken(ctx context.Context, addr, locale string) error {
	acc, err := u.Store.FindByEmail(ctx, addr)
	if errors.Is(err, ErrNotFound) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("buscar a conta: %w", err)
	}
	return u.handle(ctx, acc, addr, locale)
}

type Forgot struct{ Issuer }

func (u Forgot) Run(ctx context.Context, in Input) error {
	addr, err := accounts.NormalizeEmail(in.Email)
	if err != nil {
		return ErrInvalidEmail
	}
	acc, err := u.Store.FindByEmail(ctx, addr)
	if errors.Is(err, ErrNotFound) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("buscar a conta: %w", err)
	}
	if !acc.Verified || acc.Role != domain.RoleUser {
		return nil
	}
	return u.issue(ctx, acc.ID, addr, kindRecovery, in.Locale)
}
