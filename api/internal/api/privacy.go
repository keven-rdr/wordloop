package api

import (
	"context"

	"github.com/keven-rdr/wordloop/api/internal/api/gen"
)

// Privacy e o modulo dos direitos da pessoa sobre os dados (exportar e apagar); nulo = indisponivel.
type Privacy interface {
	ExportMyData(ctx context.Context, req gen.ExportMyDataRequestObject) (gen.ExportMyDataResponseObject, error)
	EraseMyAccount(ctx context.Context, req gen.EraseMyAccountRequestObject) (gen.EraseMyAccountResponseObject, error)
}

func (s Server) ExportMyData(ctx context.Context, req gen.ExportMyDataRequestObject) (gen.ExportMyDataResponseObject, error) {
	if s.Privacy == nil {
		return gen.ExportMyData401Response{}, nil
	}
	return s.Privacy.ExportMyData(ctx, req)
}

func (s Server) EraseMyAccount(ctx context.Context, req gen.EraseMyAccountRequestObject) (gen.EraseMyAccountResponseObject, error) {
	if s.Privacy == nil {
		return gen.EraseMyAccount401Response{}, nil
	}
	return s.Privacy.EraseMyAccount(ctx, req)
}
