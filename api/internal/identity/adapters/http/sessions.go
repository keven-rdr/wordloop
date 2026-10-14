package httpapi

import (
	"context"
	"errors"
	"net/http"

	"github.com/keven-rdr/wordloop/api/internal/api/gen"
	"github.com/keven-rdr/wordloop/api/internal/identity/app"
)

// ListSessions lista as sessoes de quem esta logado.
func (h Handler) ListSessions(ctx context.Context, _ gen.ListSessionsRequestObject) (gen.ListSessionsResponseObject, error) {
	views, err := h.List.Run(ctx, h.sessionID(ctx))
	if errors.Is(err, app.ErrNoSession) {
		return gen.ListSessions401Response{}, nil
	}
	if err != nil {
		return nil, err
	}
	out := make(gen.ListSessions200JSONResponse, 0, len(views))
	for _, v := range views {
		s := gen.Session{Id: v.ID, CreatedAt: v.CreatedAt, LastUsedAt: v.LastUsedAt, Current: v.Current}
		if v.UserAgent != "" {
			ua := v.UserAgent
			s.UserAgent = &ua
		}
		out = append(out, s)
	}
	return out, nil
}

// RevokeSession encerra uma sessao de quem esta logado; sessao de outra pessoa responde 404, como se nao existisse.
func (h Handler) RevokeSession(ctx context.Context, req gen.RevokeSessionRequestObject) (gen.RevokeSessionResponseObject, error) {
	current, err := h.Revoke.Run(ctx, h.sessionID(ctx), req.Id)
	if errors.Is(err, app.ErrNoSession) {
		return gen.RevokeSession401Response{}, nil
	}
	if errors.Is(err, app.ErrNotFound) {
		return gen.RevokeSession404Response{}, nil
	}
	if err != nil {
		return nil, err
	}
	if current {
		return sessionEnded{cookies: []*http.Cookie{h.cookies().clearSession()}}, nil
	}
	return gen.RevokeSession204Response{}, nil
}

// RevokeAllSessions encerra todas as sessoes de quem esta logado e limpa o cookie.
func (h Handler) RevokeAllSessions(ctx context.Context, _ gen.RevokeAllSessionsRequestObject) (gen.RevokeAllSessionsResponseObject, error) {
	err := h.RevokeAll.Run(ctx, h.sessionID(ctx))
	if errors.Is(err, app.ErrNoSession) {
		return gen.RevokeAllSessions401Response{}, nil
	}
	if err != nil {
		return nil, err
	}
	return sessionEnded{cookies: []*http.Cookie{h.cookies().clearSession()}}, nil
}
