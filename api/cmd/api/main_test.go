package main

import (
	"context"
	"net"
	"net/http"
	"testing"
	"time"
)

func TestRunShutsDownGracefullyWhenContextIsCancelled(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	addr := ln.Addr().String()
	_ = ln.Close()

	srv := &http.Server{Addr: addr, Handler: http.NewServeMux(), ReadHeaderTimeout: time.Second}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	ready := make(chan net.Addr, 1)
	go func() { done <- run(ctx, srv, func(a net.Addr) { ready <- a }) }()

	waitListening(t, addr)
	if got := (<-ready).String(); got != addr {
		t.Fatalf("onReady recebeu %s, esperava %s", got, addr)
	}
	cancel()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("run devolveu erro no encerramento: %v", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("run nao encerrou depois do cancelamento")
	}
}

func TestRunReturnsListenError(t *testing.T) {
	srv := &http.Server{Addr: "256.0.0.1:0", ReadHeaderTimeout: time.Second} // endereco invalido
	if err := run(context.Background(), srv, func(net.Addr) {}); err == nil {
		t.Fatal("esperava erro de listen")
	}
}

func waitListening(t *testing.T, addr string) {
	t.Helper()
	for range 100 {
		if c, err := net.Dial("tcp", addr); err == nil {
			_ = c.Close()
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatal("servidor nao subiu")
}
