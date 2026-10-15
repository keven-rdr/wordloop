package httpapi

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/keven-rdr/wordloop/api/internal/api/gen"
	"github.com/keven-rdr/wordloop/api/internal/identity/app"
	"github.com/keven-rdr/wordloop/api/internal/identity/passwords"
	"github.com/keven-rdr/wordloop/api/internal/platform/httpx"
)

type Passwords struct {
	Login passwords.Login
	Set   passwords.SetPassword
}

func (r statusReply) VisitLoginWithPasswordResponse(w http.ResponseWriter) error { return r.write(w) }
func (r statusReply) VisitSetPasswordResponse(w http.ResponseWriter) error       { return r.write(w) }

func (h Handler) LoginWithPassword(ctx context.Context, req gen.LoginWithPasswordRequestObject) (gen.LoginWithPasswordResponseObject, error) {
	if h.Passwords == nil {
		return gen.LoginWithPassword503Response{}, nil
	}
	if req.Body == nil {
		return statusReply{status: http.StatusBadRequest}, nil
	}
	opened, err := h.Passwords.Login.Run(ctx, passwords.LoginInput{
		Email: req.Body.Email, Password: req.Body.Password, UserAgent: httpx.UserAgent(ctx),
	})
	switch {
	case errors.Is(err, passwords.ErrInvalidCredentials):
		slog.Info("login por senha recusado")
		return statusReply{status: http.StatusUnauthorized}, nil
	case errors.Is(err, app.ErrMFARequired):
		slog.Warn("login por senha recusado: a conta exige 2 fatores")
		return statusReply{status: http.StatusForbidden}, nil
	case err != nil:
		return nil, err
	}
	return statusReply{status: http.StatusNoContent, cookies: []*http.Cookie{h.cookies().session(opened.SessionID, opened.ExpiresAt)}}, nil
}

func (h Handler) SetPassword(ctx context.Context, req gen.SetPasswordRequestObject) (gen.SetPasswordResponseObject, error) {
	if h.Passwords == nil {
		return gen.SetPassword503Response{}, nil
	}
	if req.Body == nil {
		return statusReply{status: http.StatusBadRequest}, nil
	}
	opened, err := h.Passwords.Set.Run(ctx, passwords.SetInput{
		Token: req.Body.Token, Password: req.Body.Password, UserAgent: httpx.UserAgent(ctx),
	})
	switch {
	case errors.Is(err, passwords.ErrUnauthorized):
		return statusReply{status: http.StatusUnauthorized}, nil
	case errors.Is(err, passwords.ErrWeak):
		return statusReply{status: http.StatusUnprocessableEntity}, nil
	case err != nil:
		return nil, err
	}
	if opened == nil {
		return statusReply{status: http.StatusNoContent}, nil
	}
	return statusReply{status: http.StatusNoContent, cookies: []*http.Cookie{h.cookies().session(opened.SessionID, opened.ExpiresAt)}}, nil
}
