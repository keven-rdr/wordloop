package main

import (
	"fmt"
	"io"
	"log/slog"
	"net"
	"strings"
)

const (
	bannerRule = "----------------------------------------------------------"
	envDev     = "dev"
	apiBase    = "/api/v1"
)

// startupInfo reune o que o aviso de inicializacao mostra.
type startupInfo struct {
	Env, Version, Commit string
	DocsEnabled          bool
}

// announce registra que a API passou a aceitar conexoes (log estruturado, em qualquer ambiente).
// Em dev, imprime tambem um quadro com as URLs de acesso, para abrir direto no navegador.
func announce(w io.Writer, info startupInfo, addr net.Addr) {
	slog.Info("api pronta", "addr", addr.String(), "version", info.Version, "commit", info.Commit, "env", info.Env)
	if info.Env == envDev {
		_, _ = fmt.Fprintln(w, devBanner(info, addr))
	}
}

func devBanner(info startupInfo, addr net.Addr) string {
	base := "http://" + net.JoinHostPort("localhost", portOf(addr)) + apiBase
	lines := []string{
		bannerRule,
		fmt.Sprintf("\twordloop API %s (%s) rodando em modo DEV. Enderecos:", info.Version, info.Commit),
		"\tAPI:     " + base,
	}
	if info.DocsEnabled {
		lines = append(lines, "\tSwagger: "+base+"/docs/")
	}
	lines = append(lines, bannerRule)
	return "\n" + strings.Join(lines, "\n")
}

// portOf devolve a porta em que o servidor escuta (o endereco pode ser ":8080", "[::]:8080", etc.).
func portOf(addr net.Addr) string {
	_, port, err := net.SplitHostPort(addr.String())
	if err != nil {
		return addr.String()
	}
	return port
}
