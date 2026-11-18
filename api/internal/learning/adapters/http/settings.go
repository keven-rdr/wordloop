package httpapi

import (
	"context"
	"errors"

	"github.com/keven-rdr/wordloop/api/internal/api/gen"
	"github.com/keven-rdr/wordloop/api/internal/learning/app"
	"github.com/keven-rdr/wordloop/api/internal/learning/domain"
)

func (h Handler) GetSettings(ctx context.Context, _ gen.GetSettingsRequestObject) (gen.GetSettingsResponseObject, error) {
	s, err := h.Settings.Get(ctx)
	switch {
	case errors.Is(err, app.ErrNoSession):
		return gen.GetSettings401Response{}, nil
	case err != nil:
		return nil, err
	}
	return gen.GetSettings200JSONResponse(settings(s)), nil
}

func (h Handler) UpdateSettings(ctx context.Context, req gen.UpdateSettingsRequestObject) (gen.UpdateSettingsResponseObject, error) {
	if req.Body == nil {
		return gen.UpdateSettings400Response{}, nil
	}
	s, err := h.Settings.Update(ctx, domain.Settings{
		TZ: req.Body.Tz, UILocale: string(req.Body.UiLocale), Theme: string(req.Body.Theme), TZAuto: req.Body.TzAuto,
	})
	switch {
	case errors.Is(err, app.ErrNoSession):
		return gen.UpdateSettings401Response{}, nil
	case errors.Is(err, domain.ErrInvalidSettings):
		return gen.UpdateSettings400Response{}, nil
	case err != nil:
		return nil, err
	}
	return gen.UpdateSettings200JSONResponse(settings(s)), nil
}

func settings(s domain.Settings) gen.Settings {
	return gen.Settings{Tz: s.TZ, TzAuto: s.TZAuto, UiLocale: gen.SettingsUiLocale(s.UILocale), Theme: gen.SettingsTheme(s.Theme)}
}
