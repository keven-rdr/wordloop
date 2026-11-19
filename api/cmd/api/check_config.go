package main

import (
	"errors"
	"io"
	"os"

	"github.com/keven-rdr/wordloop/api/internal/platform/readiness"
)

// checkConfig implementa `api check-config`: confere as variaveis do ambiente SEM abrir o banco nem subir o servidor
// e imprime o que esta ligado, desligado ou errado (nunca um segredo). O deploy roda antes de migrar: configuracao
// torta para o deploy com uma mensagem clara, em vez de deixar a API em crash-loop.
func checkConfig(out io.Writer) error {
	report := readiness.Check(os.Getenv)
	report.Write(out)
	if report.Failed() {
		return errors.New("configuracao com erro(s): corrija os itens marcados como erro")
	}
	return nil
}
