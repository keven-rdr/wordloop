package app

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
)

// StartLogin prepara o login: gera state, nonce e verifier PKCE e devolve a URL do provedor.
type StartLogin struct {
	Provider Provider
	Sealer   Sealer
	Rand     io.Reader // nil = crypto/rand
}

// LoginRedirect e para onde mandar o navegador e o cookie de estado a gravar.
type LoginRedirect struct {
	URL        string
	FlowCookie string
}

// Run inicia o login. returnTo so vale se for caminho relativo.
func (u StartLogin) Run(ctx context.Context, returnTo string) (LoginRedirect, error) {
	var tokens [3]string
	for i := range tokens {
		t, err := randomToken(u.Rand, 32)
		if err != nil {
			return LoginRedirect{}, err
		}
		tokens[i] = t
	}
	flow := Flow{State: tokens[0], Nonce: tokens[1], Verifier: tokens[2], ReturnTo: safeReturnTo(returnTo)}
	authURL, err := u.Provider.AuthURL(ctx, flow.State, flow.Nonce, flow.Verifier)
	if err != nil {
		return LoginRedirect{}, fmt.Errorf("montar a URL do provedor: %w", err)
	}
	raw, err := json.Marshal(flow)
	if err != nil {
		return LoginRedirect{}, err
	}
	sealed, err := u.Sealer.Seal(raw)
	if err != nil {
		return LoginRedirect{}, fmt.Errorf("selar o estado do login: %w", err)
	}
	return LoginRedirect{URL: authURL, FlowCookie: sealed}, nil
}
