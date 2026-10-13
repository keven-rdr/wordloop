package api

import (
	"context"

	"github.com/keven-rdr/wordloop/api/internal/api/gen"
)

// Identity e o modulo de login; quem o implementa fica em identity/adapters/http e so o cmd/api os conecta.
type Identity interface {
	StartLogin(ctx context.Context, req gen.StartLoginRequestObject) (gen.StartLoginResponseObject, error)
	FinishLogin(ctx context.Context, req gen.FinishLoginRequestObject) (gen.FinishLoginResponseObject, error)
	Logout(ctx context.Context, req gen.LogoutRequestObject) (gen.LogoutResponseObject, error)
	GetMe(ctx context.Context, req gen.GetMeRequestObject) (gen.GetMeResponseObject, error)
}

// Sem modulo de identidade ligado (testes de outras rotas): login indisponivel e ninguem logado.

// StartLogin delega ao modulo de identidade.
func (s Server) StartLogin(ctx context.Context, req gen.StartLoginRequestObject) (gen.StartLoginResponseObject, error) {
	if s.Identity == nil {
		return gen.StartLogin503Response{}, nil
	}
	return s.Identity.StartLogin(ctx, req)
}

// FinishLogin delega ao modulo de identidade.
func (s Server) FinishLogin(ctx context.Context, req gen.FinishLoginRequestObject) (gen.FinishLoginResponseObject, error) {
	if s.Identity == nil {
		return gen.FinishLogin503Response{}, nil
	}
	return s.Identity.FinishLogin(ctx, req)
}

// Logout delega ao modulo de identidade.
func (s Server) Logout(ctx context.Context, req gen.LogoutRequestObject) (gen.LogoutResponseObject, error) {
	if s.Identity == nil {
		return gen.Logout204Response{}, nil
	}
	return s.Identity.Logout(ctx, req)
}

// GetMe delega ao modulo de identidade.
func (s Server) GetMe(ctx context.Context, req gen.GetMeRequestObject) (gen.GetMeResponseObject, error) {
	if s.Identity == nil {
		return gen.GetMe401Response{}, nil
	}
	return s.Identity.GetMe(ctx, req)
}
