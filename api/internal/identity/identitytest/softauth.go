package identitytest

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"

	"github.com/fxamacker/cbor/v2"
)

const (
	credIDLen = 32
	aaguidLen = 16
	flagUP    = 0x01
	flagUV    = 0x04
	flagBE    = 0x08
	flagBS    = 0x10
	flagAT    = 0x40
)

type SoftAuthenticator struct {
	Origin     string
	RPID       string
	NoUV       bool
	BackupFlag bool
	key        *ecdsa.PrivateKey
	credID     []byte
	counter    uint32
}

func NewSoftAuthenticator(origin, rpID string) *SoftAuthenticator {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		panic(err)
	}
	id := make([]byte, credIDLen)
	if _, err := rand.Read(id); err != nil {
		panic(err)
	}
	return &SoftAuthenticator{Origin: origin, RPID: rpID, key: key, credID: id}
}

func (a *SoftAuthenticator) CredentialID() []byte { return a.credID }

func (a *SoftAuthenticator) SetCounter(n uint32) { a.counter = n }

func b64(b []byte) string { return base64.RawURLEncoding.EncodeToString(b) }

func (a *SoftAuthenticator) flags(extra byte) byte {
	f := byte(flagUP) | extra
	if !a.NoUV {
		f |= flagUV
	}
	if a.BackupFlag {
		f |= flagBE | flagBS
	}
	return f
}

func (a *SoftAuthenticator) clientData(typ, challenge string) []byte {
	out, err := json.Marshal(map[string]any{"type": typ, "challenge": challenge, "origin": a.Origin, "crossOrigin": false})
	if err != nil {
		panic(err)
	}
	return out
}

func (a *SoftAuthenticator) authData(withKey bool) []byte {
	rp := sha256.Sum256([]byte(a.RPID))
	extra := byte(0)
	if withKey {
		extra = flagAT
	}
	out := append(rp[:], a.flags(extra))
	out = binary.BigEndian.AppendUint32(out, a.counter)
	if !withKey {
		return out
	}
	out = append(out, make([]byte, aaguidLen)...)
	out = binary.BigEndian.AppendUint16(out, credIDLen)
	out = append(out, a.credID...)
	point, err := a.key.PublicKey.Bytes()
	if err != nil {
		panic(err)
	}
	x, y := point[1:33], point[33:65]
	cose, err := cbor.Marshal(map[int]any{1: 2, 3: -7, -1: 1, -2: x, -3: y})
	if err != nil {
		panic(err)
	}
	return append(out, cose...)
}

func challengeOf(options []byte) (string, error) {
	var o struct {
		Challenge string `json:"challenge"`
	}
	if err := json.Unmarshal(options, &o); err != nil || o.Challenge == "" {
		return "", fmt.Errorf("opcoes sem challenge: %v", err)
	}
	return o.Challenge, nil
}

func (a *SoftAuthenticator) Create(options []byte) ([]byte, error) {
	challenge, err := challengeOf(options)
	if err != nil {
		return nil, err
	}
	obj, err := cbor.Marshal(map[string]any{"fmt": "none", "attStmt": map[string]any{}, "authData": a.authData(true)})
	if err != nil {
		return nil, err
	}
	return json.Marshal(map[string]any{
		"id": b64(a.credID), "rawId": b64(a.credID), "type": "public-key", "clientExtensionResults": map[string]any{},
		"response": map[string]any{
			"clientDataJSON":    b64(a.clientData("webauthn.create", challenge)),
			"attestationObject": b64(obj),
			"transports":        []string{"internal"},
		},
	})
}

func (a *SoftAuthenticator) Get(options []byte, userHandle []byte) ([]byte, error) {
	challenge, err := challengeOf(options)
	if err != nil {
		return nil, err
	}
	a.counter++
	data, client := a.authData(false), a.clientData("webauthn.get", challenge)
	hash := sha256.Sum256(client)
	digest := sha256.Sum256(append(append([]byte{}, data...), hash[:]...))
	sig, err := ecdsa.SignASN1(rand.Reader, a.key, digest[:])
	if err != nil {
		return nil, err
	}
	return json.Marshal(map[string]any{
		"id": b64(a.credID), "rawId": b64(a.credID), "type": "public-key", "clientExtensionResults": map[string]any{},
		"response": map[string]any{
			"clientDataJSON": b64(client), "authenticatorData": b64(data), "signature": b64(sig), "userHandle": b64(userHandle),
		},
	})
}
