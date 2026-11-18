package main

// Ligacao dos modulos de quem estuda: estudo (learning), progresso (progress) e os direitos sobre os dados (privacy).
// Os tres descobrem quem pede pela sessao do modulo de identidade, e os modulos nunca se importam.

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	identityhttp "github.com/keven-rdr/wordloop/api/internal/identity/adapters/http"
	identity "github.com/keven-rdr/wordloop/api/internal/identity/app"
	identitydomain "github.com/keven-rdr/wordloop/api/internal/identity/domain"
	learninghttp "github.com/keven-rdr/wordloop/api/internal/learning/adapters/http"
	learningpg "github.com/keven-rdr/wordloop/api/internal/learning/adapters/postgres"
	learning "github.com/keven-rdr/wordloop/api/internal/learning/app"
	learningdomain "github.com/keven-rdr/wordloop/api/internal/learning/domain"
	privacyhttp "github.com/keven-rdr/wordloop/api/internal/privacy/adapters/http"
	privacypg "github.com/keven-rdr/wordloop/api/internal/privacy/adapters/postgres"
	privacy "github.com/keven-rdr/wordloop/api/internal/privacy/app"
	privacydomain "github.com/keven-rdr/wordloop/api/internal/privacy/domain"
	progresshttp "github.com/keven-rdr/wordloop/api/internal/progress/adapters/http"
	progresspg "github.com/keven-rdr/wordloop/api/internal/progress/adapters/postgres"
	progress "github.com/keven-rdr/wordloop/api/internal/progress/app"
	progressdomain "github.com/keven-rdr/wordloop/api/internal/progress/domain"
	"github.com/keven-rdr/wordloop/api/internal/srs"
)

type principals func(context.Context) (identitydomain.Principal, error)

// learnerViewers descobre quem estuda.
type learnerViewers struct{ who principals }

func (v learnerViewers) Viewer(ctx context.Context) (learningdomain.Viewer, error) {
	p, err := v.who(ctx)
	if errors.Is(err, identity.ErrNoSession) {
		return learningdomain.Viewer{}, learning.ErrNoSession
	}
	if err != nil {
		return learningdomain.Viewer{}, err
	}
	return learningdomain.Viewer{UserID: p.User.ID, Admin: p.User.Role == identitydomain.RoleAdmin}, nil
}

func buildLearning(pool *pgxpool.Pool, who identityhttp.Handler) learninghttp.Handler {
	repo := learningpg.NewStore(pool)
	viewers := learnerViewers{who: who.Principal}
	sched := srs.SimpleV1{}
	return learninghttp.Handler{
		Plan:     learning.PlanSession{Viewers: viewers, Repo: repo, Scheduler: sched, NewID: uuid.New},
		Submit:   learning.SubmitReviews{Viewers: viewers, Repo: repo, Scheduler: sched},
		Settings: learning.SettingsService{Viewers: viewers, Repo: repo},
	}
}

// progressViewers descobre quem pergunta pelo progresso.
type progressViewers struct{ who principals }

func (v progressViewers) Who(ctx context.Context) (progressdomain.Viewer, error) {
	p, err := v.who(ctx)
	if errors.Is(err, identity.ErrNoSession) {
		return progressdomain.Viewer{}, progress.ErrNoSession
	}
	if err != nil {
		return progressdomain.Viewer{}, err
	}
	return progressdomain.Viewer{UserID: p.User.ID}, nil
}

func buildProgress(pool *pgxpool.Pool, who identityhttp.Handler) progresshttp.Handler {
	reader := progresspg.NewStore(pool)
	identity := progressViewers{who: who.Principal}
	return progresshttp.Handler{
		Dashboard: progress.GetDashboard{Identity: identity, Reader: reader},
		Heatmap:   progress.GetHeatmap{Identity: identity, Reader: reader},
		Goals:     progress.GoalService{Identity: identity, Reader: reader, Goals: reader},
	}
}

// privacyIdentity descobre quem pede a exportacao ou a exclusao dos dados, e desde quando a sessao esta autenticada.
type privacyIdentity struct{ who principals }

func (v privacyIdentity) Who(ctx context.Context) (privacydomain.Subject, error) {
	p, err := v.who(ctx)
	if errors.Is(err, identity.ErrNoSession) {
		return privacydomain.Subject{}, privacy.ErrNoSession
	}
	if err != nil {
		return privacydomain.Subject{}, err
	}
	return privacydomain.Subject{
		UserID: p.User.ID, Admin: p.User.Role == identitydomain.RoleAdmin, AuthenticatedAt: p.AuthenticatedAt,
	}, nil
}

func buildPrivacy(pool *pgxpool.Pool, who identityhttp.Handler) privacyhttp.Handler {
	return privacyhttp.Handler{Service: privacy.Service{
		Identity: privacyIdentity{who: who.Principal}, Store: privacypg.NewStore(pool),
	}}
}
