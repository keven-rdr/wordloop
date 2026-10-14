package webauthn

import (
	"github.com/go-webauthn/webauthn/protocol"
	lib "github.com/go-webauthn/webauthn/webauthn"

	"github.com/keven-rdr/wordloop/api/internal/identity/passkeys"
)

type account struct {
	owner passkeys.Owner
	creds []lib.Credential
}

func asUser(owner passkeys.Owner, creds []passkeys.Credential) account {
	out := make([]lib.Credential, 0, len(creds))
	for _, c := range creds {
		out = append(out, toLib(c))
	}
	return account{owner: owner, creds: out}
}

func (a account) WebAuthnID() []byte {
	id := a.owner.ID
	return id[:]
}

func (a account) WebAuthnName() string {
	if a.owner.Email != "" {
		return a.owner.Email
	}
	return a.WebAuthnDisplayName()
}

func (a account) WebAuthnDisplayName() string {
	if a.owner.DisplayName != "" {
		return a.owner.DisplayName
	}
	return a.owner.ID.String()
}

func (a account) WebAuthnCredentials() []lib.Credential { return a.creds }

func toLib(c passkeys.Credential) lib.Credential {
	transports := make([]protocol.AuthenticatorTransport, 0, len(c.Transports))
	for _, t := range c.Transports {
		transports = append(transports, protocol.AuthenticatorTransport(t))
	}
	return lib.Credential{
		ID:        c.CredentialID,
		PublicKey: c.PublicKey,
		Transport: transports,
		Flags: lib.CredentialFlags{
			UserPresent: true, UserVerified: c.UserVerified, BackupEligible: c.BackupEligible, BackupState: c.BackupState,
		},
		Authenticator: lib.Authenticator{AAGUID: c.AAGUID, SignCount: c.SignCount},
	}
}

func fromLib(c *lib.Credential) passkeys.Credential {
	transports := make([]string, 0, len(c.Transport))
	for _, t := range c.Transport {
		transports = append(transports, string(t))
	}
	return passkeys.Credential{
		CredentialID:   c.ID,
		PublicKey:      c.PublicKey,
		SignCount:      c.Authenticator.SignCount,
		Transports:     transports,
		UserVerified:   c.Flags.UserVerified,
		BackupEligible: c.Flags.BackupEligible,
		BackupState:    c.Flags.BackupState,
		CloneWarning:   c.Authenticator.CloneWarning,
		AAGUID:         c.Authenticator.AAGUID,
	}
}

func descriptors(creds []passkeys.Credential) []protocol.CredentialDescriptor {
	out := make([]protocol.CredentialDescriptor, 0, len(creds))
	for _, c := range creds {
		l := toLib(c)
		out = append(out, l.Descriptor())
	}
	return out
}
