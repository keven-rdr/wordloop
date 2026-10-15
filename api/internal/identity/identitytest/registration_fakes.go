package identitytest

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/keven-rdr/wordloop/api/internal/identity/domain"
	"github.com/keven-rdr/wordloop/api/internal/identity/registration"
	"github.com/keven-rdr/wordloop/api/internal/platform/mail"
)

type MailBox struct {
	mu   sync.Mutex
	Sent []mail.Message
	Err  error
	Done chan struct{}
}

var _ mail.Sender = (*MailBox)(nil)

func (b *MailBox) Send(_ context.Context, m mail.Message) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.Done != nil {
		defer close(b.Done)
	}
	if b.Err != nil {
		return b.Err
	}
	b.Sent = append(b.Sent, m)
	return nil
}

func (b *MailBox) Count() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return len(b.Sent)
}

func (b *MailBox) Last() mail.Message {
	b.mu.Lock()
	defer b.mu.Unlock()
	if len(b.Sent) == 0 {
		return mail.Message{}
	}
	return b.Sent[len(b.Sent)-1]
}

func (b *MailBox) LastSecret() string {
	_, secret, _ := strings.Cut(b.Last().Text, "#")
	fields := strings.Fields(secret)
	if len(fields) == 0 {
		return ""
	}
	return fields[0]
}

type regRow struct {
	account registration.Account
	email   string
}

type regToken struct {
	userID  uuid.UUID
	purpose domain.TokenPurpose
	hash    []byte
	created time.Time
	expires time.Time
}

type RegistrationStore struct {
	mu        sync.Mutex
	rows      map[string]*regRow
	Tokens    []regToken
	FailFind  error
	FailNew   error
	TakenOnce bool
	HideTaken bool
}

var _ registration.Store = (*RegistrationStore)(nil)

func (s *RegistrationStore) init() {
	if s.rows == nil {
		s.rows = map[string]*regRow{}
	}
}

func (s *RegistrationStore) Seed(email string, role domain.Role, verified bool) uuid.UUID {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.init()
	id := uuid.New()
	s.rows[email] = &regRow{account: registration.Account{ID: id, Role: role, Verified: verified}, email: email}
	return id
}

func (s *RegistrationStore) Accounts() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.rows)
}

func (s *RegistrationStore) LastToken() (purpose domain.TokenPurpose, hash []byte, expires time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.Tokens) == 0 {
		return "", nil, time.Time{}
	}
	last := s.Tokens[len(s.Tokens)-1]
	return last.purpose, last.hash, last.expires
}

func (s *RegistrationStore) OpenTokens() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.Tokens)
}

func (s *RegistrationStore) FindByEmail(_ context.Context, email string) (registration.Account, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.init()
	if s.FailFind != nil {
		return registration.Account{}, s.FailFind
	}
	if r, ok := s.rows[strings.ToLower(email)]; ok {
		return r.account, nil
	}
	return registration.Account{}, registration.ErrNotFound
}

func (s *RegistrationStore) CreatePending(_ context.Context, email, _ string) (uuid.UUID, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.init()
	if s.FailNew != nil {
		return uuid.Nil, s.FailNew
	}
	if s.TakenOnce {
		s.TakenOnce = false
		if !s.HideTaken {
			s.rows[email] = &regRow{account: registration.Account{ID: uuid.New(), Role: domain.RoleUser}, email: email}
		}
		return uuid.Nil, registration.ErrEmailTaken
	}
	id := uuid.New()
	s.rows[email] = &regRow{account: registration.Account{ID: id, Role: domain.RoleUser}, email: email}
	return id, nil
}

func (s *RegistrationStore) RecentTokens(_ context.Context, userID uuid.UUID, since time.Time) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	n := 0
	for _, t := range s.Tokens {
		if t.userID == userID && t.created.After(since) {
			n++
		}
	}
	return n, nil
}

func (s *RegistrationStore) ReplaceToken(_ context.Context, userID uuid.UUID, purpose domain.TokenPurpose, hash []byte, createdAt, expiresAt time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Tokens = append(s.Tokens, regToken{userID: userID, purpose: purpose, hash: hash, created: createdAt, expires: expiresAt})
	return nil
}
