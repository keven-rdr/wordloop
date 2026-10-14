package httpapi

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/google/uuid"

	"github.com/keven-rdr/wordloop/api/internal/api/gen"
	"github.com/keven-rdr/wordloop/api/internal/identity/app"
	"github.com/keven-rdr/wordloop/api/internal/identity/domain"
	"github.com/keven-rdr/wordloop/api/internal/identity/passkeys"
	"github.com/keven-rdr/wordloop/api/internal/platform/httpx"
)

type Passkeys struct {
	BeginRegistration  passkeys.BeginRegistration
	FinishRegistration passkeys.FinishRegistration
	BeginLogin         passkeys.BeginLogin
	FinishLogin        passkeys.FinishLogin
	List               passkeys.List
	Delete             passkeys.Delete
}

func (h Handler) BeginPasskeyRegistration(ctx context.Context, req gen.BeginPasskeyRegistrationRequestObject) (gen.BeginPasskeyRegistrationResponseObject, error) {
	if h.Passkeys == nil {
		return gen.BeginPasskeyRegistration503Response{}, nil
	}
	subject, err := h.registrationSubject(ctx, req)
	if err != nil {
		return nil, err
	}
	started, err := h.Passkeys.BeginRegistration.Run(ctx, subject)
	switch {
	case errors.Is(err, passkeys.ErrUnauthorized):
		return gen.BeginPasskeyRegistration401Response{}, nil
	case errors.Is(err, passkeys.ErrTooMany):
		return gen.BeginPasskeyRegistration409Response{}, nil
	case err != nil:
		return nil, err
	}
	return h.optionsReply(started), nil
}

func (h Handler) registrationSubject(ctx context.Context, req gen.BeginPasskeyRegistrationRequestObject) (passkeys.Subject, error) {
	who, err := h.Auth.Run(ctx, h.sessionID(ctx))
	switch {
	case err == nil:
		return passkeys.Subject{SessionUserID: who.User.ID}, nil
	case !errors.Is(err, app.ErrNoSession):
		return passkeys.Subject{}, err
	case req.Body != nil && req.Body.Invite != nil:
		return passkeys.Subject{InviteToken: *req.Body.Invite}, nil
	}
	return passkeys.Subject{}, nil
}

func (h Handler) optionsReply(s passkeys.Started) jsonReply {
	return jsonReply{status: http.StatusOK, body: gen.PasskeyOptions{Options: s.Options}, cookies: []*http.Cookie{h.cookies().challenge(s.ChallengeID)}}
}

func (h Handler) FinishPasskeyRegistration(ctx context.Context, req gen.FinishPasskeyRegistrationRequestObject) (gen.FinishPasskeyRegistrationResponseObject, error) {
	if h.Passkeys == nil {
		return gen.FinishPasskeyRegistration503Response{}, nil
	}
	c := h.cookies()
	cleared := []*http.Cookie{c.clearChallenge()}
	if req.Body == nil || len(req.Body.Credential) == 0 {
		return statusReply{status: http.StatusBadRequest, cookies: cleared}, nil
	}
	name := ""
	if req.Body.Name != nil {
		name = *req.Body.Name
	}
	got, err := h.Passkeys.FinishRegistration.Run(ctx, passkeys.FinishRegistrationInput{
		ChallengeID: httpx.Cookie(ctx, c.challengeName()), Credential: req.Body.Credential, Name: name, UserAgent: httpx.UserAgent(ctx),
	})
	switch {
	case errors.Is(err, passkeys.ErrInvalid):
		return statusReply{status: http.StatusBadRequest, cookies: cleared}, nil
	case errors.Is(err, passkeys.ErrUnauthorized):
		return statusReply{status: http.StatusUnauthorized, cookies: cleared}, nil
	case errors.Is(err, passkeys.ErrDuplicate), errors.Is(err, passkeys.ErrTooMany):
		return statusReply{status: http.StatusConflict, cookies: cleared}, nil
	case err != nil:
		return nil, err
	}
	if got.Session != nil {
		cleared = append(cleared, c.session(got.Session.SessionID, got.Session.ExpiresAt))
	}
	return jsonReply{status: http.StatusCreated, body: toPasskey(got.Credential), cookies: cleared}, nil
}

func (h Handler) BeginPasskeyLogin(ctx context.Context, _ gen.BeginPasskeyLoginRequestObject) (gen.BeginPasskeyLoginResponseObject, error) {
	if h.Passkeys == nil {
		return gen.BeginPasskeyLogin503Response{}, nil
	}
	started, err := h.Passkeys.BeginLogin.Run(ctx)
	if err != nil {
		return nil, err
	}
	return h.optionsReply(started), nil
}

func (h Handler) FinishPasskeyLogin(ctx context.Context, req gen.FinishPasskeyLoginRequestObject) (gen.FinishPasskeyLoginResponseObject, error) {
	if h.Passkeys == nil {
		return gen.FinishPasskeyLogin503Response{}, nil
	}
	c := h.cookies()
	cleared := []*http.Cookie{c.clearChallenge()}
	if req.Body == nil || len(req.Body.Credential) == 0 {
		return statusReply{status: http.StatusBadRequest, cookies: cleared}, nil
	}
	opened, err := h.Passkeys.FinishLogin.Run(ctx, passkeys.FinishLoginInput{
		ChallengeID: httpx.Cookie(ctx, c.challengeName()), Credential: req.Body.Credential, UserAgent: httpx.UserAgent(ctx),
	})
	if errors.Is(err, passkeys.ErrUnauthorized) || errors.Is(err, app.ErrMFARequired) {
		slog.Info("login por passkey recusado", "err", err)
		return statusReply{status: http.StatusUnauthorized, cookies: cleared}, nil
	}
	if err != nil {
		return nil, err
	}
	return statusReply{status: http.StatusNoContent, cookies: append(cleared, c.session(opened.SessionID, opened.ExpiresAt))}, nil
}

func (h Handler) ListPasskeys(ctx context.Context, _ gen.ListPasskeysRequestObject) (gen.ListPasskeysResponseObject, error) {
	who, err := h.Auth.Run(ctx, h.sessionID(ctx))
	if errors.Is(err, app.ErrNoSession) {
		return gen.ListPasskeys401Response{}, nil
	}
	if err != nil {
		return nil, err
	}
	if h.Passkeys == nil {
		return gen.ListPasskeys200JSONResponse{}, nil
	}
	views, err := h.Passkeys.List.Run(ctx, who.User.ID)
	if err != nil {
		return nil, err
	}
	out := make(gen.ListPasskeys200JSONResponse, 0, len(views))
	for _, v := range views {
		out = append(out, toPasskeyView(v))
	}
	return out, nil
}

func (h Handler) DeletePasskey(ctx context.Context, req gen.DeletePasskeyRequestObject) (gen.DeletePasskeyResponseObject, error) {
	who, err := h.Auth.Run(ctx, h.sessionID(ctx))
	if errors.Is(err, app.ErrNoSession) {
		return gen.DeletePasskey401Response{}, nil
	}
	if err != nil {
		return nil, err
	}
	if h.Passkeys == nil {
		return gen.DeletePasskey404Response{}, nil
	}
	return deletePasskey(ctx, h.Passkeys.Delete, who, req.Id)
}

func deletePasskey(ctx context.Context, uc passkeys.Delete, who domain.Principal, id uuid.UUID) (gen.DeletePasskeyResponseObject, error) {
	err := uc.Run(ctx, who, id)
	switch {
	case errors.Is(err, passkeys.ErrNotFound):
		return gen.DeletePasskey404Response{}, nil
	case errors.Is(err, passkeys.ErrLastMethod):
		return gen.DeletePasskey409Response{}, nil
	case err != nil:
		return nil, err
	}
	return gen.DeletePasskey204Response{}, nil
}

func toPasskey(c passkeys.Credential) gen.Passkey {
	return toPasskeyView(passkeys.View{ID: c.ID, Name: c.Name, CreatedAt: c.CreatedAt, LastUsedAt: c.LastUsedAt, BackupEligible: c.BackupEligible})
}

func toPasskeyView(v passkeys.View) gen.Passkey {
	p := gen.Passkey{Id: v.ID, Name: v.Name, CreatedAt: v.CreatedAt, BackupEligible: v.BackupEligible}
	if !v.LastUsedAt.IsZero() {
		t := v.LastUsedAt
		p.LastUsedAt = &t
	}
	return p
}
