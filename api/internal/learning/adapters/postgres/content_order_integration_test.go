//go:build integration

package postgres_test

import (
	"context"
	"testing"

	"github.com/keven-rdr/wordloop/api/internal/learning/app"
)

func mustExec(t *testing.T, f fixture, sql string, args ...any) {
	t.Helper()
	if _, err := f.pool.Exec(context.Background(), sql, args...); err != nil {
		t.Fatal(err)
	}
}

// Com a carga real publicada, a pessoa estuda por licao, e dentro da licao as palavras vem antes das frases.
func TestNewCardsFollowTheLessonsAndPutWordsBeforeSentences(t *testing.T) {
	f := setup(t, false)
	ctx := context.Background()
	mustExec(t, f, `UPDATE item SET status = 'published'`) // a carga real, que o fixture voltou a rascunho
	// Quem ja estudou os 6 temas do v0: cartoes em revisao, com vencimento distante (nao entram na fila).
	mustExec(t, f, `
		INSERT INTO card (user_id, item_id, direction, state, stability, due_at)
		SELECT $1, i.id, 'receptive', 'review', 30, $2::timestamptz + interval '60 days'
		FROM item i JOIN unit u ON u.id = i.unit_id WHERE u.position <= 6`, f.user, now)

	pack, err := f.plan.Run(ctx, app.PlanRequest{Minutes: 60})
	if err != nil || len(pack.Cards) != 8 {
		t.Fatalf("60 min = 8 novas: %d %v", len(pack.Cards), err)
	}
	for _, c := range pack.Cards {
		var kind string
		var position int
		err := f.pool.QueryRow(ctx, `SELECT i.kind, u.position FROM item i JOIN unit u ON u.id = i.unit_id WHERE i.id = $1`, c.Card.ItemID).Scan(&kind, &position)
		if err != nil {
			t.Fatalf("o item novo precisa ser de uma licao: %v", err)
		}
		if kind != "word" || position != 7 {
			t.Errorf("a primeira licao (posicao 7) comeca pelas palavras: %s da posicao %d (%q)", kind, position, c.Content.Term)
		}
	}
}
