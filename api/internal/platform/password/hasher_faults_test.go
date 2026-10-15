package password_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/keven-rdr/wordloop/api/internal/platform/password"
)

const (
	validSalt = "AAAAAAAAAAAAAAAAAAAAAA"                      // 16 bytes
	validKey  = "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA" // 32 bytes
)

func encoded(version, cost, salt, key string) string {
	return strings.Join([]string{"", "argon2id", version, cost, salt, key}, "$")
}

func TestVerifyRejectsEveryMalformedPartOfTheHash(t *testing.T) {
	h := password.NewHasher(1)
	cases := map[string]string{
		"versao antiga":        encoded("v=18", "m=19456,t=2,p=1", validSalt, validKey),
		"versao nao numerica":  encoded("v=x", "m=19456,t=2,p=1", validSalt, validKey),
		"custo ilegivel":       encoded("v=19", "m=x,t=y,p=z", validSalt, validKey),
		"memoria zero":         encoded("v=19", "m=0,t=2,p=1", validSalt, validKey),
		"passadas demais":      encoded("v=19", "m=19456,t=11,p=1", validSalt, validKey),
		"paralelismo demais":   encoded("v=19", "m=19456,t=2,p=9", validSalt, validKey),
		"sal curto":            encoded("v=19", "m=19456,t=2,p=1", "AAAA", validKey),
		"chave de outro porte": encoded("v=19", "m=19456,t=2,p=1", validSalt, "AAAA"),
		"campos a menos":       "$argon2id$v=19$m=19456,t=2,p=1$" + validSalt,
	}
	for name, bad := range cases {
		if ok, err := h.Verify(bg, "x", bad); ok || !errors.Is(err, password.ErrBadHash) {
			t.Errorf("%s: esperava ErrBadHash, veio %v %v", name, ok, err)
		}
	}
}

func TestVerifyAndHashStopWhenTheContextIsAlreadyCancelled(t *testing.T) {
	h := password.NewHasher(1)
	hash, err := h.Hash(bg, "uma senha bem longa")
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(bg)
	cancel()
	if ok, err := h.Verify(ctx, "uma senha bem longa", hash); ok || !errors.Is(err, context.Canceled) {
		t.Fatalf("Verify com contexto cancelado: %v %v", ok, err)
	}
	if _, err := h.Hash(ctx, "outra"); !errors.Is(err, context.Canceled) {
		t.Fatalf("Hash com contexto cancelado: %v", err)
	}
}
