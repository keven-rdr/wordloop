// Package app tem os casos de uso dos direitos da pessoa sobre os dados: exportar tudo e apagar a conta.
package app

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/keven-rdr/wordloop/api/internal/privacy/domain"
)

// ErrNoSession e a falta de sessao valida.
var ErrNoSession = errors.New("privacidade: pedido sem sessao")

// Store e o que os casos de uso precisam do banco. Declarada aqui, pelo consumidor.
type Store interface {
	// Email devolve o e-mail da conta, vazio se nao tem; ErrNotFound se a conta nao existe.
	Email(ctx context.Context, user uuid.UUID) (string, error)
	// Export devolve os dados da pessoa (secoes em um objeto JSON) numa instantanea so.
	Export(ctx context.Context, user uuid.UUID) (domain.Export, error)
	// Erase apaga a conta e tudo o que e dela, numa transacao so, e registra a exclusao sem dado pessoal.
	Erase(ctx context.Context, user uuid.UUID, at time.Time) (domain.ErasureReport, error)
}

// Identity diz quem esta pedindo, pela sessao; sem sessao devolve ErrNoSession.
type Identity interface {
	Who(ctx context.Context) (domain.Subject, error)
}

// Clock e a fonte de tempo (trocavel nos testes); nulo usa o relogio do sistema.
type Clock func() time.Time

func (c Clock) now() time.Time {
	read := time.Now
	if c != nil {
		read = c
	}
	return read().UTC()
}

// Service junta exportar e apagar.
type Service struct {
	Identity Identity
	Store    Store
	Now      Clock
}

func (s Service) subject(ctx context.Context) (domain.Subject, error) {
	who, err := s.Identity.Who(ctx)
	if err == nil && who.UserID == uuid.Nil {
		err = ErrNoSession
	}
	return who, err
}

// Export devolve todos os dados de quem pede.
func (s Service) Export(ctx context.Context) (domain.Export, error) {
	who, err := s.subject(ctx)
	if err != nil {
		return domain.Export{}, err
	}
	out, err := s.Store.Export(ctx, who.UserID)
	out.GeneratedAt = s.Now.now()
	return out, err
}

// Erase apaga a conta de quem pede, se a confirmacao, o tipo de conta e o login recente permitirem.
func (s Service) Erase(ctx context.Context, confirmation string) (domain.ErasureReport, error) {
	who, err := s.subject(ctx)
	if err != nil {
		return domain.ErasureReport{}, err
	}
	email, err := s.Store.Email(ctx, who.UserID)
	if err != nil {
		return domain.ErasureReport{}, err
	}
	now := s.Now.now()
	if err := domain.CheckErase(who, email, confirmation, now); err != nil {
		return domain.ErasureReport{}, err
	}
	return s.Store.Erase(ctx, who.UserID, now)
}
