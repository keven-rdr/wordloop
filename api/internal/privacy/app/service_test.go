package app_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/keven-rdr/wordloop/api/internal/privacy/app"
	"github.com/keven-rdr/wordloop/api/internal/privacy/domain"
	"github.com/keven-rdr/wordloop/api/internal/privacy/privacytest"
)

var (
	now  = time.Date(2026, 11, 18, 15, 0, 0, 0, time.UTC)
	user = uuid.MustParse("00000000-0000-0000-0000-0000000000aa")
)

func service(s *privacytest.Store, who domain.Subject) app.Service {
	return app.Service{Identity: &privacytest.Identity{Is: who}, Store: s, Now: func() time.Time { return now }}
}

func recent() domain.Subject {
	return domain.Subject{UserID: user, AuthenticatedAt: now.Add(-2 * time.Minute)}
}

func withAccount() *privacytest.Store {
	s := privacytest.New()
	s.Emails[user] = "ana@example.com"
	return s
}

func TestExportReturnsTheDataAndWhenItWasGenerated(t *testing.T) {
	s := withAccount()
	s.Data = []byte(`{"profile":{"display_name":"Ana"}}`)
	out, err := service(s, recent()).Export(context.Background())
	if err != nil || string(out.Data) != `{"profile":{"display_name":"Ana"}}` || !out.GeneratedAt.Equal(now) {
		t.Errorf("exportacao: %+v %v", out, err)
	}
}

func TestEraseDeletesTheAccountOfWhoAsksWhenEverythingChecksOut(t *testing.T) {
	s := withAccount()
	s.Reviews = 42
	report, err := service(s, recent()).Erase(context.Background(), "Ana@Example.com")
	if err != nil || report.Reviews != 42 {
		t.Fatalf("apagar: %+v %v", report, err)
	}
	if len(s.Erased) != 1 || s.Erased[0] != user {
		t.Errorf("so a conta de quem pede: %v", s.Erased)
	}
}

func TestEraseDoesNotTouchAnythingWhenARuleFails(t *testing.T) {
	cases := map[string]struct {
		who   domain.Subject
		typed string
		want  error
	}{
		"confirmacao errada": {recent(), "bia@example.com", domain.ErrWrongConfirmation},
		"administrador":      {domain.Subject{UserID: user, Admin: true, AuthenticatedAt: now}, "ana@example.com", domain.ErrAdminAccount},
		"login antigo":       {domain.Subject{UserID: user, AuthenticatedAt: now.Add(-time.Hour)}, "ana@example.com", domain.ErrRecentLoginRequired},
	}
	for name, c := range cases {
		s := withAccount()
		if _, err := service(s, c.who).Erase(context.Background(), c.typed); !errors.Is(err, c.want) {
			t.Errorf("%s: %v", name, err)
		}
		if len(s.Erased) != 0 {
			t.Errorf("%s: nada deve ser apagado", name)
		}
	}
}

func TestPrivacyNeedsASessionAndPropagatesFailures(t *testing.T) {
	anonymous := service(withAccount(), domain.Subject{})
	if _, err := anonymous.Export(context.Background()); !errors.Is(err, app.ErrNoSession) {
		t.Errorf("exportar sem sessao: %v", err)
	}
	if _, err := anonymous.Erase(context.Background(), "ana@example.com"); !errors.Is(err, app.ErrNoSession) {
		t.Errorf("apagar sem sessao: %v", err)
	}
	for _, method := range []string{"Email", "Erase"} {
		s := withAccount()
		s.FailOn = method
		if _, err := service(s, recent()).Erase(context.Background(), "ana@example.com"); !errors.Is(err, privacytest.ErrBoom) {
			t.Errorf("falha em %s deveria subir: %v", method, err)
		}
	}
	s := withAccount()
	s.FailOn = "Export"
	if _, err := service(s, recent()).Export(context.Background()); !errors.Is(err, privacytest.ErrBoom) {
		t.Errorf("falha ao exportar deveria subir: %v", err)
	}
}

func TestEraseOfAnAccountThatNoLongerExistsIsNotFound(t *testing.T) {
	if _, err := service(privacytest.New(), recent()).Erase(context.Background(), "ana@example.com"); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("conta inexistente: %v", err)
	}
}
