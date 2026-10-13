// Package seal cifra e autentica pequenos segredos (cookie de estado do login, refresh token) com AES-256-GCM.
package seal

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
)

// KeySize e o tamanho da chave em bytes (AES-256).
const KeySize = 32

// ErrInvalid indica texto adulterado, truncado ou cifrado com outra chave. Nao revela o motivo.
var ErrInvalid = errors.New("valor selado invalido")

// Sealer cifra com uma chave fixa; seguro para uso concorrente.
type Sealer struct{ aead cipher.AEAD }

// New cria um Sealer. A chave vem de SESSION_KEY (32 bytes em base64).
func New(key []byte) (*Sealer, error) {
	if len(key) != KeySize {
		return nil, fmt.Errorf("a chave precisa ter %d bytes, tem %d", KeySize, len(key))
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	return &Sealer{aead: aead}, nil
}

// ParseKey le uma chave em base64 (padrao ou URL, com ou sem preenchimento).
func ParseKey(b64 string) ([]byte, error) {
	for _, enc := range []*base64.Encoding{base64.StdEncoding, base64.RawStdEncoding, base64.URLEncoding, base64.RawURLEncoding} {
		if key, err := enc.DecodeString(b64); err == nil && len(key) == KeySize {
			return key, nil
		}
	}
	return nil, fmt.Errorf("SESSION_KEY invalida: esperava %d bytes em base64", KeySize)
}

// Seal cifra plain e devolve texto seguro para cookie (base64url de nonce + cifrado).
func (s *Sealer) Seal(plain []byte) (string, error) {
	nonce := make([]byte, s.aead.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", fmt.Errorf("gerar nonce: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(s.aead.Seal(nonce, nonce, plain, nil)), nil
}

// Open decifra o que Seal produziu; qualquer alteracao devolve ErrInvalid.
func (s *Sealer) Open(sealed string) ([]byte, error) {
	raw, err := base64.RawURLEncoding.DecodeString(sealed)
	if err != nil || len(raw) < s.aead.NonceSize() {
		return nil, ErrInvalid
	}
	nonce, body := raw[:s.aead.NonceSize()], raw[s.aead.NonceSize():]
	plain, err := s.aead.Open(nil, nonce, body, nil)
	if err != nil {
		return nil, ErrInvalid
	}
	return plain, nil
}
