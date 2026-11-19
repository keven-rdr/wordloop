package main

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

func TestCheckConfigPassesAGoodDevelopmentConfigAndNeverPrintsTheSecret(t *testing.T) {
	t.Setenv("APP_ENV", "dev")
	t.Setenv("DATABASE_URL", "postgres://canary-user:canary-db-value@localhost/app?sslmode=disable")
	var out bytes.Buffer
	if err := checkConfig(&out); err != nil {
		t.Fatalf("dev so com banco deve passar: %v\n%s", err, out.String())
	}
	if strings.Contains(out.String(), "canary-db-value") || !strings.Contains(out.String(), "DATABASE_URL definida") {
		t.Fatalf("saida inesperada:\n%s", out.String())
	}
}

func TestCheckConfigFailsWithAWrongProductionConfigAndAnswersThroughExecute(t *testing.T) {
	t.Setenv("APP_ENV", "prd")
	t.Setenv("DATABASE_URL", "postgres://u:p@db.example/app?sslmode=require")
	t.Setenv("APP_PUBLIC_URL", "http://app.exemplo.com.br") // http em producao
	var out bytes.Buffer
	if err := checkConfig(&out); err == nil || !strings.Contains(out.String(), "erro") {
		t.Fatalf("esperava erro, veio %v\n%s", err, out.String())
	}
	if err := execute(context.Background(), []string{"check-config"}); err == nil {
		t.Fatal("o subcomando check-config deve devolver o erro (codigo de saida 1)")
	}
}
