//go:build integration

package postgres_test

import (
	"context"
	"testing"

	"github.com/keven-rdr/wordloop/api/internal/learning/app"
	"github.com/keven-rdr/wordloop/api/internal/learning/domain"
)

func mustExec(t *testing.T, f fixture, sql string, args ...any) {
	t.Helper()
	if _, err := f.pool.Exec(context.Background(), sql, args...); err != nil {
		t.Fatal(err)
	}
}

// Com a carga real publicada, a pessoa estuda por licao: dentro da licao as palavras vem antes das frases, em ordem
// sorteada a cada sessao; 1 em 5 palavras novas e desafio, de duas ou mais licoes adiante.
func TestNewCardsFollowTheLessonsShuffledWithStretchWordsFromLaterLessons(t *testing.T) {
	f := setup(t, false)
	ctx := context.Background()
	mustExec(t, f, `UPDATE item SET status = 'published'`) // a carga real, que o fixture voltou a rascunho
	// Quem ja estudou os 6 temas do v0: cartoes em revisao, com vencimento distante (nao entram na fila).
	mustExec(t, f, `
		INSERT INTO card (user_id, item_id, direction, state, stability, due_at)
		SELECT $1, i.id, 'receptive', 'review', 30, $2::timestamptz + interval '60 days'
		FROM item i JOIN unit u ON u.id = i.unit_id WHERE u.position <= 6`, f.user, now)

	first, err := f.plan.Run(ctx, app.PlanRequest{Minutes: 5, Mode: domain.ModeLearn})
	if err != nil || len(first.Cards) != 9 {
		t.Fatalf("5 min = 9 novas: %d %v", len(first.Cards), err)
	}
	lesson, stretch := 0, 0
	for _, c := range first.Cards {
		var kind string
		var position int
		err := f.pool.QueryRow(ctx, `SELECT i.kind, u.position FROM item i JOIN unit u ON u.id = i.unit_id WHERE i.id = $1`, c.Card.ItemID).Scan(&kind, &position)
		if err != nil {
			t.Fatalf("o item novo precisa ser de uma licao: %v", err)
		}
		switch {
		case kind != "sentence" && position == 7:
			lesson++
		case kind != "sentence" && position >= 9:
			stretch++
		default:
			t.Errorf("fora da licao atual e do desafio: %s da posicao %d (%q)", kind, position, c.Content.Term)
		}
	}
	if lesson != 8 || stretch != 1 {
		t.Errorf("8 itens (nao frases) da licao 7 e 1 de desafio: %d e %d", lesson, stretch)
	}
	// Outra sessao (os cartoes ainda nao respondidos sao apagados para comparar a escolha): outra ordem ou outras palavras.
	mustExec(t, f, `DELETE FROM card WHERE user_id = $1 AND state = 'new'`, f.user)
	second, err := f.plan.Run(ctx, app.PlanRequest{Minutes: 5, Mode: domain.ModeLearn})
	if err != nil || len(second.Cards) != 9 {
		t.Fatalf("segunda sessao: %d %v", len(second.Cards), err)
	}
	if terms(first) == terms(second) {
		t.Errorf("outra sessao deveria sortear outra ordem ou outras palavras: %s", terms(first))
	}
}

func terms(p domain.Pack) string {
	out := ""
	for _, c := range p.Cards {
		out += c.Content.Term + ","
	}
	return out
}
