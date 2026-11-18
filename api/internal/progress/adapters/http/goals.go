package httpapi

import (
	"context"
	"errors"

	"github.com/keven-rdr/wordloop/api/internal/api/gen"
	"github.com/keven-rdr/wordloop/api/internal/progress/app"
	"github.com/keven-rdr/wordloop/api/internal/progress/domain"
)

func (h Handler) GetCurrentGoal(ctx context.Context, _ gen.GetCurrentGoalRequestObject) (gen.GetCurrentGoalResponseObject, error) {
	cur, err := h.Goals.Current(ctx)
	switch {
	case errors.Is(err, app.ErrNoSession):
		return gen.GetCurrentGoal401Response{}, nil
	case err != nil:
		return nil, err
	}
	return gen.GetCurrentGoal200JSONResponse(currentGoal(cur)), nil
}

func (h Handler) GetGoalSuggestions(ctx context.Context, _ gen.GetGoalSuggestionsRequestObject) (gen.GetGoalSuggestionsResponseObject, error) {
	suggestions, fromHistory, err := h.Goals.Suggest(ctx)
	switch {
	case errors.Is(err, app.ErrNoSession):
		return gen.GetGoalSuggestions401Response{}, nil
	case err != nil:
		return nil, err
	}
	return gen.GetGoalSuggestions200JSONResponse(goalSuggestions(suggestions, fromHistory)), nil
}

func (h Handler) StartGoal(ctx context.Context, req gen.StartGoalRequestObject) (gen.StartGoalResponseObject, error) {
	if req.Body == nil {
		return gen.StartGoal400Response{}, nil
	}
	cur, err := h.Goals.Start(ctx, goalRequest(*req.Body))
	switch {
	case errors.Is(err, app.ErrNoSession):
		return gen.StartGoal401Response{}, nil
	case errors.Is(err, domain.ErrInvalidGoal):
		return gen.StartGoal400Response{}, nil
	case err != nil:
		return nil, err
	}
	return gen.StartGoal201JSONResponse(currentGoal(cur)), nil
}
