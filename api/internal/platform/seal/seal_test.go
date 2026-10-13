package seal_test

import (
	"bytes"
	"encoding/base64"
	"errors"
	"testing"

	"github.com/keven-rdr/wordloop/api/internal/platform/seal"
)

func newSealer(t *testing.T) *seal.Sealer {
	t.Helper()
	s, err := seal.New(bytes.Repeat([]byte{7}, seal.KeySize))
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func TestSealOpenRoundTrip(t *testing.T) {
	s := newSealer(t)
	sealed, err := s.Seal([]byte("segredo"))
	if err != nil {
		t.Fatal(err)
	}
	got, err := s.Open(sealed)
	if err != nil || string(got) != "segredo" {
		t.Fatalf("Open = %q, %v", got, err)
	}
	other, _ := s.Seal([]byte("segredo"))
	if sealed == other {
		t.Fatal("o mesmo texto selado duas vezes nao pode ser igual (nonce aleatorio)")
	}
}

func TestOpenRejectsTamperedOrForeignValues(t *testing.T) {
	s := newSealer(t)
	sealed, _ := s.Seal([]byte("segredo"))
	raw, _ := base64.RawURLEncoding.DecodeString(sealed)
	raw[len(raw)-1] ^= 1
	foreign, _ := seal.New(bytes.Repeat([]byte{9}, seal.KeySize))
	for name, in := range map[string]struct {
		s     *seal.Sealer
		value string
	}{
		"adulterado":  {s, base64.RawURLEncoding.EncodeToString(raw)},
		"outra chave": {foreign, sealed},
		"vazio":       {s, ""},
		"nao base64":  {s, "***"},
		"curto":       {s, "AAAA"},
	} {
		if _, err := in.s.Open(in.value); !errors.Is(err, seal.ErrInvalid) {
			t.Errorf("%s: esperava ErrInvalid, veio %v", name, err)
		}
	}
}

func TestNewAndParseKeyValidateSize(t *testing.T) {
	if _, err := seal.New([]byte("curta")); err == nil {
		t.Fatal("esperava erro para chave curta")
	}
	key := bytes.Repeat([]byte{1}, seal.KeySize)
	for _, enc := range []*base64.Encoding{base64.StdEncoding, base64.RawURLEncoding} {
		got, err := seal.ParseKey(enc.EncodeToString(key))
		if err != nil || !bytes.Equal(got, key) {
			t.Fatalf("ParseKey: %v", err)
		}
	}
	if _, err := seal.ParseKey("nao-e-uma-chave"); err == nil {
		t.Fatal("esperava erro para chave invalida")
	}
}
