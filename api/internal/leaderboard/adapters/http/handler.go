// Package httpapi expoe o ranking pelo contrato gerado (strict server).
package httpapi

import (
	"context"
	"errors"

	openapi_types "github.com/oapi-codegen/runtime/types"

	"github.com/keven-rdr/wordloop/api/internal/api/gen"
	"github.com/keven-rdr/wordloop/api/internal/leaderboard/app"
	"github.com/keven-rdr/wordloop/api/internal/leaderboard/domain"
)

// defaultLimit e quantas posicoes vem sem limit no pedido.
const defaultLimit = 50

// Handler expoe o servico do ranking.
type Handler struct{ Service app.Service }

func (h Handler) GetLeaderboard(ctx context.Context, req gen.GetLeaderboardRequestObject) (gen.GetLeaderboardResponseObject, error) {
	period, limit := domain.Week, defaultLimit
	if req.Params.Period != nil {
		period = domain.Period(*req.Params.Period)
	}
	if req.Params.Limit != nil {
		limit = *req.Params.Limit
	}
	b, err := h.Service.Board(ctx, period, limit)
	switch {
	case errors.Is(err, app.ErrNoSession):
		return gen.GetLeaderboard401Response{}, nil
	case errors.Is(err, domain.ErrInvalidPeriod):
		return gen.GetLeaderboard400Response{}, nil
	case err != nil:
		return nil, err
	}
	return gen.GetLeaderboard200JSONResponse(board(b)), nil
}

func (h Handler) GetMyScore(ctx context.Context, _ gen.GetMyScoreRequestObject) (gen.GetMyScoreResponseObject, error) {
	s, err := h.Service.Score(ctx)
	switch {
	case errors.Is(err, app.ErrNoSession):
		return gen.GetMyScore401Response{}, nil
	case err != nil:
		return nil, err
	}
	return gen.GetMyScore200JSONResponse{
		WeekPoints: s.WeekPoints, TotalPoints: s.TotalPoints, WeekDaysMet: s.WeekDaysMet,
		Level: s.Level.Number, LevelFloor: s.Level.Floor, NextLevel: s.Level.Next,
	}, nil
}

func (h Handler) GetLeaderboardProfile(ctx context.Context, _ gen.GetLeaderboardProfileRequestObject) (gen.GetLeaderboardProfileResponseObject, error) {
	p, err := h.Service.Profile(ctx)
	switch {
	case errors.Is(err, app.ErrNoSession):
		return gen.GetLeaderboardProfile401Response{}, nil
	case err != nil:
		return nil, err
	}
	return gen.GetLeaderboardProfile200JSONResponse{Nickname: p.Nickname, Visible: p.Visible}, nil
}

func (h Handler) SaveLeaderboardProfile(ctx context.Context, req gen.SaveLeaderboardProfileRequestObject) (gen.SaveLeaderboardProfileResponseObject, error) {
	if req.Body == nil {
		return gen.SaveLeaderboardProfile400Response{}, nil
	}
	p, err := h.Service.SaveProfile(ctx, domain.Profile{Nickname: req.Body.Nickname, Visible: req.Body.Visible})
	switch {
	case errors.Is(err, app.ErrNoSession):
		return gen.SaveLeaderboardProfile401Response{}, nil
	case errors.Is(err, domain.ErrInvalidNickname):
		return gen.SaveLeaderboardProfile400Response{}, nil
	case errors.Is(err, domain.ErrNicknameTaken):
		return gen.SaveLeaderboardProfile409Response{}, nil
	case err != nil:
		return nil, err
	}
	return gen.SaveLeaderboardProfile200JSONResponse{Nickname: p.Nickname, Visible: p.Visible}, nil
}

func board(b domain.Board) gen.Leaderboard {
	out := gen.Leaderboard{
		Period: gen.LeaderboardPeriod(b.Period), From: openapi_types.Date{Time: b.From}, To: openapi_types.Date{Time: b.To},
		Entries: make([]gen.LeaderboardEntry, 0, len(b.Entries)),
	}
	for _, e := range b.Entries {
		out.Entries = append(out.Entries, entry(e, b.Me != nil && b.Me.UserID == e.UserID))
	}
	if b.Me != nil {
		me := entry(*b.Me, true)
		out.Me = &me
	}
	return out
}

func entry(e domain.Entry, me bool) gen.LeaderboardEntry {
	out := gen.LeaderboardEntry{Position: e.Position, Points: e.Points, Me: me}
	if e.Nickname != "" {
		out.Nickname = &e.Nickname
	}
	return out
}
