package identitytest

import (
	"context"
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/keven-rdr/wordloop/api/internal/identity/domain"
	"github.com/keven-rdr/wordloop/api/internal/identity/passwords"
)

type PlainHasher struct {
	mu    sync.Mutex
	Calls []string
}

func (h *PlainHasher) Hash(_ context.Context, plain string) (string, error) { return "h:" + plain, nil }

func (h *PlainHasher) Verify(_ context.Context, plain, encoded string) (bool, error) {
	h.mu.Lock()
	h.Calls = append(h.Calls, encoded)
	h.mu.Unlock()
	if !strings.HasPrefix(encoded, "h:") {
		return false, errors.New("hash invalido")
	}
	return encoded == "h:"+plain, nil
}

type pwRow struct {
	passwords.Account
	Email    string
	Attempts int
	Verified bool
}

type pwInvite struct {
	owner passwords.Owner
	used  bool
	until time.Time
}

type PasswordStore struct {
	mu      sync.Mutex
	rows    map[uuid.UUID]*pwRow
	invites map[string]pwInvite
}

var _ passwords.Store = (*PasswordStore)(nil)

func (s *PasswordStore) init() {
	if s.rows == nil {
		s.rows, s.invites = map[uuid.UUID]*pwRow{}, map[string]pwInvite{}
	}
}

func (s *PasswordStore) Add(id uuid.UUID, email, hash string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.init()
	s.rows[id] = &pwRow{Account: passwords.Account{ID: id, Hash: hash}, Email: email}
}

func (s *PasswordStore) AddInvite(hash []byte, id uuid.UUID, email string, until time.Time) {
	s.AddToken(hash, passwords.Owner{ID: id, Email: email, Role: domain.RoleUser, Purpose: domain.TokenInvite}, until)
}

func (s *PasswordStore) AddToken(hash []byte, owner passwords.Owner, until time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.init()
	if _, ok := s.rows[owner.ID]; !ok {
		s.rows[owner.ID] = &pwRow{Account: passwords.Account{ID: owner.ID}, Email: owner.Email}
	}
	s.invites[string(hash)] = pwInvite{owner: owner, until: until}
}

func (s *PasswordStore) Verified(id uuid.UUID) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.rows[id] != nil && s.rows[id].Verified
}

func (s *PasswordStore) Row(id uuid.UUID) (hash string, attempts int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if r, ok := s.rows[id]; ok {
		return r.Hash, r.Attempts
	}
	return "", 0
}

func (s *PasswordStore) InviteUsed(hash []byte) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.invites[string(hash)].used
}

func (s *PasswordStore) FindByEmail(_ context.Context, email string) (passwords.Account, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, r := range s.rows {
		if strings.EqualFold(r.Email, email) {
			return r.Account, nil
		}
	}
	return passwords.Account{}, passwords.ErrNotFound
}

func (s *PasswordStore) BeginAttempt(_ context.Context, id uuid.UUID, at time.Time) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	r := s.rows[id]
	if r.LockedUntil.After(at) {
		return false, nil
	}
	r.Attempts++
	if r.Attempts >= 5 {
		r.LockedUntil = at.Add(time.Minute)
	}
	return true, nil
}

func (s *PasswordStore) RegisterSuccess(_ context.Context, id uuid.UUID) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	r := s.rows[id]
	r.Attempts, r.LockedUntil = 0, time.Time{}
	return nil
}

func (s *PasswordStore) PeekToken(_ context.Context, hash []byte, now time.Time) (passwords.Owner, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	inv, ok := s.invites[string(hash)]
	if !ok || inv.used || !inv.until.After(now) {
		return passwords.Owner{}, passwords.ErrNotFound
	}
	return inv.owner, nil
}

func (s *PasswordStore) SetWithToken(_ context.Context, hash []byte, userID uuid.UUID, encoded string, now time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	inv, ok := s.invites[string(hash)]
	if !ok || inv.used || !inv.until.After(now) || inv.owner.ID != userID {
		return passwords.ErrUnauthorized
	}
	inv.used = true
	s.invites[string(hash)] = inv
	r := s.rows[userID]
	r.Hash, r.Attempts, r.LockedUntil = encoded, 0, time.Time{}
	if inv.owner.Purpose != domain.TokenInvite {
		r.Verified = true
	}
	return nil
}
