//go:build integration

package postgres_test

import (
	"context"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/keven-rdr/wordloop/api/internal/learning/app"
	"github.com/keven-rdr/wordloop/api/internal/learning/domain"
	"github.com/keven-rdr/wordloop/api/internal/srs"
)

// Uma frase da carga real vira "montar frase" no cartao produtivo, e a resposta fica no log com o tipo novo
// (o CHECK de review_log.exercise_type da migracao 00019 precisa aceitar `assemble`).
func TestAnAssembledSentenceIsPlannedGradedAndLoggedWithTheNewExerciseType(t *testing.T) {
	f := setup(t, false)
	ctx := context.Background()
	mustExec(t, f, `UPDATE item SET status = 'published'`)
	var item uuid.UUID
	var term string
	err := f.pool.QueryRow(ctx, `
		SELECT i.id, t.body FROM item i JOIN item_text t ON t.item_id = i.id AND t.role = 'term' AND t.locale = 'en-US'
		WHERE i.kind = 'sentence' ORDER BY i.id LIMIT 1`).Scan(&item, &term)
	if err != nil {
		t.Fatal(err)
	}
	var cardID uuid.UUID
	err = f.pool.QueryRow(ctx, `
		INSERT INTO card (user_id, item_id, direction, state, stability, due_at, last_review_at, reps)
		VALUES ($1, $2, 'productive', 'review', 3, $3, $3::timestamptz - interval '4 days', 3) RETURNING id`, f.user, item, now).Scan(&cardID)
	if err != nil {
		t.Fatal(err)
	}

	pack, err := f.plan.Run(ctx, app.PlanRequest{Minutes: 15, Mode: domain.ModeReview})
	if err != nil {
		t.Fatal(err)
	}
	var tiles []string
	for _, c := range pack.Cards {
		if c.Card.ID == cardID {
			if c.Exercise != domain.Assemble {
				t.Fatalf("frase no cartao produtivo: %v", c.Exercise)
			}
			tiles = c.Tiles
		}
	}
	words := strings.Fields(strings.Trim(term, ".?!"))
	if len(tiles) != len(words) || slices.Equal(tiles, words) {
		t.Fatalf("as pecas sao as palavras de %q, fora da ordem: %v", term, tiles)
	}

	a := domain.Answer{
		ID: uuid.New(), CardID: cardID, SessionID: pack.SessionID, Exercise: domain.Assemble, Text: term,
		ShownAt: now.Add(-time.Minute), AnsweredAt: now.Add(-50 * time.Second), ResponseMS: 8000, Difficulty: srs.DifficultyMedium, ClientSeq: 1,
	}
	res, err := f.submit.Run(ctx, []domain.Answer{a})
	if err != nil || res[0].Status != domain.Applied || !res[0].Correct {
		t.Fatalf("a frase certa e aceita: %+v %v", res, err)
	}
	var exercise string
	var correct bool
	if err := f.pool.QueryRow(ctx, `SELECT exercise_type, is_correct FROM review_log WHERE id = $1`, a.ID).Scan(&exercise, &correct); err != nil || exercise != "assemble" || !correct {
		t.Fatalf("o log guarda o tipo novo: %q %v %v", exercise, correct, err)
	}
}
