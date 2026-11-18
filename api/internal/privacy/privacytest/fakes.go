// Package privacytest tem dubles em memoria da privacidade; so para testes.
package privacytest

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/keven-rdr/wordloop/api/internal/privacy/app"
	"github.com/keven-rdr/wordloop/api/internal/privacy/domain"
)

// Identity devolve sempre o mesmo Subject (ou o erro configurado).
type Identity struct {
	Is  domain.Subject
	Err error
}

var _ app.Identity = (*Identity)(nil)

func (i *Identity) Who(context.Context) (domain.Subject, error) {
	if i.Err != nil {
		return domain.Subject{}, i.Err
	}
	return i.Is, nil
}

// ErrBoom e a falha simulada.
var ErrBoom = errors.New("falha simulada")

// Store e um repositorio em memoria: o e-mail da conta, o que exporta e o que foi apagado.
type Store struct {
	Emails  map[uuid.UUID]string
	Data    json.RawMessage
	Erased  []uuid.UUID
	Reviews int
	FailOn  string // nome do metodo que deve falhar
}

var _ app.Store = (*Store)(nil)

// New cria um repositorio vazio que exporta um objeto JSON vazio.
func New() *Store { return &Store{Emails: map[uuid.UUID]string{}, Data: json.RawMessage(`{}`)} }

func (s *Store) fail(method string) error {
	if s.FailOn == method {
		return ErrBoom
	}
	return nil
}

func (s *Store) Email(_ context.Context, user uuid.UUID) (string, error) {
	email, found := s.Emails[user]
	if !found {
		return "", domain.ErrNotFound
	}
	return email, s.fail("Email")
}

func (s *Store) Export(context.Context, uuid.UUID) (domain.Export, error) {
	return domain.Export{Data: s.Data}, s.fail("Export")
}

func (s *Store) Erase(_ context.Context, user uuid.UUID, _ time.Time) (domain.ErasureReport, error) {
	if err := s.fail("Erase"); err != nil {
		return domain.ErasureReport{}, err
	}
	s.Erased = append(s.Erased, user)
	delete(s.Emails, user)
	return domain.ErasureReport{Reviews: s.Reviews}, nil
}
