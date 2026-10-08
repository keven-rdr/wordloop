package main

import (
	"bytes"
	"net"
	"strings"
	"testing"
)

func tcpAddr(t *testing.T, s string) net.Addr {
	t.Helper()
	a, err := net.ResolveTCPAddr("tcp", s)
	if err != nil {
		t.Fatal(err)
	}
	return a
}

func TestAnnounceInDevPrintsUrls(t *testing.T) {
	var buf bytes.Buffer
	announce(&buf, startupInfo{Env: "dev", Version: "0.1.0", Commit: "abc1234", DocsEnabled: true}, tcpAddr(t, "[::]:8080"))
	out := buf.String()
	for _, want := range []string{
		"0.1.0 (abc1234)",
		"API:     http://localhost:8080/api/v1",
		"Swagger: http://localhost:8080/api/v1/docs/",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("faltou %q no quadro:\n%s", want, out)
		}
	}
}

func TestAnnounceInDevWithoutDocsOmitsSwagger(t *testing.T) {
	var buf bytes.Buffer
	announce(&buf, startupInfo{Env: "dev"}, tcpAddr(t, "127.0.0.1:18080"))
	out := buf.String()
	if strings.Contains(out, "Swagger") {
		t.Errorf("nao devia citar o Swagger com a documentacao desligada:\n%s", out)
	}
	if !strings.Contains(out, "http://localhost:18080/api/v1") {
		t.Errorf("faltou a URL da API:\n%s", out)
	}
}

func TestAnnounceOutsideDevPrintsNoBanner(t *testing.T) {
	for _, env := range []string{"tst", "prd"} {
		var buf bytes.Buffer
		announce(&buf, startupInfo{Env: env, DocsEnabled: true}, tcpAddr(t, ":8080"))
		if buf.Len() != 0 {
			t.Errorf("%s: nao devia imprimir o quadro, veio %q", env, buf.String())
		}
	}
}
