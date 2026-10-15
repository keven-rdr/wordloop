package app

import "errors"

var (
	ErrNotFound         = errors.New("nao encontrado")
	ErrNoSession        = errors.New("sem sessao valida")
	ErrInvalidLogin     = errors.New("login invalido")
	ErrUnknownProvider  = errors.New("provedor de login desconhecido ou desligado")
	ErrEmailNotVerified = errors.New("o provedor nao informou um e-mail confirmado")
	ErrMFARequired      = errors.New("a conta ligou a verificacao em duas etapas e este login nao a comprovou")
)
