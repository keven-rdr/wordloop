package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/keven-rdr/wordloop/api/internal/learning/adapters/postgres/db"
	"github.com/keven-rdr/wordloop/api/internal/learning/domain"
)

func optUUID(id uuid.UUID) *uuid.UUID {
	if id == uuid.Nil {
		return nil
	}
	return &id
}

func optText(s string) pgtype.Text { return pgtype.Text{String: s, Valid: s != ""} }

func float(v float64) pgtype.Float8 { return pgtype.Float8{Float64: v, Valid: true} }

// InsertReview grava a linha do log; false se o UUID ja existia (reenvio do mesmo lote).
func (t *txn) InsertReview(ctx context.Context, r domain.LogRow) (bool, error) {
	n, err := t.q.InsertReview(ctx, params(r))
	if err != nil {
		return false, fmt.Errorf("gravar revisao: %w", err)
	}
	return n == 1, nil
}

func params(r domain.LogRow) db.InsertReviewParams {
	p := db.InsertReviewParams{
		ID: r.ID, UserID: r.UserID, CardID: r.CardID, SessionID: optUUID(r.SessionID), DeviceID: optUUID(r.DeviceID),
		ExerciseType: string(r.Exercise), Source: r.Source, ShownAt: ts(r.ShownAt), AnsweredAt: ts(r.AnsweredAt),
		Tz: r.TZ, DayCutoffHour: int16(r.CutoffHour), //nolint:gosec // 0 a 8
		LocalDate: pgtype.Date{Time: r.LocalDate, Valid: true},
		AnswerRaw: optText(r.Text), ChosenItemID: optUUID(r.ChosenItemID),
		IsCorrect: r.Correct, TypoFlag: r.Typo, HintUsed: r.HintUsed, ExplanationViewed: r.Explained, ImageBeforeAnswer: r.ImageBefore,
		DifficultySuggested: optText(r.DifficultyHint), DifficultyChosen: optText(string(r.Difficulty)), ErrorKind: optText(string(r.ErrorKind)),
		Grade:       int16(r.Grade), //nolint:gosec // nota de 1 a 4
		StateBefore: string(r.Before.State), SBefore: float(r.Before.Stability),
		RAtReview: float(r.RAtReview), ElapsedDays: float(r.ElapsedDays), StateAfter: string(r.After.State),
		SAfter: float(r.After.Stability), ScheduledDays: float(r.ScheduledDays), DueAfter: ts(r.After.DueAt),
		Scheduler: r.Scheduler, SchedulerVersion: r.SchedulerVersion,
	}
	if r.Source == "" {
		p.Source = "session"
	}
	p.ClientSeq = pgtype.Int8{Int64: r.ClientSeq, Valid: r.ClientSeq != 0}
	p.ResponseMs = pgtype.Int4{Int32: int32(r.ResponseMS), Valid: r.ResponseMS > 0} //nolint:gosec // ms de uma resposta
	if r.HasClientCorrect && r.ClientCorrect != nil {
		p.ClientCorrect = pgtype.Bool{Bool: *r.ClientCorrect, Valid: true}
	}
	return p
}
