package app

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"io"
	"time"
)

// DefaultSessionTTL e a duracao da sessao.
const DefaultSessionTTL = 30 * 24 * time.Hour

// FinishLogin conclui o login: valida o estado, troca o code, cria/atualiza o perfil e abre a sessao.
type FinishLogin struct {
	Provider   Provider
	Users      Users
	Sessions   Sessions
	Sealer     Sealer
	Rand       io.Reader        // nil = crypto/rand
	Now        func() time.Time // injetado nos testes
	SessionTTL time.Duration    // 0 = DefaultSessionTTL
}

// FinishInput vem da query do callback e do cookie de estado.
type FinishInput struct {
	Code, State, ProviderError string
	FlowCookie                 string
	UserAgent                  string
}

// LoggedIn e a sessao aberta: o id vai ao cookie; ReturnTo e para onde voltar.
type LoggedIn struct {
	SessionID string
	ReturnTo  string
	ExpiresAt time.Time
}

// Run valida e abre a sessao; qualquer problema com o pedido vira ErrInvalidLogin.
func (u FinishLogin) Run(ctx context.Context, in FinishInput) (LoggedIn, error) {
	flow, err := u.readFlow(in)
	if err != nil {
		return LoggedIn{}, err
	}
	tokens, err := u.Provider.Exchange(ctx, in.Code, flow.Verifier, flow.Nonce)
	if err != nil {
		return LoggedIn{}, fmt.Errorf("%w: trocar o code: %w", ErrInvalidLogin, err)
	}
	user, err := u.Users.Upsert(ctx, tokens.Identity)
	if err != nil {
		return LoggedIn{}, fmt.Errorf("gravar o perfil: %w", err)
	}
	return u.openSession(ctx, user.ID, tokens.RefreshToken, in.UserAgent, flow.ReturnTo)
}

func (u FinishLogin) readFlow(in FinishInput) (Flow, error) {
	if in.ProviderError != "" || in.Code == "" || in.State == "" || in.FlowCookie == "" {
		return Flow{}, ErrInvalidLogin
	}
	raw, err := u.Sealer.Open(in.FlowCookie)
	if err != nil {
		return Flow{}, ErrInvalidLogin
	}
	var flow Flow
	if err := json.Unmarshal(raw, &flow); err != nil || subtle.ConstantTimeCompare([]byte(flow.State), []byte(in.State)) != 1 {
		return Flow{}, ErrInvalidLogin
	}
	return flow, nil
}
