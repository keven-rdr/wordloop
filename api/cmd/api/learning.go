package main

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
	"github.com/keven-rdr/wordloop/api/internal/srs"
)

// learnerViewers descobre quem estuda pela sessao do modulo de identidade; os modulos nunca se importam.
type learnerViewers struct {
	who func(context.Context) (identitydomain.Principal, error)
}

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
		Plan:   learning.PlanSession{Viewers: viewers, Repo: repo, Scheduler: sched, NewID: uuid.New},
		Submit: learning.SubmitReviews{Viewers: viewers, Repo: repo, Scheduler: sched},
	}
}
