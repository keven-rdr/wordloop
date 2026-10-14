package app

import "errors"

var (
	ErrNotFound     = errors.New("nao encontrado")
	ErrNoSession    = errors.New("sem sessao valida")
	ErrInvalidLogin = errors.New("login invalido")
	ErrMFARequired  = errors.New("a conta ligou a verificacao em duas etapas e este login nao a comprovou")
)
