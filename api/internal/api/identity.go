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
