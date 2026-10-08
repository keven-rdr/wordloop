// Package openapi embute o contrato da API (openapi.yaml) no binario.
// O arquivo continua sendo a unica fonte da verdade: o servidor e o cliente sao gerados dele
// e a tela de documentacao o le daqui.
package openapi

import _ "embed"

// Spec e o conteudo do openapi.yaml.
//
//go:embed openapi.yaml
var Spec []byte
