package httpapi

import (
	"context"
	"errors"

	"github.com/keven-rdr/wordloop/api/internal/api/gen"
	"github.com/keven-rdr/wordloop/api/internal/identity/accounts"
	"github.com/keven-rdr/wordloop/api/internal/identity/app"
)

func (h Handler) SetTwoFactor(ctx context.Context, req gen.SetTwoFactorRequestObject) (gen.SetTwoFactorResponseObject, error) {
	who, err := h.Auth.Run(ctx, h.sessionID(ctx))
	if errors.Is(err, app.ErrNoSession) {
		return gen.SetTwoFactor401Response{}, nil
	}
	if err != nil {
		return nil, err
	}
	err = h.TwoFactor.Set(ctx, who, req.Body.Enabled)
	switch {
	case errors.Is(err, accounts.ErrForbidden):
		return gen.SetTwoFactor403Response{}, nil
	case errors.Is(err, accounts.ErrWeakSession):
		return gen.SetTwoFactor409Response{}, nil
	case err != nil:
		return nil, err
	}
	return gen.SetTwoFactor204Response{}, nil
}
