package httpapi

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/keven-rdr/wordloop/api/internal/api/gen"
	"github.com/keven-rdr/wordloop/api/internal/identity/app"
	"github.com/keven-rdr/wordloop/api/internal/platform/httpx"
)

// Handler implementa as operacoes de identidade do contrato. Start e Finish nulos = login nao configurado (503).
type Handler struct {
	Start     *app.StartLogin
	Finish    *app.FinishLogin
	Auth      app.Authenticate
	Out       app.Logout
	List      app.ListSessions
	Revoke    app.RevokeSession
	RevokeAll app.RevokeAllSessions
	Secure    bool // cookies Secure e com prefixo (URL publica em https)
}

func (h Handler) cookies() cookies { return cookies{secure: h.Secure} }

// sessionID le o cookie da sessao do pedido em andamento (o nome muda com o prefixo em HTTPS).
func (h Handler) sessionID(ctx context.Context) string {
	return httpx.Cookie(ctx, h.cookies().sessionName())
}

func (h Handler) enabled() bool { return h.Start != nil && h.Finish != nil }

// StartLogin redireciona ao provedor de identidade.
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

// FinishLogin conclui o login e abre a sessao.
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
		slog.Warn("login de administrador recusado: sem 2 fatores comprovados pelo provedor")
		return mfaRequired{cookies: []*http.Cookie{c.clearFlow()}}, nil
	}
	if err != nil {
		return nil, err
	}
	return redirect{location: done.ReturnTo, cookies: []*http.Cookie{c.session(done.SessionID, done.ExpiresAt), c.clearFlow()}}, nil
}

// Logout encerra a sessao.
func (h Handler) Logout(ctx context.Context, _ gen.LogoutRequestObject) (gen.LogoutResponseObject, error) {
	if err := h.Out.Run(ctx, h.sessionID(ctx)); err != nil {
		return nil, err
	}
	return loggedOut{cookies: []*http.Cookie{h.cookies().clearSession()}}, nil
}

// GetMe devolve o perfil de quem esta logado.
func (h Handler) GetMe(ctx context.Context, _ gen.GetMeRequestObject) (gen.GetMeResponseObject, error) {
	who, err := h.Auth.Run(ctx, h.sessionID(ctx))
	if errors.Is(err, app.ErrNoSession) {
		return gen.GetMe401Response{}, nil
	}
	if err != nil {
		return nil, err
	}
	return gen.GetMe200JSONResponse{Id: who.User.ID, DisplayName: who.User.DisplayName, Role: gen.MeRole(who.User.Role)}, nil
}

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
