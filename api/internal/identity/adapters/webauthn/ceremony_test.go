package webauthn_test

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/keven-rdr/wordloop/api/internal/identity/adapters/webauthn"
	"github.com/keven-rdr/wordloop/api/internal/identity/identitytest"
	"github.com/keven-rdr/wordloop/api/internal/identity/passkeys"
)

const (
	origin = "https://tst.wordloop.com.br"
	rpID   = "tst.wordloop.com.br"
)

func newCeremony(t *testing.T) *webauthn.Ceremony {
	t.Helper()
	c, err := webauthn.New(webauthn.Config{RPID: rpID, RPName: "wordloop", Origins: []string{origin}})
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func owner() passkeys.Owner {
	return passkeys.Owner{ID: uuid.New(), Email: "voce@exemplo.com", DisplayName: "Voce"}
}

func register(t *testing.T, c *webauthn.Ceremony, o passkeys.Owner, a *identitytest.SoftAuthenticator) passkeys.Credential {
	t.Helper()
	options, state, err := c.BeginRegistration(o, nil)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := a.Create(options)
	if err != nil {
		t.Fatal(err)
	}
	cred, err := c.FinishRegistration(o, state, resp)
	if err != nil {
		t.Fatalf("registro valido recusado: %v", err)
	}
	cred.UserID = o.ID
	return cred
}

func lookupOf(cred passkeys.Credential, o passkeys.Owner) passkeys.Lookup {
	return func(id []byte) (passkeys.Credential, passkeys.Owner, error) {
		if string(id) != string(cred.CredentialID) {
			return passkeys.Credential{}, passkeys.Owner{}, errors.New("desconhecida")
		}
		return cred, o, nil
	}
}

func userHandle(o passkeys.Owner) []byte {
	id := o.ID
	return id[:]
}

func TestBeginRegistrationAsksForDiscoverableCredentialWithUserVerification(t *testing.T) {
	options, _, err := newCeremony(t).BeginRegistration(owner(), nil)
	if err != nil {
		t.Fatal(err)
	}
	var o struct {
		RP                     struct{ ID string }                            `json:"rp"`
		AuthenticatorSelection struct{ ResidentKey, UserVerification string } `json:"authenticatorSelection"`
		Attestation            string                                         `json:"attestation"`
		Challenge              string                                         `json:"challenge"`
		User                   struct{ ID, Name string }                      `json:"user"`
		Exclude                []map[string]any                               `json:"excludeCredentials"`
	}
	if err := json.Unmarshal(options, &o); err != nil {
		t.Fatal(err)
	}
	if o.RP.ID != rpID || o.AuthenticatorSelection.ResidentKey != "required" || o.AuthenticatorSelection.UserVerification != "required" {
		t.Fatalf("opcoes sem passkey descobrivel com UV: %s", options)
	}
	if o.Attestation != "none" || len(o.Challenge) < 32 || o.User.Name != "voce@exemplo.com" {
		t.Fatalf("opcoes inesperadas: %s", options)
	}
}

func TestRegistrationExcludesExistingCredentials(t *testing.T) {
	c, o := newCeremony(t), owner()
	existing := register(t, c, o, identitytest.NewSoftAuthenticator(origin, rpID))
	options, _, err := c.BeginRegistration(o, []passkeys.Credential{existing})
	if err != nil {
		t.Fatal(err)
	}
	var out struct {
		Exclude []struct{ ID string } `json:"excludeCredentials"`
	}
	if err := json.Unmarshal(options, &out); err != nil || len(out.Exclude) != 1 {
		t.Fatalf("a passkey ja cadastrada deve entrar na lista de exclusao: %s (%v)", options, err)
	}
}

func TestRegistrationStoresPublicKeyAndFlags(t *testing.T) {
	a := identitytest.NewSoftAuthenticator(origin, rpID)
	a.BackupFlag = true
	cred := register(t, newCeremony(t), owner(), a)
	if string(cred.CredentialID) != string(a.CredentialID()) || len(cred.PublicKey) == 0 || !cred.UserVerified || !cred.BackupEligible || !cred.BackupState {
		t.Fatalf("credencial incompleta: %+v", cred)
	}
}

func TestRegistrationRejectsWrongOriginMissingUVAndWrongChallenge(t *testing.T) {
	c, o := newCeremony(t), owner()
	cases := map[string]func(a *identitytest.SoftAuthenticator){
		"origem errada":              func(a *identitytest.SoftAuthenticator) { a.Origin = "https://evil.example" },
		"rpId errado":                func(a *identitytest.SoftAuthenticator) { a.RPID = "evil.example" },
		"sem verificacao do usuario": func(a *identitytest.SoftAuthenticator) { a.NoUV = true },
	}
	for name, mutate := range cases {
		a := identitytest.NewSoftAuthenticator(origin, rpID)
		mutate(a)
		options, state, _ := c.BeginRegistration(o, nil)
		resp, _ := a.Create(options)
		if _, err := c.FinishRegistration(o, state, resp); err == nil {
			t.Errorf("%s: deveria ser recusado", name)
		}
	}
	options1, _, _ := c.BeginRegistration(o, nil)
	_, state2, _ := c.BeginRegistration(o, nil)
	resp, _ := identitytest.NewSoftAuthenticator(origin, rpID).Create(options1)
	if _, err := c.FinishRegistration(o, state2, resp); err == nil {
		t.Error("resposta de outro desafio deveria ser recusada")
	}
	if _, err := c.FinishRegistration(o, []byte("lixo"), resp); err == nil {
		t.Error("estado corrompido deveria ser recusado")
	}
}

func TestLoginWithRegisteredPasskeySucceedsAndAdvancesCounter(t *testing.T) {
	c, o := newCeremony(t), owner()
	a := identitytest.NewSoftAuthenticator(origin, rpID)
	cred := register(t, c, o, a)
	options, state, err := c.BeginLogin()
	if err != nil {
		t.Fatal(err)
	}
	resp, _ := a.Get(options, userHandle(o))
	got, gotOwner, err := c.FinishLogin(state, resp, lookupOf(cred, o))
	if err != nil {
		t.Fatalf("login valido recusado: %v", err)
	}
	if gotOwner.ID != o.ID || got.SignCount != 1 || !got.UserVerified || got.CloneWarning {
		t.Fatalf("resultado inesperado: %+v %+v", got, gotOwner)
	}
}

func TestLoginRejectsForgeriesAndReplays(t *testing.T) {
	c, o := newCeremony(t), owner()
	a := identitytest.NewSoftAuthenticator(origin, rpID)
	cred := register(t, c, o, a)
	other := owner()

	options, state, _ := c.BeginLogin()
	good, _ := a.Get(options, userHandle(o))
	if _, _, err := c.FinishLogin(state, good, lookupOf(cred, other)); err == nil {
		t.Error("userHandle de outra pessoa deveria ser recusado")
	}
	if _, _, err := c.FinishLogin(state, good, func([]byte) (passkeys.Credential, passkeys.Owner, error) {
		return passkeys.Credential{}, passkeys.Owner{}, errors.New("desconhecida")
	}); err == nil {
		t.Error("credencial desconhecida deveria ser recusada")
	}

	_, otherState, _ := c.BeginLogin()
	if _, _, err := c.FinishLogin(otherState, good, lookupOf(cred, o)); err == nil {
		t.Error("resposta de outro desafio deveria ser recusada")
	}

	impostor := identitytest.NewSoftAuthenticator(origin, rpID)
	forged, _ := impostor.Get(options, userHandle(o))
	if _, _, err := c.FinishLogin(state, forged, lookupOf(cred, o)); err == nil {
		t.Error("assinatura de outra chave deveria ser recusada")
	}

	a.NoUV = true
	noUV, _ := a.Get(options, userHandle(o))
	if _, _, err := c.FinishLogin(state, noUV, lookupOf(cred, o)); err == nil {
		t.Error("login sem verificacao do usuario deveria ser recusado")
	}
}

func TestLoginFlagsRegressiveCounterAsPossibleClone(t *testing.T) {
	c, o := newCeremony(t), owner()
	a := identitytest.NewSoftAuthenticator(origin, rpID)
	cred := register(t, c, o, a)
	cred.SignCount = 10
	a.SetCounter(2)
	options, state, _ := c.BeginLogin()
	resp, _ := a.Get(options, userHandle(o))
	got, _, err := c.FinishLogin(state, resp, lookupOf(cred, o))
	if err == nil && !got.CloneWarning {
		t.Fatal("contador regressivo deve levantar o aviso de clone (ou ser recusado)")
	}
}
