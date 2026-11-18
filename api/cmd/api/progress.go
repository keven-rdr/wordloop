package main

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgxpool"

	identityhttp "github.com/keven-rdr/wordloop/api/internal/identity/adapters/http"
	identity "github.com/keven-rdr/wordloop/api/internal/identity/app"
	identitydomain "github.com/keven-rdr/wordloop/api/internal/identity/domain"
	progresshttp "github.com/keven-rdr/wordloop/api/internal/progress/adapters/http"
	progresspg "github.com/keven-rdr/wordloop/api/internal/progress/adapters/postgres"
	progress "github.com/keven-rdr/wordloop/api/internal/progress/app"
	progressdomain "github.com/keven-rdr/wordloop/api/internal/progress/domain"
)

// progressViewers descobre quem pergunta pela sessao do modulo de identidade; os modulos nunca se importam.
type progressViewers struct {
	who func(context.Context) (identitydomain.Principal, error)
}

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
	}
}
