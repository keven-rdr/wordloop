// Package identitytest tem implementacoes em memoria das interfaces do modulo identity, SO para testes de unidade.
// O banco e o provedor reais sao cobertos pelos testes de integracao.
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

// GoodCode e o unico code que Provider aceita.
const GoodCode = "code-ok"

// Provider simula o provedor OIDC: aceita GoodCode e devolve Identity.
type Provider struct {
	Identity domain.Identity
	Refresh  string

	mu           sync.Mutex
	LastState    string
	LastNonce    string
	LastVerifier string
}

var _ app.Provider = (*Provider)(nil)

// AuthURL registra o que recebeu e devolve uma URL fixa.
func (p *Provider) AuthURL(_ context.Context, state, nonce, verifier string) (string, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.LastState, p.LastNonce, p.LastVerifier = state, nonce, verifier
	return "https://idp.test/auth?state=" + state, nil
}

// Exchange aceita so GoodCode, com o verifier e o nonce do AuthURL.
func (p *Provider) Exchange(_ context.Context, code, verifier, nonce string) (app.Tokens, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if code != GoodCode || verifier != p.LastVerifier || nonce != p.LastNonce {
		return app.Tokens{}, errors.New("provedor recusou a troca")
	}
	return app.Tokens{Identity: p.Identity, RefreshToken: p.Refresh}, nil
}

// Users e um repositorio de perfis em memoria.
type Users struct {
	mu   sync.Mutex
	byID map[uuid.UUID]domain.User
	ids  map[domain.Identity]uuid.UUID
}

var _ app.Users = (*Users)(nil)

// Upsert cria o perfil na primeira vez e reaproveita o id nas seguintes.
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
	user := domain.User{ID: uid, DisplayName: id.DisplayName}
	u.byID[uid] = user
	return user, nil
}

// Get devolve app.ErrNotFound se o perfil nao existe.
func (u *Users) Get(_ context.Context, id uuid.UUID) (domain.User, error) {
	u.mu.Lock()
	defer u.mu.Unlock()
	if user, ok := u.byID[id]; ok {
		return user, nil
	}
	return domain.User{}, app.ErrNotFound
}

// Remove apaga o perfil (simula usuario excluido).
func (u *Users) Remove(id uuid.UUID) {
	u.mu.Lock()
	defer u.mu.Unlock()
	delete(u.byID, id)
}

// Sessions e um repositorio de sessoes em memoria, indexado pelo hash do id.
type Sessions struct {
	mu      sync.Mutex
	items   map[string]app.NewSession
	Touched int
}

var _ app.Sessions = (*Sessions)(nil)

// Create guarda a sessao.
func (s *Sessions) Create(_ context.Context, n app.NewSession) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.items == nil {
		s.items = map[string]app.NewSession{}
	}
	s.items[string(n.IDHash)] = n
	return nil
}

// Get devolve app.ErrNotFound se nao existe.
func (s *Sessions) Get(_ context.Context, hash []byte) (app.SessionRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	n, ok := s.items[string(hash)]
	if !ok {
		return app.SessionRecord{}, app.ErrNotFound
	}
	return app.SessionRecord{UserID: n.UserID, ExpiresAt: n.ExpiresAt}, nil
}

// Touch conta os usos.
func (s *Sessions) Touch(context.Context, []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Touched++
	return nil
}

// Delete apaga a sessao.
func (s *Sessions) Delete(_ context.Context, hash []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.items, string(hash))
	return nil
}

// Len diz quantas sessoes existem.
func (s *Sessions) Len() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.items)
}

// Last devolve a unica sessao guardada (para conferir o que foi gravado).
func (s *Sessions) Last() (app.NewSession, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, n := range s.items {
		return n, true
	}
	return app.NewSession{}, false
}

// Clock devolve um relogio fixo para os testes.
func Clock(t time.Time) func() time.Time { return func() time.Time { return t } }
