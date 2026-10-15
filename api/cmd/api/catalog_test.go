package main

import (
	"context"
	"errors"
	"testing"

	"github.com/keven-rdr/wordloop/api/internal/identity/app"
	"github.com/keven-rdr/wordloop/api/internal/identity/domain"
)

func TestSessionViewersMapsThePrincipalToWhoIsAsking(t *testing.T) {
	cases := map[string]struct {
		who       func(context.Context) (domain.Principal, error)
		signedIn  bool
		admin     bool
		wantError bool
	}{
		"sem sessao": {who: func(context.Context) (domain.Principal, error) { return domain.Principal{}, app.ErrNoSession }},
		"aluno": {who: func(context.Context) (domain.Principal, error) {
			return domain.Principal{User: domain.User{Role: domain.RoleUser}}, nil
		}, signedIn: true},
		"admin": {who: func(context.Context) (domain.Principal, error) {
			return domain.Principal{User: domain.User{Role: domain.RoleAdmin}}, nil
		}, signedIn: true, admin: true},
		"erro do banco": {who: func(context.Context) (domain.Principal, error) { return domain.Principal{}, errors.New("banco fora") }, wantError: true},
	}
	for name, c := range cases {
		v, err := sessionViewers{who: c.who}.Viewer(context.Background())
		if (err != nil) != c.wantError || v.Authenticated != c.signedIn || v.Admin != c.admin {
			t.Errorf("%s: %+v %v", name, v, err)
		}
	}
}
