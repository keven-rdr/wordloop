package app

import "errors"

var (
	// ErrNotFound e devolvido pelos repositorios quando o registro nao existe.
	ErrNotFound = errors.New("nao encontrado")
	// ErrNoSession significa sem sessao valida (ausente, expirada, ociosa ou de usuario removido).
	ErrNoSession = errors.New("sem sessao valida")
	// ErrInvalidLogin significa login recusado ou expirado (state, code, nonce ou cookie de estado invalidos).
	ErrInvalidLogin = errors.New("login invalido")
	// ErrMFARequired significa que a conta e de administrador e o provedor nao provou 2 fatores; nenhuma sessao e criada.
	ErrMFARequired = errors.New("conta de administrador exige verificacao em duas etapas")
)
