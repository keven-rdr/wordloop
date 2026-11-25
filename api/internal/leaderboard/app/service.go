// Package app tem os casos de uso do ranking: o podio do periodo, o resumo de pontos e o perfil (apelido).
package app

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/keven-rdr/wordloop/api/internal/leaderboard/domain"
)

// ErrNoSession e o pedido sem sessao valida.
var ErrNoSession = errors.New("sem sessao valida")

// MaxLimit e o maximo de posicoes de uma pagina do ranking.
const MaxLimit = 100

// Identity descobre quem pergunta; sem sessao devolve ErrNoSession. Declarada aqui, pelo consumidor.
type Identity interface {
	Who(ctx context.Context) (domain.Viewer, error)
}

// Store le os pontos e grava o perfil. Ranking devolve as limit primeiras posicoes e a linha de me.
type Store interface {
	Clock(ctx context.Context, user uuid.UUID) (tz string, cutoffHour int, err error)
	Ranking(ctx context.Context, me uuid.UUID, from, to time.Time, limit int) ([]domain.Entry, error)
	DaysMet(ctx context.Context, user uuid.UUID, from, to time.Time) (int, error)
	Profile(ctx context.Context, user uuid.UUID) (domain.Profile, bool, error)
	// SaveProfile devolve domain.ErrNicknameTaken se outra pessoa ja usa o apelido.
	SaveProfile(ctx context.Context, user uuid.UUID, p domain.Profile) (domain.Profile, error)
}

// Service junta os casos de uso.
type Service struct {
	Identity Identity
	Store    Store
	Now      func() time.Time
}

func (s Service) now() time.Time {
	if s.Now == nil {
		return time.Now().UTC()
	}
	return s.Now().UTC()
}

func (s Service) who(ctx context.Context) (domain.Viewer, error) {
	v, err := s.Identity.Who(ctx)
	if err == nil && v.UserID == uuid.Nil {
		err = ErrNoSession
	}
	return v, err
}

// today e o dia de estudo de hoje de quem pergunta (a semana do ranking e a da pessoa).
func (s Service) today(ctx context.Context, user uuid.UUID) (time.Time, error) {
	tz, cutoff, err := s.Store.Clock(ctx, user)
	if err != nil {
		return time.Time{}, err
	}
	return domain.Today(s.now(), tz, cutoff), nil
}

// Board e o ranking do periodo: as primeiras limit posicoes (1 a MaxLimit) e a linha de quem pergunta.
func (s Service) Board(ctx context.Context, period domain.Period, limit int) (domain.Board, error) {
	v, err := s.who(ctx)
	if err != nil {
		return domain.Board{}, err
	}
	today, err := s.today(ctx, v.UserID)
	if err != nil {
		return domain.Board{}, err
	}
	from, to, err := period.Range(today)
	if err != nil {
		return domain.Board{}, err
	}
	limit = min(MaxLimit, max(1, limit))
	rows, err := s.Store.Ranking(ctx, v.UserID, from, to, limit)
	if err != nil {
		return domain.Board{}, err
	}
	return split(period, from, to, rows, v.UserID, limit), nil
}

// split separa as primeiras posicoes da linha de quem pergunta (que pode estar nelas ou abaixo).
func split(period domain.Period, from, to time.Time, rows []domain.Entry, me uuid.UUID, limit int) domain.Board {
	b := domain.Board{Period: period, From: from, To: to, Entries: []domain.Entry{}}
	for i := range rows {
		if rows[i].UserID == me {
			b.Me = &rows[i]
		}
		if rows[i].Position <= limit && (rows[i].Nickname != "" || rows[i].UserID != me) {
			b.Entries = append(b.Entries, rows[i])
		}
	}
	return b
}

// Score e o resumo de quem pergunta: pontos da semana e de sempre, nivel e dias da meta cumpridos na semana.
func (s Service) Score(ctx context.Context) (domain.Score, error) {
	v, err := s.who(ctx)
	if err != nil {
		return domain.Score{}, err
	}
	today, err := s.today(ctx, v.UserID)
	if err != nil {
		return domain.Score{}, err
	}
	week, err := s.points(ctx, v.UserID, domain.Week, today)
	if err != nil {
		return domain.Score{}, err
	}
	total, err := s.points(ctx, v.UserID, domain.All, today)
	if err != nil {
		return domain.Score{}, err
	}
	from, to, _ := domain.Week.Range(today)
	met, err := s.Store.DaysMet(ctx, v.UserID, from, to)
	if err != nil {
		return domain.Score{}, err
	}
	return domain.Score{WeekPoints: week, TotalPoints: total, WeekDaysMet: met, Level: domain.LevelOf(total)}, nil
}

// points sao os pontos de user no periodo: a linha dele no ranking (sem as demais posicoes); sem linha, 0.
func (s Service) points(ctx context.Context, user uuid.UUID, period domain.Period, today time.Time) (int, error) {
	from, to, _ := period.Range(today)
	rows, err := s.Store.Ranking(ctx, user, from, to, 0)
	if err != nil {
		return 0, err
	}
	for _, r := range rows {
		if r.UserID == user {
			return r.Points, nil
		}
	}
	return 0, nil
}
