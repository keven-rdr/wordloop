package main

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	identityhttp "github.com/keven-rdr/wordloop/api/internal/identity/adapters/http"
	identity "github.com/keven-rdr/wordloop/api/internal/identity/app"
	identitydomain "github.com/keven-rdr/wordloop/api/internal/identity/domain"
	learning "github.com/keven-rdr/wordloop/api/internal/learning/app"
	privacy "github.com/keven-rdr/wordloop/api/internal/privacy/app"
	progress "github.com/keven-rdr/wordloop/api/internal/progress/app"
)

var (
	someone   = uuid.MustParse("00000000-0000-0000-0000-0000000000aa")
	loggedAt  = time.Date(2026, 11, 18, 15, 0, 0, 0, time.UTC)
	errBroken = errors.New("falha simulada")
)

func principal(role identitydomain.Role) principals {
	return func(context.Context) (identitydomain.Principal, error) {
		return identitydomain.Principal{User: identitydomain.User{ID: someone, Role: role}, AuthenticatedAt: loggedAt}, nil
	}
}

func failing(err error) principals {
	return func(context.Context) (identitydomain.Principal, error) { return identitydomain.Principal{}, err }
}

func TestLearnerViewersMapTheSessionToWhoStudies(t *testing.T) {
	got, err := learnerViewers{who: principal(identitydomain.RoleUser)}.Viewer(context.Background())
	if err != nil || got.UserID != someone || got.Admin {
		t.Errorf("aluno: %+v %v", got, err)
	}
	got, err = learnerViewers{who: principal(identitydomain.RoleAdmin)}.Viewer(context.Background())
	if err != nil || !got.Admin {
		t.Errorf("admin: %+v %v", got, err)
	}
	if _, err := (learnerViewers{who: failing(identity.ErrNoSession)}).Viewer(context.Background()); !errors.Is(err, learning.ErrNoSession) {
		t.Errorf("sem sessao vira o erro do modulo: %v", err)
	}
	if _, err := (learnerViewers{who: failing(errBroken)}).Viewer(context.Background()); !errors.Is(err, errBroken) {
		t.Errorf("outro erro passa direto: %v", err)
	}
}

func TestProgressViewersMapTheSessionToWhoAsks(t *testing.T) {
	got, err := progressViewers{who: principal(identitydomain.RoleUser)}.Who(context.Background())
	if err != nil || got.UserID != someone {
		t.Errorf("pergunta: %+v %v", got, err)
	}
	if _, err := (progressViewers{who: failing(identity.ErrNoSession)}).Who(context.Background()); !errors.Is(err, progress.ErrNoSession) {
		t.Errorf("sem sessao vira o erro do modulo: %v", err)
	}
	if _, err := (progressViewers{who: failing(errBroken)}).Who(context.Background()); !errors.Is(err, errBroken) {
		t.Errorf("outro erro passa direto: %v", err)
	}
}

func TestPrivacyIdentityCarriesTheRoleAndWhenTheSessionWasAuthenticated(t *testing.T) {
	got, err := privacyIdentity{who: principal(identitydomain.RoleUser)}.Who(context.Background())
	if err != nil || got.UserID != someone || got.Admin || !got.AuthenticatedAt.Equal(loggedAt) {
		t.Errorf("aluno: %+v %v", got, err)
	}
	got, err = privacyIdentity{who: principal(identitydomain.RoleAdmin)}.Who(context.Background())
	if err != nil || !got.Admin {
		t.Errorf("admin: %+v %v", got, err)
	}
	if _, err := (privacyIdentity{who: failing(identity.ErrNoSession)}).Who(context.Background()); !errors.Is(err, privacy.ErrNoSession) {
		t.Errorf("sem sessao vira o erro do modulo: %v", err)
	}
	if _, err := (privacyIdentity{who: failing(errBroken)}).Who(context.Background()); !errors.Is(err, errBroken) {
		t.Errorf("outro erro passa direto: %v", err)
	}
}

// A ligacao monta cada modulo com o repositorio e a identidade; sem banco real basta ver que nada fica solto.
func TestTheModulesAreWiredToTheirRepositoryAndTheSession(t *testing.T) {
	who := identityhttp.Handler{}
	if l := buildLearning(nil, who); l.Plan.Repo == nil || l.Submit.Repo == nil || l.Settings.Repo == nil || l.Plan.Viewers == nil {
		t.Errorf("estudo: %+v", l)
	}
	if p := buildProgress(nil, who); p.Dashboard.Reader == nil || p.Heatmap.Identity == nil || p.Goals.Goals == nil {
		t.Errorf("progresso: %+v", p)
	}
	if p := buildPrivacy(nil, who); p.Service.Store == nil || p.Service.Identity == nil {
		t.Errorf("privacidade: %+v", p)
	}
}
