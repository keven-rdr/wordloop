package api

import (
	"context"

	"github.com/keven-rdr/wordloop/api/internal/api/gen"
)

// Leaderboard e o modulo do ranking (pontos, nivel e podio); nulo = ranking indisponivel.
type Leaderboard interface {
	GetLeaderboard(ctx context.Context, req gen.GetLeaderboardRequestObject) (gen.GetLeaderboardResponseObject, error)
	GetMyScore(ctx context.Context, req gen.GetMyScoreRequestObject) (gen.GetMyScoreResponseObject, error)
	GetLeaderboardProfile(ctx context.Context, req gen.GetLeaderboardProfileRequestObject) (gen.GetLeaderboardProfileResponseObject, error)
	SaveLeaderboardProfile(ctx context.Context, req gen.SaveLeaderboardProfileRequestObject) (gen.SaveLeaderboardProfileResponseObject, error)
}

func (s Server) GetLeaderboard(ctx context.Context, req gen.GetLeaderboardRequestObject) (gen.GetLeaderboardResponseObject, error) {
	if s.Leaderboard == nil {
		return gen.GetLeaderboard401Response{}, nil
	}
	return s.Leaderboard.GetLeaderboard(ctx, req)
}

func (s Server) GetMyScore(ctx context.Context, req gen.GetMyScoreRequestObject) (gen.GetMyScoreResponseObject, error) {
	if s.Leaderboard == nil {
		return gen.GetMyScore401Response{}, nil
	}
	return s.Leaderboard.GetMyScore(ctx, req)
}

func (s Server) GetLeaderboardProfile(ctx context.Context, req gen.GetLeaderboardProfileRequestObject) (gen.GetLeaderboardProfileResponseObject, error) {
	if s.Leaderboard == nil {
		return gen.GetLeaderboardProfile401Response{}, nil
	}
	return s.Leaderboard.GetLeaderboardProfile(ctx, req)
}

func (s Server) SaveLeaderboardProfile(ctx context.Context, req gen.SaveLeaderboardProfileRequestObject) (gen.SaveLeaderboardProfileResponseObject, error) {
	if s.Leaderboard == nil {
		return gen.SaveLeaderboardProfile401Response{}, nil
	}
	return s.Leaderboard.SaveLeaderboardProfile(ctx, req)
}
