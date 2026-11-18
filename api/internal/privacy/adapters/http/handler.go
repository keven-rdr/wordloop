// Package httpapi expoe a exportacao e a exclusao de dados pelo contrato gerado (strict server).
package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/keven-rdr/wordloop/api/internal/api/gen"
	"github.com/keven-rdr/wordloop/api/internal/privacy/app"
	"github.com/keven-rdr/wordloop/api/internal/privacy/domain"
)

// SchemaVersion e a versao do formato do arquivo exportado; sobe quando uma secao muda de forma.
const SchemaVersion = 1

// Handler junta os casos de uso de privacidade.
type Handler struct{ Service app.Service }

func (h Handler) ExportMyData(ctx context.Context, _ gen.ExportMyDataRequestObject) (gen.ExportMyDataResponseObject, error) {
	out, err := h.Service.Export(ctx)
	if errors.Is(err, app.ErrNoSession) {
		return gen.ExportMyData401Response{}, nil
	}
	if err != nil {
		return nil, err
	}
	body := gen.ExportData{}
	if err := json.Unmarshal(out.Data, &body); err != nil {
		return nil, fmt.Errorf("montar a exportacao: %w", err)
	}
	body["schema_version"] = SchemaVersion
	body["generated_at"] = out.GeneratedAt
	name := fmt.Sprintf(`attachment; filename="wordloop-dados-%s.json"`, out.GeneratedAt.Format("2006-01-02"))
	return gen.ExportMyData200JSONResponse{Body: body, Headers: gen.ExportMyData200ResponseHeaders{ContentDisposition: &name}}, nil
}

func (h Handler) EraseMyAccount(ctx context.Context, req gen.EraseMyAccountRequestObject) (gen.EraseMyAccountResponseObject, error) {
	if req.Body == nil {
		return gen.EraseMyAccount400Response{}, nil
	}
	_, err := h.Service.Erase(ctx, req.Body.Confirmation)
	switch {
	case errors.Is(err, app.ErrNoSession), errors.Is(err, domain.ErrNotFound):
		return gen.EraseMyAccount401Response{}, nil
	case errors.Is(err, domain.ErrAdminAccount):
		return gen.EraseMyAccount403Response{}, nil
	case errors.Is(err, domain.ErrRecentLoginRequired):
		return gen.EraseMyAccount428Response{}, nil
	case errors.Is(err, domain.ErrWrongConfirmation):
		return gen.EraseMyAccount400Response{}, nil
	case err != nil:
		return nil, err
	}
	return gen.EraseMyAccount204Response{}, nil
}
