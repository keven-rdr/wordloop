package token_test

import (
	"bytes"
	"errors"
	"testing"

	"github.com/keven-rdr/wordloop/api/internal/platform/token"
)

type brokenReader struct{}

func (brokenReader) Read([]byte) (int, error) { return 0, errors.New("sem entropia") }

func TestNewReturnsA256BitTokenAndItsHash(t *testing.T) {
	tok, hash, err := token.New(nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(tok) != 43 || len(hash) != 32 || !bytes.Equal(hash, token.Hash(tok)) {
		t.Fatalf("token=%q hash=%x", tok, hash)
	}
	other, _, _ := token.New(nil)
	if other == tok {
		t.Fatal("dois tokens nao podem ser iguais")
	}
}

func TestNewFailsWithoutRandomness(t *testing.T) {
	if _, _, err := token.New(brokenReader{}); err == nil {
		t.Fatal("esperava erro sem aleatoriedade")
	}
}

func TestHashIsStableAndDependsOnTheToken(t *testing.T) {
	if !bytes.Equal(token.Hash("a"), token.Hash("a")) || bytes.Equal(token.Hash("a"), token.Hash("b")) {
		t.Fatal("hash deve ser deterministico e distinguir tokens")
	}
}
