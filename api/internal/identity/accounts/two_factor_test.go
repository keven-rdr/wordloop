package accounts_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/keven-rdr/wordloop/api/internal/identity/accounts"
	"github.com/keven-rdr/wordloop/api/internal/identity/domain"
)

type mfaStore struct {
	got  map[uuid.UUID]bool
	fail error
}

func (s *mfaStore) SetMFARequired(_ context.Context, id uuid.UUID, required bool) error {
	if s.fail != nil {
		return s.fail
	}
	s.got[id] = required
	return nil
}

func principal(role domain.Role, mfa bool) domain.Principal {
	return domain.Principal{User: domain.User{ID: uuid.New(), Role: role}, MFA: mfa}
}

func TestTwoFactorOnlyAdminsCanChangeIt(t *testing.T) {
	store := &mfaStore{got: map[uuid.UUID]bool{}}
	uc := accounts.TwoFactor{Store: store}
	user := principal(domain.RoleUser, true)
	if err := uc.Set(context.Background(), user, true); !errors.Is(err, accounts.ErrForbidden) {
		t.Fatalf("aluno: esperava ErrForbidden, veio %v", err)
	}
	if len(store.got) != 0 {
		t.Fatal("nada pode ser gravado para quem nao e admin")
	}
}

func TestTwoFactorTurningOnNeedsASessionThatAlreadyProvedIt(t *testing.T) {
	store := &mfaStore{got: map[uuid.UUID]bool{}}
	uc := accounts.TwoFactor{Store: store}
	weak := principal(domain.RoleAdmin, false)
	if err := uc.Set(context.Background(), weak, true); !errors.Is(err, accounts.ErrWeakSession) {
		t.Fatalf("sessao sem 2 fatores: esperava ErrWeakSession, veio %v", err)
	}
	if len(store.got) != 0 {
		t.Fatal("ligar sem prova poderia trancar o admin fora")
	}
	if err := uc.Set(context.Background(), weak, false); err != nil || store.got[weak.User.ID] {
		t.Fatalf("desligar sempre pode: %v %v", err, store.got)
	}
}

func TestTwoFactorAdminWithProofTurnsItOnAndOff(t *testing.T) {
	store := &mfaStore{got: map[uuid.UUID]bool{}}
	uc := accounts.TwoFactor{Store: store}
	admin := principal(domain.RoleAdmin, true)
	if err := uc.Set(context.Background(), admin, true); err != nil || !store.got[admin.User.ID] {
		t.Fatalf("ligar: %v %v", err, store.got)
	}
	if err := uc.Set(context.Background(), admin, false); err != nil || store.got[admin.User.ID] {
		t.Fatalf("desligar: %v %v", err, store.got)
	}
}

func TestTwoFactorPropagatesStoreFailures(t *testing.T) {
	boom := errors.New("banco fora")
	uc := accounts.TwoFactor{Store: &mfaStore{got: map[uuid.UUID]bool{}, fail: boom}}
	if err := uc.Set(context.Background(), principal(domain.RoleAdmin, true), true); !errors.Is(err, boom) {
		t.Fatalf("esperava o erro do banco, veio %v", err)
	}
}
