package app

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type SessionView struct {
	ID         uuid.UUID
	CreatedAt  time.Time
	LastUsedAt time.Time
	UserAgent  string
	Current    bool
}

type ListSessions struct {
	Auth     Authenticate
	Sessions Sessions
}

func (u ListSessions) Run(ctx context.Context, sessionID string) ([]SessionView, error) {
	who, err := u.Auth.Run(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	recs, err := u.Sessions.ListByUser(ctx, who.User.ID)
	if err != nil {
		return nil, fmt.Errorf("listar as sessoes: %w", err)
	}
	now := u.Auth.now()
	views := make([]SessionView, 0, len(recs))
	for _, r := range recs {
		if !u.Auth.alive(r, who.User.Role, now) {
			continue
		}
		views = append(views, SessionView{
			ID: r.PublicID, CreatedAt: r.CreatedAt, LastUsedAt: r.LastUsedAt, UserAgent: r.UserAgent, Current: r.PublicID == who.SessionID,
		})
	}
	return views, nil
}
