//go:build integration

package main

import (
	"bytes"
	"context"
	"errors"
	"net"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/keven-rdr/wordloop/api/internal/identity/accounts"
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

func TestCreateAdminPrintsAnInviteAndNeverTakesOverAnAccount(t *testing.T) {
	url, _ := dbtest.Postgres(t)
	t.Setenv("DATABASE_URL", url)
	t.Setenv("APP_PUBLIC_URL", "http://localhost:5173")
	ctx := context.Background()
	if err := execute(ctx, []string{"migrate"}); err != nil {
		t.Fatal(err)
	}
	var out bytes.Buffer
	if err := createAdmin(ctx, []string{"--email", "Voce@Exemplo.com", "--name", "Voce"}, &out); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out.String(), "Admin criado: voce@exemplo.com") || !strings.Contains(out.String(), "http://localhost:5173/invite#") {
		t.Fatalf("saida inesperada: %q", out.String())
	}
	out.Reset()
	if err := execute(ctx, []string{"create-admin", "--email", "voce@exemplo.com"}); err != nil {
		t.Fatalf("reemitir o convite de um admin sem como entrar deve funcionar: %v", err)
	}

	pool, err := db.Open(ctx, url)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	var unused int
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM auth_token WHERE purpose = 'invite' AND used_at IS NULL`).Scan(&unused); err != nil || unused != 1 {
		t.Fatalf("so o convite mais novo pode valer: %d (%v)", unused, err)
	}
	if _, err := pool.Exec(ctx, `INSERT INTO app_user (display_name, email, role) VALUES ('Aluno', 'aluno@exemplo.com', 'user')`); err != nil {
		t.Fatal(err)
	}
	if err := createAdmin(ctx, []string{"--email", "aluno@exemplo.com"}, &out); !errors.Is(err, accounts.ErrAccountExists) {
		t.Fatalf("nao pode transformar um aluno em admin: %v", err)
	}
	if out.Len() != 0 {
		t.Fatalf("nada deve ser impresso quando recusa: %q", out.String())
	}
}
