package identitytest

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/keven-rdr/wordloop/api/internal/identity/app"
	"github.com/keven-rdr/wordloop/api/internal/identity/domain"
)

const GoodCode = "code-ok"

type Provider struct {
	Identity domain.Identity
	Refresh  string

	mu           sync.Mutex
	LastState    string
	LastNonce    string
	LastVerifier string
}

var _ app.Provider = (*Provider)(nil)

func (p *Provider) AuthURL(_ context.Context, state, nonce, verifier string) (string, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.LastState, p.LastNonce, p.LastVerifier = state, nonce, verifier
	return "https://idp.test/auth?state=" + state, nil
}

func (p *Provider) Exchange(_ context.Context, code, verifier, nonce string) (app.Tokens, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if code != GoodCode || verifier != p.LastVerifier || nonce != p.LastNonce {
		return app.Tokens{}, errors.New("provedor recusou a troca")
	}
	return app.Tokens{Identity: p.Identity, RefreshToken: p.Refresh}, nil
}

type Users struct {
	mu   sync.Mutex
	byID map[uuid.UUID]domain.User
	ids  map[domain.Identity]uuid.UUID
}

var _ app.Users = (*Users)(nil)

func (u *Users) Upsert(_ context.Context, id domain.Identity) (domain.User, error) {
	u.mu.Lock()
	defer u.mu.Unlock()
	if u.byID == nil {
		u.byID, u.ids = map[uuid.UUID]domain.User{}, map[domain.Identity]uuid.UUID{}
	}
	key := domain.Identity{Issuer: id.Issuer, Subject: id.Subject}
	uid, ok := u.ids[key]
	if !ok {
		uid = uuid.New()
		u.ids[key] = uid
	}
	user := domain.User{ID: uid, DisplayName: id.DisplayName, Role: domain.RoleUser}
	if old, ok := u.byID[uid]; ok {
		user.Role, user.MFARequired = old.Role, old.MFARequired
	}
	u.byID[uid] = user
	return user, nil
}

func (u *Users) SetRole(id uuid.UUID, role domain.Role) {
	u.mu.Lock()
	defer u.mu.Unlock()
	user := u.byID[id]
	user.Role = role
	u.byID[id] = user
}

func (u *Users) SetMFARequired(_ context.Context, id uuid.UUID, required bool) error {
	u.mu.Lock()
	defer u.mu.Unlock()
	user, ok := u.byID[id]
	if !ok {
		return app.ErrNotFound
	}
	user.MFARequired = required
	u.byID[id] = user
	return nil
}

func (u *Users) Get(_ context.Context, id uuid.UUID) (domain.User, error) {
	u.mu.Lock()
	defer u.mu.Unlock()
	if user, ok := u.byID[id]; ok {
		return user, nil
	}
	return domain.User{}, app.ErrNotFound
}

func (u *Users) Remove(id uuid.UUID) {
	u.mu.Lock()
	defer u.mu.Unlock()
	delete(u.byID, id)
}

type row struct {
	app.NewSession
	PublicID   uuid.UUID
	LastUsedAt time.Time
}

func (r row) record() app.SessionRecord {
	return app.SessionRecord{
		PublicID: r.PublicID, UserID: r.UserID, CreatedAt: r.CreatedAt, LastUsedAt: r.LastUsedAt, ExpiresAt: r.ExpiresAt,
		AuthenticatedAt: r.AuthenticatedAt, MFA: r.MFA, UserAgent: r.UserAgent,
	}
}

type Sessions struct {
	mu      sync.Mutex
	items   map[string]row
	Touched int
}

var _ app.Sessions = (*Sessions)(nil)

func (s *Sessions) Create(_ context.Context, n app.NewSession) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.items == nil {
		s.items = map[string]row{}
	}
	s.items[string(n.IDHash)] = row{NewSession: n, PublicID: uuid.New(), LastUsedAt: n.CreatedAt}
	return nil
}

func (s *Sessions) Get(_ context.Context, hash []byte) (app.SessionRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	r, ok := s.items[string(hash)]
	if !ok {
		return app.SessionRecord{}, app.ErrNotFound
	}
	return r.record(), nil
}

func (s *Sessions) Touch(_ context.Context, hash []byte, at time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Touched++
	if r, ok := s.items[string(hash)]; ok {
		r.LastUsedAt = at
		s.items[string(hash)] = r
	}
	return nil
}

func (s *Sessions) Delete(_ context.Context, hash []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.items, string(hash))
	return nil
}

func (s *Sessions) ListByUser(_ context.Context, userID uuid.UUID) ([]app.SessionRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	var out []app.SessionRecord
	for _, r := range s.items {
		if r.UserID == userID {
			out = append(out, r.record())
		}
	}
	return out, nil
}

func (s *Sessions) DeleteByPublicID(_ context.Context, userID, publicID uuid.UUID) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for k, r := range s.items {
		if r.PublicID == publicID && r.UserID == userID {
			delete(s.items, k)
			return nil
		}
	}
	return app.ErrNotFound
}

func (s *Sessions) DeleteAllByUser(_ context.Context, userID uuid.UUID) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for k, r := range s.items {
		if r.UserID == userID {
			delete(s.items, k)
		}
	}
	return nil
}

func (s *Sessions) Len() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.items)
}

func (s *Sessions) Last() (app.NewSession, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, r := range s.items {
		return r.NewSession, true
	}
	return app.NewSession{}, false
}

func Clock(t time.Time) func() time.Time { return func() time.Time { return t } }
