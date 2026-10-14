package webauthn

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/go-webauthn/webauthn/protocol"
	lib "github.com/go-webauthn/webauthn/webauthn"

	"github.com/keven-rdr/wordloop/api/internal/identity/passkeys"
)

type Config struct {
	RPID    string
	RPName  string
	Origins []string
}

type Ceremony struct{ wa *lib.WebAuthn }

var _ passkeys.Ceremony = (*Ceremony)(nil)

func New(cfg Config) (*Ceremony, error) {
	wa, err := lib.New(&lib.Config{RPID: cfg.RPID, RPDisplayName: cfg.RPName, RPOrigins: cfg.Origins})
	if err != nil {
		return nil, fmt.Errorf("configurar o webauthn: %w", err)
	}
	return &Ceremony{wa: wa}, nil
}

func (c *Ceremony) BeginRegistration(owner passkeys.Owner, existing []passkeys.Credential) ([]byte, []byte, error) {
	required := true
	creation, session, err := c.wa.BeginRegistration(
		asUser(owner, existing),
		lib.WithAuthenticatorSelection(protocol.AuthenticatorSelection{
			RequireResidentKey: &required,
			ResidentKey:        protocol.ResidentKeyRequirementRequired,
			UserVerification:   protocol.VerificationRequired,
		}),
		lib.WithConveyancePreference(protocol.PreferNoAttestation),
		lib.WithExclusions(descriptors(existing)),
	)
	if err != nil {
		return nil, nil, err
	}
	return marshalPair(creation.Response, session)
}

func (c *Ceremony) FinishRegistration(owner passkeys.Owner, state, credentialJSON []byte) (passkeys.Credential, error) {
	var session lib.SessionData
	if err := json.Unmarshal(state, &session); err != nil {
		return passkeys.Credential{}, fmt.Errorf("estado do desafio: %w", err)
	}
	parsed, err := protocol.ParseCredentialCreationResponseBytes(credentialJSON)
	if err != nil {
		return passkeys.Credential{}, err
	}
	cred, err := c.wa.CreateCredential(asUser(owner, nil), session, parsed)
	if err != nil {
		return passkeys.Credential{}, err
	}
	return fromLib(cred), nil
}

func (c *Ceremony) BeginLogin() ([]byte, []byte, error) {
	assertion, session, err := c.wa.BeginDiscoverableLogin(lib.WithUserVerification(protocol.VerificationRequired))
	if err != nil {
		return nil, nil, err
	}
	return marshalPair(assertion.Response, session)
}

func (c *Ceremony) FinishLogin(state, credentialJSON []byte, lookup passkeys.Lookup) (passkeys.Credential, passkeys.Owner, error) {
	var session lib.SessionData
	if err := json.Unmarshal(state, &session); err != nil {
		return passkeys.Credential{}, passkeys.Owner{}, fmt.Errorf("estado do desafio: %w", err)
	}
	parsed, err := protocol.ParseCredentialRequestResponseBytes(credentialJSON)
	if err != nil {
		return passkeys.Credential{}, passkeys.Owner{}, err
	}
	var owner passkeys.Owner
	handler := func(rawID, userHandle []byte) (lib.User, error) {
		stored, found, err := lookup(rawID)
		if err != nil {
			return nil, err
		}
		if !sameUser(userHandle, found) {
			return nil, errors.New("user handle diferente do dono da credencial")
		}
		owner = found
		return asUser(found, []passkeys.Credential{stored}), nil
	}
	_, cred, err := c.wa.ValidatePasskeyLogin(handler, session, parsed)
	if err != nil {
		return passkeys.Credential{}, passkeys.Owner{}, err
	}
	return fromLib(cred), owner, nil
}

func marshalPair(options any, session *lib.SessionData) ([]byte, []byte, error) {
	opts, err := json.Marshal(options)
	if err != nil {
		return nil, nil, err
	}
	state, err := json.Marshal(session)
	if err != nil {
		return nil, nil, err
	}
	return opts, state, nil
}

func sameUser(userHandle []byte, owner passkeys.Owner) bool {
	id := owner.ID
	return string(userHandle) == string(id[:])
}
