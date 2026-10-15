package httpapi

import (
	"context"
	"errors"
	"net/http"

	"github.com/keven-rdr/wordloop/api/internal/api/gen"
	"github.com/keven-rdr/wordloop/api/internal/identity/registration"
)

type Registration struct {
	Register registration.Register
	Forgot   registration.Forgot
}

func (r statusReply) VisitRegisterAccountResponse(w http.ResponseWriter) error { return r.write(w) }
func (r statusReply) VisitRequestPasswordResetResponse(w http.ResponseWriter) error {
	return r.write(w)
}

func locale(l *gen.EmailRequestLocale) string {
	if l == nil {
		return ""
	}
	return string(*l)
}

func (h Handler) RegisterAccount(ctx context.Context, req gen.RegisterAccountRequestObject) (gen.RegisterAccountResponseObject, error) {
	if h.Registration == nil {
		return gen.RegisterAccount503Response{}, nil
	}
	if req.Body == nil {
		return statusReply{status: http.StatusBadRequest}, nil
	}
	err := h.Registration.Register.Run(ctx, registration.Input{Email: req.Body.Email, Locale: locale(req.Body.Locale)})
	return acceptedOrBad(err)
}

func (h Handler) RequestPasswordReset(ctx context.Context, req gen.RequestPasswordResetRequestObject) (gen.RequestPasswordResetResponseObject, error) {
	if h.Registration == nil {
		return gen.RequestPasswordReset503Response{}, nil
	}
	if req.Body == nil {
		return statusReply{status: http.StatusBadRequest}, nil
	}
	err := h.Registration.Forgot.Run(ctx, registration.Input{Email: req.Body.Email, Locale: locale(req.Body.Locale)})
	return acceptedOrBad(err)
}

type accepted interface {
	gen.RegisterAccountResponseObject
	gen.RequestPasswordResetResponseObject
}

func acceptedOrBad(err error) (accepted, error) {
	if errors.Is(err, registration.ErrInvalidEmail) {
		return statusReply{status: http.StatusBadRequest}, nil
	}
	if err != nil {
		return nil, err
	}
	return statusReply{status: http.StatusAccepted}, nil
}
