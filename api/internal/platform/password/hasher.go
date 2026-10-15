package password

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

const (
	memoryKiB = 19 * 1024
	passes    = 2
	threads   = 1
	saltBytes = 16
	keyBytes  = 32
)

var ErrBadHash = errors.New("hash de senha invalido")

type Hasher struct{ slots chan struct{} }

func NewHasher(concurrency int) *Hasher {
	return &Hasher{slots: make(chan struct{}, max(concurrency, 1))}
}

func (h *Hasher) Hash(ctx context.Context, plain string) (string, error) {
	salt := make([]byte, saltBytes)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("gerar o sal: %w", err)
	}
	key, err := h.derive(ctx, plain, salt, memoryKiB, passes, threads, keyBytes)
	if err != nil {
		return "", err
	}
	enc := base64.RawStdEncoding
	return fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s", argon2.Version, memoryKiB, passes, threads,
		enc.EncodeToString(salt), enc.EncodeToString(key)), nil
}

func (h *Hasher) Verify(ctx context.Context, plain, encoded string) (bool, error) {
	p, err := parse(encoded)
	if err != nil {
		return false, err
	}
	key, err := h.derive(ctx, plain, p.salt, p.memory, p.passes, p.threads, keyBytes)
	if err != nil {
		return false, err
	}
	return subtle.ConstantTimeCompare(key, p.key) == 1, nil
}

func (h *Hasher) derive(ctx context.Context, plain string, salt []byte, memory, passes uint32, threads uint8, keyLen uint32) ([]byte, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	select {
	case h.slots <- struct{}{}:
		defer func() { <-h.slots }()
	case <-ctx.Done():
		return nil, ctx.Err()
	}
	return argon2.IDKey([]byte(plain), salt, passes, memory, threads, keyLen), nil
}

type params struct {
	memory, passes uint32
	threads        uint8
	salt, key      []byte
}

func parse(encoded string) (params, error) {
	parts := strings.Split(encoded, "$")
	if len(parts) != 6 || parts[1] != "argon2id" {
		return params{}, ErrBadHash
	}
	var version int
	if _, err := fmt.Sscanf(parts[2], "v=%d", &version); err != nil || version != argon2.Version {
		return params{}, ErrBadHash
	}
	p, err := parseCost(parts[3])
	if err != nil {
		return params{}, err
	}
	enc := base64.RawStdEncoding
	salt, err1 := enc.DecodeString(parts[4])
	key, err2 := enc.DecodeString(parts[5])
	if err1 != nil || err2 != nil || len(salt) < 8 || len(key) != keyBytes {
		return params{}, ErrBadHash
	}
	p.salt, p.key = salt, key
	return p, nil
}

func parseCost(s string) (params, error) {
	var p params
	var threads uint32
	if _, err := fmt.Sscanf(s, "m=%d,t=%d,p=%d", &p.memory, &p.passes, &threads); err != nil {
		return params{}, ErrBadHash
	}
	if p.memory == 0 || p.memory > 256*1024 || p.passes == 0 || p.passes > 10 || threads == 0 || threads > 8 {
		return params{}, ErrBadHash
	}
	p.threads = uint8(threads)
	return p, nil
}
