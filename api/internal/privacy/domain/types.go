// Package domain guarda as regras dos direitos da pessoa sobre os próprios dados (LGPD): quem pode exportar e
// quem pode se apagar, e quando. Não faz I/O.
package domain

import (
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

// RecentLogin e quanto tempo depois de entrar a pessoa ainda pode pedir a exclusao sem entrar de novo.
const RecentLogin = 15 * time.Minute

var (
	// ErrWrongConfirmation: o que a pessoa digitou nao e o e-mail da conta (ou a conta nao tem e-mail).
	ErrWrongConfirmation = errors.New("confirmacao diferente do e-mail da conta")
	// ErrAdminAccount: administrador nao se apaga pela API (so por comando no servidor), para nao trancar o sistema.
	ErrAdminAccount = errors.New("conta de administrador")
	// ErrRecentLoginRequired: a sessao foi aberta ha mais de RecentLogin; a pessoa precisa entrar de novo.
	ErrRecentLoginRequired = errors.New("login recente necessario")
	// ErrNotFound: a conta ja nao existe (apagada por outra chamada).
	ErrNotFound = errors.New("conta nao encontrada")
)

// Subject e quem pede: so a sessao decide, nunca o pedido.
type Subject struct {
	UserID          uuid.UUID
	Admin           bool
	AuthenticatedAt time.Time
}

// CheckErase diz se a pessoa pode se apagar agora: nao ser administradora, digitar o e-mail da conta e ter entrado
// ha pouco. A ordem e a do que se pode corrigir sem recomecar: primeiro quem nunca pode, depois o texto, por ultimo o login.
func CheckErase(who Subject, email, confirmation string, now time.Time) error {
	if who.Admin {
		return ErrAdminAccount
	}
	if email == "" || !strings.EqualFold(strings.TrimSpace(confirmation), email) {
		return ErrWrongConfirmation
	}
	if now.Sub(who.AuthenticatedAt) > RecentLogin {
		return ErrRecentLoginRequired
	}
	return nil
}

// Export e o que a pessoa leva embora: os dados, montados pelo banco, mais de quando sao.
type Export struct {
	GeneratedAt time.Time
	Data        json.RawMessage // objeto JSON com as secoes (perfil, vinculos, estudo, log...)
}

// ErasureReport resume o que foi apagado, sem nenhum dado pessoal.
type ErasureReport struct{ Reviews int }
