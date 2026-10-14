package api

import (
	"context"

	"github.com/keven-rdr/wordloop/api/internal/api/gen"
)

type Identity interface {
	StartLogin(ctx context.Context, req gen.StartLoginRequestObject) (gen.StartLoginResponseObject, error)
	FinishLogin(ctx context.Context, req gen.FinishLoginRequestObject) (gen.FinishLoginResponseObject, error)
	Logout(ctx context.Context, req gen.LogoutRequestObject) (gen.LogoutResponseObject, error)
	GetMe(ctx context.Context, req gen.GetMeRequestObject) (gen.GetMeResponseObject, error)
	ListSessions(ctx context.Context, req gen.ListSessionsRequestObject) (gen.ListSessionsResponseObject, error)
	RevokeSession(ctx context.Context, req gen.RevokeSessionRequestObject) (gen.RevokeSessionResponseObject, error)
	RevokeAllSessions(ctx context.Context, req gen.RevokeAllSessionsRequestObject) (gen.RevokeAllSessionsResponseObject, error)
	BeginPasskeyRegistration(ctx context.Context, req gen.BeginPasskeyRegistrationRequestObject) (gen.BeginPasskeyRegistrationResponseObject, error)
	FinishPasskeyRegistration(ctx context.Context, req gen.FinishPasskeyRegistrationRequestObject) (gen.FinishPasskeyRegistrationResponseObject, error)
	BeginPasskeyLogin(ctx context.Context, req gen.BeginPasskeyLoginRequestObject) (gen.BeginPasskeyLoginResponseObject, error)
	FinishPasskeyLogin(ctx context.Context, req gen.FinishPasskeyLoginRequestObject) (gen.FinishPasskeyLoginResponseObject, error)
	ListPasskeys(ctx context.Context, req gen.ListPasskeysRequestObject) (gen.ListPasskeysResponseObject, error)
	DeletePasskey(ctx context.Context, req gen.DeletePasskeyRequestObject) (gen.DeletePasskeyResponseObject, error)
}

func (s Server) StartLogin(ctx context.Context, req gen.StartLoginRequestObject) (gen.StartLoginResponseObject, error) {
	if s.Identity == nil {
		return gen.StartLogin503Response{}, nil
	}
	return s.Identity.StartLogin(ctx, req)
}

func (s Server) FinishLogin(ctx context.Context, req gen.FinishLoginRequestObject) (gen.FinishLoginResponseObject, error) {
	if s.Identity == nil {
		return gen.FinishLogin503Response{}, nil
	}
	return s.Identity.FinishLogin(ctx, req)
}

func (s Server) Logout(ctx context.Context, req gen.LogoutRequestObject) (gen.LogoutResponseObject, error) {
	if s.Identity == nil {
		return gen.Logout204Response{}, nil
	}
	return s.Identity.Logout(ctx, req)
}

func (s Server) GetMe(ctx context.Context, req gen.GetMeRequestObject) (gen.GetMeResponseObject, error) {
	if s.Identity == nil {
		return gen.GetMe401Response{}, nil
	}
	return s.Identity.GetMe(ctx, req)
}

func (s Server) ListSessions(ctx context.Context, req gen.ListSessionsRequestObject) (gen.ListSessionsResponseObject, error) {
	if s.Identity == nil {
		return gen.ListSessions401Response{}, nil
	}
	return s.Identity.ListSessions(ctx, req)
}

func (s Server) RevokeSession(ctx context.Context, req gen.RevokeSessionRequestObject) (gen.RevokeSessionResponseObject, error) {
	if s.Identity == nil {
		return gen.RevokeSession401Response{}, nil
	}
	return s.Identity.RevokeSession(ctx, req)
}

func (s Server) RevokeAllSessions(ctx context.Context, req gen.RevokeAllSessionsRequestObject) (gen.RevokeAllSessionsResponseObject, error) {
	if s.Identity == nil {
		return gen.RevokeAllSessions401Response{}, nil
	}
	return s.Identity.RevokeAllSessions(ctx, req)
}

func (s Server) BeginPasskeyRegistration(ctx context.Context, req gen.BeginPasskeyRegistrationRequestObject) (gen.BeginPasskeyRegistrationResponseObject, error) {
	if s.Identity == nil {
		return gen.BeginPasskeyRegistration503Response{}, nil
	}
	return s.Identity.BeginPasskeyRegistration(ctx, req)
}

func (s Server) FinishPasskeyRegistration(ctx context.Context, req gen.FinishPasskeyRegistrationRequestObject) (gen.FinishPasskeyRegistrationResponseObject, error) {
	if s.Identity == nil {
		return gen.FinishPasskeyRegistration503Response{}, nil
	}
	return s.Identity.FinishPasskeyRegistration(ctx, req)
}

func (s Server) BeginPasskeyLogin(ctx context.Context, req gen.BeginPasskeyLoginRequestObject) (gen.BeginPasskeyLoginResponseObject, error) {
	if s.Identity == nil {
		return gen.BeginPasskeyLogin503Response{}, nil
	}
	return s.Identity.BeginPasskeyLogin(ctx, req)
}

func (s Server) FinishPasskeyLogin(ctx context.Context, req gen.FinishPasskeyLoginRequestObject) (gen.FinishPasskeyLoginResponseObject, error) {
	if s.Identity == nil {
		return gen.FinishPasskeyLogin503Response{}, nil
	}
	return s.Identity.FinishPasskeyLogin(ctx, req)
}

func (s Server) ListPasskeys(ctx context.Context, req gen.ListPasskeysRequestObject) (gen.ListPasskeysResponseObject, error) {
	if s.Identity == nil {
		return gen.ListPasskeys401Response{}, nil
	}
	return s.Identity.ListPasskeys(ctx, req)
}

func (s Server) DeletePasskey(ctx context.Context, req gen.DeletePasskeyRequestObject) (gen.DeletePasskeyResponseObject, error) {
	if s.Identity == nil {
		return gen.DeletePasskey401Response{}, nil
	}
	return s.Identity.DeletePasskey(ctx, req)
}
