package token

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
)

const Bytes = 32

func New(r io.Reader) (string, []byte, error) {
	if r == nil {
		r = rand.Reader
	}
	buf := make([]byte, Bytes)
	if _, err := io.ReadFull(r, buf); err != nil {
		return "", nil, fmt.Errorf("gerar o token: %w", err)
	}
	t := base64.RawURLEncoding.EncodeToString(buf)
	return t, Hash(t), nil
}

func Hash(t string) []byte {
	sum := sha256.Sum256([]byte(t))
	return sum[:]
}
