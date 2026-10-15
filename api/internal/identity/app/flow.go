package app

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"strings"
)

// Flow e o estado de um login em andamento; viaja selado no cookie wl_oidc.
type Flow struct {
	State    string `json:"s"`
	Nonce    string `json:"n"`
	Verifier string `json:"v"`
	ReturnTo string `json:"r"`
	Provider string `json:"p,omitempty"`
}

// randomToken devolve n bytes aleatorios em base64url (sem preenchimento).
func randomToken(r io.Reader, n int) (string, error) {
	if r == nil {
		r = rand.Reader
	}
	buf := make([]byte, n)
	if _, err := io.ReadFull(r, buf); err != nil {
		return "", fmt.Errorf("gerar aleatorio: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

// hashSession e o que vai ao banco no lugar do id da sessao.
func hashSession(id string) []byte {
	sum := sha256.Sum256([]byte(id))
	return sum[:]
}

// safeReturnTo so aceita caminho relativo do proprio app; qualquer outra coisa vira "/" (evita redirecionamento aberto).
func safeReturnTo(p string) string {
	if p == "" || len(p) > 512 || p[0] != '/' || strings.HasPrefix(p, "//") || strings.ContainsAny(p, "\\\r\n") {
		return "/"
	}
	return p
}
