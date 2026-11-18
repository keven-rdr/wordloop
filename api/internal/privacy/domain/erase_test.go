package domain_test

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/keven-rdr/wordloop/api/internal/privacy/domain"
)

var now = time.Date(2026, 11, 18, 15, 0, 0, 0, time.UTC)

func subject(loggedMinutesAgo int, admin bool) domain.Subject {
	return domain.Subject{UserID: uuid.New(), Admin: admin, AuthenticatedAt: now.Add(-time.Duration(loggedMinutesAgo) * time.Minute)}
}

func TestCheckEraseAllowsTheRightPersonAtTheRightTime(t *testing.T) {
	if err := domain.CheckErase(subject(3, false), "ana@example.com", "ana@example.com", now); err != nil {
		t.Errorf("e-mail certo e login recente: %v", err)
	}
}

func TestCheckEraseComparesTheEmailWithoutCaseAndWithoutSpaces(t *testing.T) {
	for _, typed := range []string{"ANA@Example.com", "  ana@example.com  ", "Ana@EXAMPLE.COM\n"} {
		if err := domain.CheckErase(subject(1, false), "ana@example.com", typed, now); err != nil {
			t.Errorf("%q deveria valer: %v", typed, err)
		}
	}
}

func TestCheckEraseRefusesWhatCannotBeDone(t *testing.T) {
	cases := map[string]struct {
		who          domain.Subject
		email, typed string
		want         error
	}{
		"confirmacao vazia":                {subject(1, false), "ana@example.com", "", domain.ErrWrongConfirmation},
		"confirmacao de outro e-mail":      {subject(1, false), "ana@example.com", "bia@example.com", domain.ErrWrongConfirmation},
		"parte do e-mail":                  {subject(1, false), "ana@example.com", "ana", domain.ErrWrongConfirmation},
		"conta sem e-mail nao se confirma": {subject(1, false), "", "", domain.ErrWrongConfirmation},
		"administrador":                    {subject(1, true), "ana@example.com", "ana@example.com", domain.ErrAdminAccount},
		"login de 16 minutos":              {subject(16, false), "ana@example.com", "ana@example.com", domain.ErrRecentLoginRequired},
		"login de ontem":                   {subject(24*60, false), "ana@example.com", "ana@example.com", domain.ErrRecentLoginRequired},
	}
	for name, c := range cases {
		if err := domain.CheckErase(c.who, c.email, c.typed, now); !errors.Is(err, c.want) {
			t.Errorf("%s: %v, esperava %v", name, err, c.want)
		}
	}
}

func TestCheckEraseLimitIsFifteenMinutesInclusive(t *testing.T) {
	if err := domain.CheckErase(subject(15, false), "a@b.co", "a@b.co", now); err != nil {
		t.Errorf("15 minutos ainda vale: %v", err)
	}
	later := now.Add(time.Second)
	if err := domain.CheckErase(subject(15, false), "a@b.co", "a@b.co", later); !errors.Is(err, domain.ErrRecentLoginRequired) {
		t.Errorf("15 minutos e 1 segundo nao vale: %v", err)
	}
}

// Quem nao pode nunca pode, mesmo digitando certo e com login recente; e o aviso certo vem antes do outro.
func TestCheckEraseReportsTheMostBasicProblemFirst(t *testing.T) {
	if err := domain.CheckErase(subject(60, true), "a@b.co", "errado", now); !errors.Is(err, domain.ErrAdminAccount) {
		t.Errorf("admin antes de tudo: %v", err)
	}
	if err := domain.CheckErase(subject(60, false), "a@b.co", "errado", now); !errors.Is(err, domain.ErrWrongConfirmation) {
		t.Errorf("texto errado antes do login antigo: %v", err)
	}
}
