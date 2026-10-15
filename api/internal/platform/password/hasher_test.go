package password_test

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"

	"github.com/keven-rdr/wordloop/api/internal/platform/password"
)

var bg = context.Background()

func TestHashThenVerifyAcceptsOnlyTheSamePassword(t *testing.T) {
	h := password.NewHasher(2)
	hash, err := h.Hash(bg, "uma senha bem longa")
	if err != nil || !strings.HasPrefix(hash, "$argon2id$v=19$m=19456,t=2,p=1$") {
		t.Fatalf("hash: %q %v", hash, err)
	}
	if ok, err := h.Verify(bg, "uma senha bem longa", hash); err != nil || !ok {
		t.Fatalf("senha certa: %v %v", ok, err)
	}
	if ok, err := h.Verify(bg, "uma senha bem longA", hash); err != nil || ok {
		t.Fatalf("senha errada: %v %v", ok, err)
	}
}

func TestTheSameSecretNeverProducesTheSameHash(t *testing.T) {
	h := password.NewHasher(2)
	a, _ := h.Hash(bg, "igual igual igual")
	b, _ := h.Hash(bg, "igual igual igual")
	if a == b || strings.Contains(a, "igual") {
		t.Fatalf("o sal deve variar e o hash nao pode conter a senha: %q %q", a, b)
	}
}

func TestVerifyRejectsMalformedOrAbusiveHashes(t *testing.T) {
	h := password.NewHasher(1)
	for _, bad := range []string{"", "x", "$argon2i$v=19$m=1,t=1,p=1$AAAAAAAAAAA$AAAAAAAAAAAAAAAAAAAAAA",
		"$argon2id$v=19$m=4294967295,t=1,p=1$AAAAAAAAAAA$AAAAAAAAAAAAAAAAAAAAAA", "$argon2id$v=19$m=19456,t=2,p=1$!!$!!"} {
		if ok, err := h.Verify(bg, "x", bad); ok || !errors.Is(err, password.ErrBadHash) {
			t.Errorf("%q: esperava ErrBadHash, veio %v %v", bad, ok, err)
		}
	}
}

func TestConcurrencyIsBoundedAndHonoursTheContext(t *testing.T) {
	h := password.NewHasher(1)
	var wg sync.WaitGroup
	for range 3 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := h.Hash(bg, "x"); err != nil {
				t.Error(err)
			}
		}()
	}
	wg.Wait()
	ctx, cancel := context.WithCancel(bg)
	cancel()
	if _, err := h.Hash(ctx, "x"); !errors.Is(err, context.Canceled) {
		t.Fatalf("contexto cancelado: esperava context.Canceled, veio %v", err)
	}
}
