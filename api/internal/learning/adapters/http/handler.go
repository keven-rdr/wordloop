// Package httpapi expoe o estudo pelo contrato gerado (strict server).
package httpapi

import (
	"context"
	"errors"

	"github.com/keven-rdr/wordloop/api/internal/api/gen"
	"github.com/keven-rdr/wordloop/api/internal/learning/app"
	"github.com/keven-rdr/wordloop/api/internal/learning/domain"
)

// Handler junta os casos de uso de estudo.
type Handler struct {
	Plan   app.PlanSession
	Submit app.SubmitReviews
}

func (h Handler) PlanSession(ctx context.Context, req gen.PlanSessionRequestObject) (gen.PlanSessionResponseObject, error) {
	if req.Body == nil {
		return gen.PlanSession400Response{}, nil
	}
	pack, err := h.Plan.Run(ctx, planRequest(*req.Body))
	switch {
	case errors.Is(err, app.ErrNoSession):
		return gen.PlanSession401Response{}, nil
	case errors.Is(err, domain.ErrInvalidMinutes), errors.Is(err, domain.ErrInvalidMode):
		return gen.PlanSession400Response{}, nil
	case err != nil:
		return nil, err
	}
	return gen.PlanSession200JSONResponse(studyPack(pack)), nil
}

func (h Handler) SubmitReviews(ctx context.Context, req gen.SubmitReviewsRequestObject) (gen.SubmitReviewsResponseObject, error) {
	if req.Body == nil {
		return gen.SubmitReviews400Response{}, nil
	}
	results, err := h.Submit.Run(ctx, answers(req.Body.Reviews))
	switch {
	case errors.Is(err, app.ErrNoSession):
		return gen.SubmitReviews401Response{}, nil
	case errors.Is(err, app.ErrInvalidBatch):
		return gen.SubmitReviews400Response{}, nil
	case err != nil:
		return nil, err
	}
	return gen.SubmitReviews200JSONResponse(batchResult(results)), nil
}
