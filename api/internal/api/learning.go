package api

import (
	"context"

	"github.com/keven-rdr/wordloop/api/internal/api/gen"
)

// Learning e o modulo de estudo (plano da sessao e envio de revisoes); nulo = estudo indisponivel.
type Learning interface {
	PlanSession(ctx context.Context, req gen.PlanSessionRequestObject) (gen.PlanSessionResponseObject, error)
	SubmitReviews(ctx context.Context, req gen.SubmitReviewsRequestObject) (gen.SubmitReviewsResponseObject, error)
	GetSettings(ctx context.Context, req gen.GetSettingsRequestObject) (gen.GetSettingsResponseObject, error)
	UpdateSettings(ctx context.Context, req gen.UpdateSettingsRequestObject) (gen.UpdateSettingsResponseObject, error)
}

func (s Server) PlanSession(ctx context.Context, req gen.PlanSessionRequestObject) (gen.PlanSessionResponseObject, error) {
	if s.Learning == nil {
		return gen.PlanSession401Response{}, nil
	}
	return s.Learning.PlanSession(ctx, req)
}

func (s Server) SubmitReviews(ctx context.Context, req gen.SubmitReviewsRequestObject) (gen.SubmitReviewsResponseObject, error) {
	if s.Learning == nil {
		return gen.SubmitReviews401Response{}, nil
	}
	return s.Learning.SubmitReviews(ctx, req)
}

func (s Server) GetSettings(ctx context.Context, req gen.GetSettingsRequestObject) (gen.GetSettingsResponseObject, error) {
	if s.Learning == nil {
		return gen.GetSettings401Response{}, nil
	}
	return s.Learning.GetSettings(ctx, req)
}

func (s Server) UpdateSettings(ctx context.Context, req gen.UpdateSettingsRequestObject) (gen.UpdateSettingsResponseObject, error) {
	if s.Learning == nil {
		return gen.UpdateSettings401Response{}, nil
	}
	return s.Learning.UpdateSettings(ctx, req)
}
