// Package version guarda a identidade do build, preenchida por -ldflags -X no Dockerfile/CI.
package version

var (
	// Version vem de `git describe` (ex.: 0.1.0 ou 0.1.0-5-gabc123).
	Version = "dev"
	// Commit e o hash curto do commit.
	Commit = "unknown"
	// Env e o ambiente padrao (dev, tst, prd); em execucao vem de APP_ENV.
	Env = "dev"
)
