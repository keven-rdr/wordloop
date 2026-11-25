//go:build integration

package postgres_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/keven-rdr/wordloop/api/internal/learning/app"
	"github.com/keven-rdr/wordloop/api/internal/learning/domain"
	"github.com/keven-rdr/wordloop/api/internal/srs"
)

// A resposta soma no study_day na mesma transacao do log; reenviar o mesmo UUID nao soma de novo.
func TestSubmitSumsTheAnswerIntoTheStudyDayOnceAndInTheSameTransaction(t *testing.T) {
	f := setup(t, false)
	ctx := context.Background()
	pack, _ := f.plan.Run(ctx, app.PlanRequest{Minutes: 15})
	card := pack.Cards[0].Card
	right := domain.Answer{
		ID: uuid.New(), CardID: card.ID, Exercise: domain.Choose, ChosenItemID: card.ItemID,
		ShownAt: now.Add(-time.Minute), AnsweredAt: now.Add(-50 * time.Second), ResponseMS: 3200, Difficulty: srs.DifficultyMedium,
	}
	wrong := domain.Answer{
		ID: uuid.New(), CardID: pack.Cards[1].Card.ID, Exercise: domain.Choose, ChosenItemID: card.ItemID,
		ShownAt: now.Add(-40 * time.Second), AnsweredAt: now.Add(-30 * time.Second), ResponseMS: 5000,
	}
	for _, batch := range [][]domain.Answer{{right, wrong}, {right, wrong}} { // o 2o envio e so reenvio
		if _, err := f.submit.Run(ctx, batch); err != nil {
			t.Fatal(err)
		}
	}
	var reviews, correct, lapses, fresh, known int
	var ms int64
	err := f.pool.QueryRow(ctx, `SELECT reviews, correct, lapses, new_items, active_ms, COALESCE(known_lexemes, -1)
		FROM study_day WHERE user_id = $1 AND local_date = '2026-10-19'`, f.user).Scan(&reviews, &correct, &lapses, &fresh, &ms, &known)
	if err != nil || reviews != 2 || correct != 1 || lapses != 0 || fresh != 2 || ms != 8200 || known != 0 {
		t.Fatalf("dia: reviews=%d correct=%d lapses=%d new=%d ms=%d known=%d (%v)", reviews, correct, lapses, fresh, ms, known, err)
	}
	if count(t, f.pool, `SELECT count(*) FROM study_day`) != 1 {
		t.Error("as duas respostas do mesmo dia ficam numa linha so")
	}
	// Pontos: o acerto pelo nivel do item mais 2 pelo erro; o reenvio nao soma de novo.
	var cefr string
	if err := f.pool.QueryRow(ctx, `SELECT COALESCE(cefr, '') FROM item WHERE id = $1`, card.ItemID).Scan(&cefr); err != nil {
		t.Fatal(err)
	}
	want := domain.ReviewPoints(true, false, cefr) + domain.ReviewPoints(false, false, "")
	if got := count(t, f.pool, `SELECT points FROM study_day WHERE user_id = $1`, f.user); got != want {
		t.Errorf("pontos do dia: %d, esperava %d", got, want)
	}
}
