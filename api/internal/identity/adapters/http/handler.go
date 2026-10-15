package httpapi

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/keven-rdr/wordloop/api/internal/api/gen"
	"github.com/keven-rdr/wordloop/api/internal/identity/accounts"
	"github.com/keven-rdr/wordloop/api/internal/identity/app"
	"github.com/keven-rdr/wordloop/api/internal/platform/httpx"
)

type Handler struct {
	Start        *app.StartLogin
	Finish       *app.FinishLogin
	Auth         app.Authenticate
	Out          app.Logout
	List         app.ListSessions
	Revoke       app.RevokeSession
	RevokeAll    app.RevokeAllSessions
	Passkeys     *Passkeys
	Passwords    *Passwords
	Registration *Registration
	TwoFactor    accounts.TwoFactor
	Secure       bool
}

func (h Handler) cookies() cookies { return cookies{secure: h.Secure} }

func (h Handler) sessionID(ctx context.Context) string {
	return httpx.Cookie(ctx, h.cookies().sessionName())
}

func (h Handler) enabled() bool { return h.Start != nil && h.Finish != nil }

func (h Handler) StartLogin(ctx context.Context, req gen.StartLoginRequestObject) (gen.StartLoginResponseObject, error) {
	if !h.enabled() {
		return gen.StartLogin503Response{}, nil
	}
	returnTo := ""
	if req.Params.ReturnTo != nil {
		returnTo = *req.Params.ReturnTo
	}
	login, err := h.Start.Run(ctx, returnTo)
	if err != nil {
		slog.Warn("login indisponivel", "err", err)
		return gen.StartLogin503Response{}, nil
	}
	return redirect{location: login.URL, cookies: []*http.Cookie{h.cookies().flow(login.FlowCookie)}}, nil
}

func (h Handler) FinishLogin(ctx context.Context, req gen.FinishLoginRequestObject) (gen.FinishLoginResponseObject, error) {
	if !h.enabled() {
		return gen.FinishLogin503Response{}, nil
	}
	c := h.cookies()
	in := app.FinishInput{
		Code: deref(req.Params.Code), State: deref(req.Params.State), ProviderError: deref(req.Params.Error),
		FlowCookie: httpx.Cookie(ctx, c.flowName()), UserAgent: httpx.UserAgent(ctx),
	}
	done, err := h.Finish.Run(ctx, in)
	if errors.Is(err, app.ErrInvalidLogin) {
		slog.Info("login recusado", "err", err)
		return badLogin{cookies: []*http.Cookie{c.clearFlow()}}, nil
	}
	if errors.Is(err, app.ErrMFARequired) {
		slog.Warn("login recusado: a conta exige 2 fatores e o provedor nao os comprovou")
		return mfaRequired{cookies: []*http.Cookie{c.clearFlow()}}, nil
	}
	if err != nil {
		return nil, err
	}
	return redirect{location: done.ReturnTo, cookies: []*http.Cookie{c.session(done.SessionID, done.ExpiresAt), c.clearFlow()}}, nil
}

func (h Handler) Logout(ctx context.Context, _ gen.LogoutRequestObject) (gen.LogoutResponseObject, error) {
	if err := h.Out.Run(ctx, h.sessionID(ctx)); err != nil {
		return nil, err
	}
	return loggedOut{cookies: []*http.Cookie{h.cookies().clearSession()}}, nil
}

func (h Handler) GetMe(ctx context.Context, _ gen.GetMeRequestObject) (gen.GetMeResponseObject, error) {
	who, err := h.Auth.Run(ctx, h.sessionID(ctx))
	if errors.Is(err, app.ErrNoSession) {
		return gen.GetMe401Response{}, nil
	}
	if err != nil {
		return nil, err
	}
	return gen.GetMe200JSONResponse{
		Id: who.User.ID, DisplayName: who.User.DisplayName, Role: gen.MeRole(who.User.Role), TwoFactorEnabled: who.User.MFARequired,
	}, nil
}

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
