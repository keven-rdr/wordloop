package main

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgxpool"

	catalogapi "github.com/keven-rdr/wordloop/api/internal/catalog/adapters/http"
	catalogpg "github.com/keven-rdr/wordloop/api/internal/catalog/adapters/postgres"
	catalog "github.com/keven-rdr/wordloop/api/internal/catalog/app"
	catalogdomain "github.com/keven-rdr/wordloop/api/internal/catalog/domain"
	identityhttp "github.com/keven-rdr/wordloop/api/internal/identity/adapters/http"
	identity "github.com/keven-rdr/wordloop/api/internal/identity/app"
	identitydomain "github.com/keven-rdr/wordloop/api/internal/identity/domain"
)

// sessionViewers descobre quem pergunta ao catalogo pela sessao do modulo de identidade.
// E aqui, na composicao, que um modulo conhece o outro: eles nunca se importam.
type sessionViewers struct {
	who func(context.Context) (identitydomain.Principal, error)
}

func (v sessionViewers) Viewer(ctx context.Context) (catalogdomain.Viewer, error) {
	p, err := v.who(ctx)
	if errors.Is(err, identity.ErrNoSession) {
		return catalogdomain.Viewer{}, nil
	}
	if err != nil {
		return catalogdomain.Viewer{}, err
	}
	return catalogdomain.Viewer{Authenticated: true, Admin: p.User.Role == identitydomain.RoleAdmin}, nil
}

func buildCatalog(pool *pgxpool.Pool, who identityhttp.Handler) catalogapi.Handler {
	repo := catalogpg.NewCatalog(pool)
	viewers := sessionViewers{who: who.Principal}
	return catalogapi.Handler{
		Units:   catalog.ListUnits{Viewers: viewers, Repo: repo},
		Items:   catalog.ListItems{Viewers: viewers, Repo: repo},
		Item:    catalog.GetItem{Viewers: viewers, Repo: repo},
		Credits: catalog.ListCredits{Repo: repo},
	}
}
