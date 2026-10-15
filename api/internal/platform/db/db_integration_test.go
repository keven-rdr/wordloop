//go:build integration

package db_test

import (
	"context"
	"testing"
	"testing/fstest"

	"github.com/keven-rdr/wordloop/api/internal/platform/db"
	"github.com/keven-rdr/wordloop/api/internal/platform/db/dbtest"
	"github.com/keven-rdr/wordloop/api/migrations"
)

func TestMigrateCreatesAppUserAndIsIdempotent(t *testing.T) {
	ctx := context.Background()
	url, _ := dbtest.Postgres(t)
	pool, err := db.Open(ctx, url)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()

	for i := 0; i < 2; i++ { // a segunda execucao nao pode falhar nem recriar nada
		if err := db.Migrate(ctx, pool, migrations.FS); err != nil {
			t.Fatalf("migracao %d: %v", i+1, err)
		}
	}
	var exists bool
	err = pool.QueryRow(ctx, `SELECT to_regclass('public.app_user') IS NOT NULL`).Scan(&exists)
	if err != nil || !exists {
		t.Fatalf("tabela app_user ausente (err=%v)", err)
	}
}

func TestOpenFailsOnInvalidOrUnreachableURL(t *testing.T) {
	ctx := context.Background()
	if _, err := db.Open(ctx, "isto nao e uma url"); err == nil {
		t.Fatal("esperava erro para URL invalida")
	}
	if _, err := db.Open(ctx, "postgres://u:p@127.0.0.1:1/x?sslmode=disable&connect_timeout=1"); err == nil {
		t.Fatal("esperava erro para banco inalcancavel")
	}
}

// Migracoes vazias e SQL invalido falham com erro, sem derrubar o processo.
func TestMigrateReportsBrokenMigrations(t *testing.T) {
	ctx := context.Background()
	url, _ := dbtest.Postgres(t)
	pool, err := db.Open(ctx, url)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()

	if err := db.Migrate(ctx, pool, fstest.MapFS{}); err == nil {
		t.Fatal("esperava erro sem nenhuma migracao")
	}
	const badSQL = "-- +goose Up\nSELECT * FROM tabela_que_nao_existe;\n"
	broken := fstest.MapFS{"00001_quebrada.sql": {Data: []byte(badSQL)}}
	if err := db.Migrate(ctx, pool, broken); err == nil {
		t.Fatal("esperava erro com SQL invalido")
	}
}

// Um seed grande deve ir num bloco StatementBegin/End: o goose o manda inteiro de uma vez (uma ida e volta ao banco)
// em vez de uma instrucao por vez (ver api/AGENTS.md, secao Banco). Aqui: o bloco com varias instrucoes roda inteiro.
func TestMigrateRunsAStatementBlockWithManyInsertsAtOnce(t *testing.T) {
	ctx := context.Background()
	url, _ := dbtest.Postgres(t)
	pool, err := db.Open(ctx, url)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()

	const seed = `-- +goose Up
CREATE TABLE palavra (id int PRIMARY KEY, texto text NOT NULL);
-- +goose StatementBegin
INSERT INTO palavra VALUES (1, 'um');
INSERT INTO palavra VALUES (2, 'dois');
INSERT INTO palavra VALUES (3, 'tres');
-- +goose StatementEnd
`
	if err := db.Migrate(ctx, pool, fstest.MapFS{"00001_seed.sql": {Data: []byte(seed)}}); err != nil {
		t.Fatal(err)
	}
	var n int
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM palavra`).Scan(&n); err != nil || n != 3 {
		t.Fatalf("esperava 3 linhas do bloco, veio %d (err=%v)", n, err)
	}
}
