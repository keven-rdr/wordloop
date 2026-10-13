//go:build integration

// Package dbtest sobe um PostgreSQL REAL (Testcontainers) para os testes de integracao. Exige Docker ligado.
package dbtest

import (
	"context"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

// Postgres devolve a URL de um banco novo e uma funcao que derruba o contêiner (para simular queda).
func Postgres(t *testing.T) (url string, stop func()) {
	t.Helper()
	ctx := context.Background()
	c, err := tcpostgres.Run(ctx, "postgres:17-alpine",
		tcpostgres.WithDatabase("wordloop_test"), tcpostgres.WithUsername("wordloop"), tcpostgres.WithPassword("teste"),
		testcontainers.WithWaitStrategy(wait.ForLog("database system is ready to accept connections").
			WithOccurrence(2).WithStartupTimeout(90*time.Second)),
	)
	if err != nil {
		t.Fatalf("subir o Postgres de teste (Docker ligado?): %v", err)
	}
	t.Cleanup(func() { _ = testcontainers.TerminateContainer(c) })
	url, err = c.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatal(err)
	}
	return url, func() { _ = testcontainers.TerminateContainer(c) }
}
