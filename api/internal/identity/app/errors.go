package app

import "errors"

var (
	ErrNotFound     = errors.New("nao encontrado")
	ErrNoSession    = errors.New("sem sessao valida")
	ErrInvalidLogin = errors.New("login invalido")
	ErrMFARequired  = errors.New("conta de administrador exige verificacao em duas etapas")
)
