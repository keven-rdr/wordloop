// Package migrations embute as migracoes SQL (goose) no binario: o deploy roda `api migrate` sem arquivos extras.
package migrations

import "embed"

// FS contem os arquivos NNNNN_nome.sql desta pasta.
//
//go:embed *.sql
var FS embed.FS
