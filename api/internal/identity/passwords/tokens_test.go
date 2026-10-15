package passwords_test

import (
	"errors"
	"testing"
	"time"

	"github.com/keven-rdr/wordloop/api/internal/identity/domain"
	"github.com/keven-rdr/wordloop/api/internal/identity/passwords"
	"github.com/keven-rdr/wordloop/api/internal/platform/token"
)

func addToken(e env, purpose domain.TokenPurpose, role domain.Role) []byte {
	hash := token.Hash("segredo-do-link")
	e.store.AddToken(hash, passwords.Owner{ID: e.id, Email: email, Role: role, Purpose: purpose}, now.Add(10*time.Minute))
	return hash
}

func TestEmailConfirmationTokenSetsThePasswordConfirmsTheEmailAndSignsIn(t *testing.T) {
	e := newEnv(t, domain.RoleUser, false)
	hash := addToken(e, domain.TokenVerifyEmail, domain.RoleUser)
	opened, err := e.set.Run(bg, passwords.SetInput{Token: "segredo-do-link", Password: secret})
	if err != nil || opened == nil {
		t.Fatalf("confirmar: %+v %v", opened, err)
	}
	if stored, _ := e.store.Row(e.id); stored != "h:"+secret || !e.store.Verified(e.id) || !e.store.InviteUsed(hash) {
		t.Fatal("senha gravada, e-mail confirmado e token consumido")
	}
	if err := e.signIn(email, secret); err != nil {
		t.Fatalf("depois de confirmar, entra com e-mail e senha: %v", err)
	}
}

func TestRecoveryTokenReplacesThePasswordOfAStudent(t *testing.T) {
	e := newEnv(t, domain.RoleUser, false).withPassword()
	addToken(e, domain.TokenRecovery, domain.RoleUser)
	if _, err := e.set.Run(bg, passwords.SetInput{Token: "segredo-do-link", Password: "uma senha nova e comprida"}); err != nil {
		t.Fatal(err)
	}
	if err := e.signIn(email, secret); !errors.Is(err, passwords.ErrInvalidCredentials) {
		t.Fatalf("a senha antiga deixa de valer: %v", err)
	}
	if err := e.signIn(email, "uma senha nova e comprida"); err != nil {
		t.Fatalf("a nova senha entra: %v", err)
	}
}

func TestRecoveryTokenNeverWorksForAnAdministrator(t *testing.T) {
	e := newEnv(t, domain.RoleAdmin, false).withPassword()
	hash := addToken(e, domain.TokenRecovery, domain.RoleAdmin)
	if _, err := e.set.Run(bg, passwords.SetInput{Token: "segredo-do-link", Password: "uma senha nova e comprida"}); !errors.Is(err, passwords.ErrUnauthorized) {
		t.Fatalf("admin so se recupera pelo servidor: %v", err)
	}
	if stored, _ := e.store.Row(e.id); stored != "h:"+secret || e.store.InviteUsed(hash) {
		t.Fatal("nada pode mudar nem o token ser gasto")
	}
}

func TestAnInviteDoesNotConfirmAnyEmail(t *testing.T) {
	e := newEnv(t, domain.RoleAdmin, false)
	invite(e, now.Add(5*time.Minute))
	if _, err := e.set.Run(bg, passwords.SetInput{Token: "segredo-do-convite", Password: secret}); err != nil {
		t.Fatal(err)
	}
	if e.store.Verified(e.id) {
		t.Fatal("o convite do admin ja nasce com o e-mail confirmado no banco; o fake so marca confirmacao por e-mail")
	}
}
