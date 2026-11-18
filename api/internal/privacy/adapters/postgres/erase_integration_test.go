//go:build integration

package postgres_test

import (
	"context"
	"errors"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/keven-rdr/wordloop/api/internal/privacy/domain"
)

func mustUUID(t *testing.T, s string) uuid.UUID {
	t.Helper()
	id, err := uuid.Parse(s)
	if err != nil {
		t.Fatal(err)
	}
	return id
}

// tablesWithUserID lista toda tabela do schema public que tem uma coluna user_id: se amanha nascer uma tabela nova com
// dados da pessoa, este teste ja cobre.
func (f fixture) tablesWithUserID(t *testing.T) []string {
	t.Helper()
	rows, err := f.pool.Query(context.Background(), `SELECT table_name FROM information_schema.columns
		WHERE table_schema = 'public' AND column_name = 'user_id' ORDER BY table_name`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			t.Fatal(err)
		}
		out = append(out, name)
	}
	return out
}

func TestEraseLeavesNothingOfThePersonInAnyTable(t *testing.T) {
	f := setup(t)
	ctx := context.Background()
	ana := f.person(t, "Ana", "ana@example.com")
	bia := f.person(t, "Bia", "bia@example.com")

	tables := f.tablesWithUserID(t)
	for _, want := range []string{"auth_identity", "auth_session", "auth_token", "card", "goal_cycle", "passkey_credential", "review_log", "study_day", "study_session", "user_settings"} {
		if !slices.Contains(tables, want) {
			t.Fatalf("o teste deveria conhecer a tabela %s (achou %v)", want, tables)
		}
	}
	biaBefore := map[string]int{}
	for _, table := range tables {
		biaBefore[table] = f.count(t, `SELECT count(*) FROM `+table+` WHERE user_id = $1`, bia)
	}

	report, err := f.store.Erase(ctx, ana, time.Now())
	if err != nil || report.Reviews != 2 {
		t.Fatalf("apagar: %+v %v", report, err)
	}

	for _, table := range tables {
		if n := f.count(t, `SELECT count(*) FROM `+table+` WHERE user_id = $1`, ana); n != 0 {
			t.Errorf("%s ainda tem %d linhas da pessoa", table, n)
		}
		if n := f.count(t, `SELECT count(*) FROM `+table+` WHERE user_id = $1`, bia); n != biaBefore[table] {
			t.Errorf("%s: os dados da outra pessoa mudaram (%d -> %d)", table, biaBefore[table], n)
		}
	}
	if n := f.count(t, `SELECT count(*) FROM app_user WHERE id = $1`, ana); n != 0 {
		t.Error("o usuario ainda existe")
	}
	if n := f.count(t, `SELECT count(*) FROM app_user WHERE id = $1`, bia); n != 1 {
		t.Error("a outra pessoa nao pode ser tocada")
	}
	if n := f.count(t, `SELECT count(*) FROM review_void`); n != 1 {
		t.Errorf("as correcoes de Ana saem com o log, as de Bia ficam: %d", n)
	}
}

func TestEraseRecordsOnlyThatItHappened(t *testing.T) {
	f := setup(t)
	ana := f.person(t, "Ana", "ana@example.com")
	if _, err := f.store.Erase(context.Background(), ana, time.Now()); err != nil {
		t.Fatal(err)
	}
	if n := f.count(t, `SELECT review_count FROM account_erasure`); n != 2 {
		t.Errorf("registro da exclusao: %d", n)
	}
	rows, err := f.pool.Query(context.Background(), `SELECT column_name FROM information_schema.columns WHERE table_name = 'account_erasure' ORDER BY column_name`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	var cols []string
	for rows.Next() {
		var c string
		if err := rows.Scan(&c); err != nil {
			t.Fatal(err)
		}
		cols = append(cols, c)
	}
	if !slices.Equal(cols, []string{"erased_at", "id", "review_count"}) {
		t.Errorf("o registro nao pode guardar quem foi apagado: %v", cols)
	}
}

func TestEraseTwiceOrOfAStrangerFindsNothingAndWritesNoRecord(t *testing.T) {
	f := setup(t)
	ctx := context.Background()
	ana := f.person(t, "Ana", "ana@example.com")
	if _, err := f.store.Erase(ctx, ana, time.Now()); err != nil {
		t.Fatal(err)
	}
	if _, err := f.store.Erase(ctx, ana, time.Now()); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("2a vez: %v", err)
	}
	if _, err := f.store.Erase(ctx, uuid.New(), time.Now()); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("desconhecida: %v", err)
	}
	if n := f.count(t, `SELECT count(*) FROM account_erasure`); n != 1 {
		t.Errorf("so a exclusao que aconteceu fica registrada: %d", n)
	}
	if _, err := f.store.Email(ctx, ana); !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("e-mail de conta apagada: %v", err)
	}
}

func TestEmailOfTheAccount(t *testing.T) {
	f := setup(t)
	ana := f.person(t, "Ana", "ana@example.com")
	if got, err := f.store.Email(context.Background(), ana); err != nil || got != "ana@example.com" {
		t.Errorf("e-mail: %q %v", got, err)
	}
	var noEmail uuid.UUID
	if err := f.pool.QueryRow(context.Background(), `INSERT INTO app_user (display_name) VALUES ('Sem e-mail') RETURNING id`).Scan(&noEmail); err != nil {
		t.Fatal(err)
	}
	if got, err := f.store.Email(context.Background(), noEmail); err != nil || got != "" {
		t.Errorf("conta sem e-mail: %q %v", got, err)
	}
}

// A excecao do gatilho e estreita: fora da exclusao, o log continua imutavel.
func TestTheReviewLogStaysImmutableOutsideTheErasure(t *testing.T) {
	f := setup(t)
	ctx := context.Background()
	ana := f.person(t, "Ana", "ana@example.com")
	cases := []struct {
		name, sql string
		declared  string // o que a transacao declara em wordloop.erase_user (vazio = nada)
	}{
		{"UPDATE", `UPDATE review_log SET grade = 1`, ""},
		{"DELETE de uma pessoa", `DELETE FROM review_log WHERE user_id = $1`, ""},
		{"DELETE de todo o log", `DELETE FROM review_log`, ""},
		{"UPDATE mesmo declarando a pessoa", `UPDATE review_log SET grade = 1 WHERE user_id = $1`, ana.String()},
		{"DELETE declarando OUTRA pessoa", `DELETE FROM review_log WHERE user_id = $1`, uuid.New().String()},
	}
	for _, c := range cases {
		tx, err := f.pool.Begin(ctx)
		if err != nil {
			t.Fatal(err)
		}
		if c.declared != "" {
			if _, err := tx.Exec(ctx, `SELECT set_config('wordloop.erase_user', $1, true)`, c.declared); err != nil {
				t.Fatal(err)
			}
		}
		var execErr error
		if strings.Contains(c.sql, "$1") {
			_, execErr = tx.Exec(ctx, c.sql, ana)
		} else {
			_, execErr = tx.Exec(ctx, c.sql)
		}
		_ = tx.Rollback(ctx)
		if execErr == nil {
			t.Errorf("%s: o log e append-only, mas aceitou", c.name)
		}
	}
}
