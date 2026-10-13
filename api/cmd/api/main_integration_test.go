//go:build integration

package main

import (
	"context"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/keven-rdr/wordloop/api/internal/platform/db"
	"github.com/keven-rdr/wordloop/api/internal/platform/db/dbtest"
)

// `api migrate` contra um Postgres real cria a tabela do perfil.
func TestExecuteMigrateAppliesMigrations(t *testing.T) {
	url, _ := dbtest.Postgres(t)
	t.Setenv("DATABASE_URL", url)
	ctx := context.Background()
	if err := execute(ctx, []string{"migrate"}); err != nil {
		t.Fatal(err)
	}
	if got := realMain([]string{"migrate"}); got != 0 { // idempotente: a segunda execucao tambem sai com 0
		t.Fatalf("realMain migrate: esperava 0, veio %d", got)
	}
	pool, err := db.Open(ctx, url)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	var exists bool
	if err := pool.QueryRow(ctx, `SELECT to_regclass('public.app_user') IS NOT NULL`).Scan(&exists); err != nil || !exists {
		t.Fatalf("app_user ausente (err=%v)", err)
	}
}

// Sem argumento a API sobe com o banco real, fica pronta (readiness 200) e encerra limpo ao cancelar o contexto.
func TestExecuteServesReadyAndStopsCleanly(t *testing.T) {
	url, _ := dbtest.Postgres(t)
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	addr := ln.Addr().String()
	_ = ln.Close()
	t.Setenv("DATABASE_URL", url)
	t.Setenv("API_ADDR", addr)

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- execute(ctx, nil) }()
	waitListening(t, addr)

	resp, err := http.Get("http://" + addr + "/api/v1/health/ready")
	if err != nil {
		t.Fatal(err)
	}
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("readiness: esperava 200, veio %d", resp.StatusCode)
	}
	cancel()
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("encerramento com erro: %v", err)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("a API nao encerrou")
	}
}

// Porta ocupada: o servidor nao sobe e o erro chega a quem chamou (e ao codigo de saida).
func TestServeFailsWhenPortIsTaken(t *testing.T) {
	url, _ := dbtest.Postgres(t)
	busy, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = busy.Close() }()
	t.Setenv("DATABASE_URL", url)
	t.Setenv("API_ADDR", busy.Addr().String())
	if err := execute(context.Background(), nil); err == nil {
		t.Fatal("esperava erro com a porta ocupada")
	}
}
