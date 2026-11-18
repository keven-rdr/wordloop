// Package progresstest tem dubles em memoria do progresso; so para testes.
package progresstest

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/keven-rdr/wordloop/api/internal/progress/app"
	"github.com/keven-rdr/wordloop/api/internal/progress/domain"
)

// Identity devolve sempre o mesmo Viewer (ou o erro configurado).
type Identity struct {
	Is  domain.Viewer
	Err error
}

var _ app.Identity = (*Identity)(nil)

func (v *Identity) Who(context.Context) (domain.Viewer, error) {
	if v.Err != nil {
		return domain.Viewer{}, v.Err
	}
	if v.Is.UserID == uuid.Nil {
		return domain.Viewer{}, app.ErrNoSession
	}
	return v.Is, nil
}

// Store e um leitor em memoria: o que o estudo gravou para uma pessoa.
type Store struct {
	Config domain.Settings
	Rows   []domain.Day
	Due    map[int]int
	Known  int
	Missed []domain.Missed
	FailOn string // nome do metodo que deve falhar
	Err    error
}

var _ app.Reader = (*Store)(nil)

// New cria um leitor com os ajustes padrao.
func New() *Store { return &Store{Config: domain.DefaultSettings, Due: map[int]int{}} }

func (s *Store) fail(method string) error {
	if s.FailOn == method {
		return s.Err
	}
	return nil
}

func (s *Store) Settings(context.Context, uuid.UUID) (domain.Settings, error) {
	return s.Config, s.fail("Settings")
}

func (s *Store) Days(_ context.Context, _ uuid.UUID, from, to time.Time) ([]domain.Day, error) {
	var out []domain.Day
	for _, d := range s.Rows {
		if !d.Date.Before(from) && !d.Date.After(to) && d.Reviews > 0 {
			out = append(out, d)
		}
	}
	return out, s.fail("Days")
}

func (s *Store) Forecast(context.Context, uuid.UUID, domain.Settings, time.Time, int) (map[int]int, error) {
	return s.Due, s.fail("Forecast")
}

func (s *Store) KnownWords(context.Context, uuid.UUID) (int, error) {
	return s.Known, s.fail("KnownWords")
}

func (s *Store) MostMissed(context.Context, uuid.UUID, time.Time, string, int) ([]domain.Missed, error) {
	return s.Missed, s.fail("MostMissed")
}
