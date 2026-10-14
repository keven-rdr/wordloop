package identitytest

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/keven-rdr/wordloop/api/internal/identity/passkeys"
	"github.com/keven-rdr/wordloop/api/internal/platform/token"
)

type PasskeyStore struct {
	mu         sync.Mutex
	owners     map[uuid.UUID]passkeys.Owner
	creds      []passkeys.Credential
	invites    map[string]invite
	identities map[uuid.UUID]int
	Failing    error
}

type invite struct {
	user    uuid.UUID
	expires time.Time
	used    bool
}

var _ passkeys.Store = (*PasskeyStore)(nil)

func NewPasskeyStore() *PasskeyStore {
	return &PasskeyStore{owners: map[uuid.UUID]passkeys.Owner{}, invites: map[string]invite{}, identities: map[uuid.UUID]int{}}
}

func (s *PasskeyStore) AddOwner(o passkeys.Owner) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.owners[o.ID] = o
}

func (s *PasskeyStore) AddInvite(tok string, user uuid.UUID, expires time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.invites[string(token.Hash(tok))] = invite{user: user, expires: expires}
}

func (s *PasskeyStore) SetIdentities(user uuid.UUID, n int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.identities[user] = n
}

func (s *PasskeyStore) Credentials() []passkeys.Credential {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]passkeys.Credential(nil), s.creds...)
}

func (s *PasskeyStore) InviteUsed(tok string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.invites[string(token.Hash(tok))].used
}

func (s *PasskeyStore) List(_ context.Context, userID uuid.UUID) ([]passkeys.Credential, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	var out []passkeys.Credential
	for _, c := range s.creds {
		if c.UserID == userID {
			out = append(out, c)
		}
	}
	return out, s.Failing
}

func (s *PasskeyStore) Find(_ context.Context, credentialID []byte) (passkeys.Credential, passkeys.Owner, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, c := range s.creds {
		if string(c.CredentialID) == string(credentialID) {
			return c, s.owners[c.UserID], nil
		}
	}
	return passkeys.Credential{}, passkeys.Owner{}, passkeys.ErrNotFound
}

func (s *PasskeyStore) Owner(_ context.Context, userID uuid.UUID) (passkeys.Owner, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	o, ok := s.owners[userID]
	if !ok {
		return passkeys.Owner{}, passkeys.ErrNotFound
	}
	return o, nil
}

func (s *PasskeyStore) PeekInvite(_ context.Context, hash []byte, now time.Time) (uuid.UUID, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	inv, ok := s.invites[string(hash)]
	if !ok || inv.used || !now.Before(inv.expires) {
		return uuid.Nil, passkeys.ErrUnauthorized
	}
	return inv.user, nil
}

func (s *PasskeyStore) Add(_ context.Context, c passkeys.Credential) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.insert(c)
}

func (s *PasskeyStore) AddWithInvite(_ context.Context, c passkeys.Credential, hash []byte, now time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	inv, ok := s.invites[string(hash)]
	if !ok || inv.used || !now.Before(inv.expires) || inv.user != c.UserID {
		return passkeys.ErrUnauthorized
	}
	if err := s.insert(c); err != nil {
		return err
	}
	inv.used = true
	s.invites[string(hash)] = inv
	return nil
}

func (s *PasskeyStore) insert(c passkeys.Credential) error {
	n := 0
	for _, e := range s.creds {
		if string(e.CredentialID) == string(c.CredentialID) {
			return passkeys.ErrDuplicate
		}
		if e.UserID == c.UserID {
			n++
		}
	}
	if n >= passkeys.MaxPerAccount {
		return passkeys.ErrTooMany
	}
	s.creds = append(s.creds, c)
	return nil
}

func (s *PasskeyStore) RecordLogin(_ context.Context, c passkeys.Credential, at time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, e := range s.creds {
		if string(e.CredentialID) == string(c.CredentialID) {
			s.creds[i].SignCount, s.creds[i].LastUsedAt = c.SignCount, at
		}
	}
	return nil
}

func (s *PasskeyStore) Delete(_ context.Context, userID, id uuid.UUID, allow func(passkeys, identities int) bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	owned, at := 0, -1
	for i, c := range s.creds {
		if c.UserID != userID {
			continue
		}
		owned++
		if c.ID == id {
			at = i
		}
	}
	if at < 0 {
		return passkeys.ErrNotFound
	}
	if !allow(owned, s.identities[userID]) {
		return passkeys.ErrLastMethod
	}
	s.creds = append(s.creds[:at], s.creds[at+1:]...)
	return nil
}

type MemChallenges struct {
	mu    sync.Mutex
	items map[string]passkeys.Challenge
}

var _ passkeys.Challenges = (*MemChallenges)(nil)

func NewMemChallenges() *MemChallenges { return &MemChallenges{items: map[string]passkeys.Challenge{}} }

func (m *MemChallenges) Save(_ context.Context, c passkeys.Challenge) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.items[string(c.IDHash)] = c
	return nil
}

func (m *MemChallenges) Take(_ context.Context, hash []byte, purpose passkeys.Purpose, now time.Time) (passkeys.Challenge, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	c, ok := m.items[string(hash)]
	if !ok || c.Purpose != purpose || !now.Before(c.ExpiresAt) {
		return passkeys.Challenge{}, passkeys.ErrNotFound
	}
	delete(m.items, string(hash))
	return c, nil
}

func (m *MemChallenges) Len() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.items)
}
