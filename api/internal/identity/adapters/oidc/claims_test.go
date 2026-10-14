package oidc

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"testing"
	"time"

	gooidc "github.com/coreos/go-oidc/v3/oidc"
	jose "github.com/go-jose/go-jose/v4"
)

func TestHasMFA(t *testing.T) {
	cases := map[string]struct {
		amr  []string
		want bool
	}{
		"sem amr (provedor nao informa)": {nil, false},
		"so senha":                       {[]string{"pwd"}, false},
		"hwk sozinho (1 fator)":          {[]string{"hwk"}, false},
		"otp sozinho (1 fator)":          {[]string{"otp"}, false},
		"senha e sms nao reconhecido":    {[]string{"pwd", "sms"}, false},
		"mfa":                            {[]string{"mfa"}, true},
		"senha e otp":                    {[]string{"pwd", "otp"}, true},
		"senha e chave de hardware":      {[]string{"pwd", "hwk"}, true},
		"maiuscula nao conta":            {[]string{"pwd", "OTP"}, false},
	}
	for name, c := range cases {
		if got := hasMFA(c.amr); got != c.want {
			t.Errorf("%s: hasMFA(%v) = %v, esperava %v", name, c.amr, got, c.want)
		}
	}
}

// idToken assina um ID token de verdade (RS256) e o valida como o go-oidc faz em producao.
func idToken(t *testing.T, extra map[string]any) *gooidc.IDToken {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	signer, err := jose.NewSigner(jose.SigningKey{Algorithm: jose.RS256, Key: key}, nil)
	if err != nil {
		t.Fatal(err)
	}
	claims := map[string]any{"iss": "https://idp.test", "aud": "wordloop", "sub": "abc", "exp": time.Now().Add(time.Hour).Unix(), "iat": time.Now().Unix()}
	for k, v := range extra {
		claims[k] = v
	}
	payload, _ := json.Marshal(claims)
	obj, err := signer.Sign(payload)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := obj.CompactSerialize()
	if err != nil {
		t.Fatal(err)
	}
	v := gooidc.NewVerifier("https://idp.test", &gooidc.StaticKeySet{PublicKeys: []crypto.PublicKey{&key.PublicKey}}, &gooidc.Config{ClientID: "wordloop"})
	idt, err := v.Verify(context.Background(), raw)
	if err != nil {
		t.Fatal(err)
	}
	return idt
}

func TestIdentityFromReadsNameAuthTimeAndMFA(t *testing.T) {
	authTime := time.Date(2026, 10, 14, 9, 30, 0, 0, time.UTC)
	got, err := identityFrom(idToken(t, map[string]any{
		"name": "Maria", "preferred_username": "maria", "amr": []string{"pwd", "otp"}, "auth_time": authTime.Unix(),
	}))
	if err != nil {
		t.Fatal(err)
	}
	if got.Issuer != "https://idp.test" || got.Subject != "abc" || got.DisplayName != "Maria" || !got.MFA || !got.AuthTime.Equal(authTime) {
		t.Fatalf("identidade inesperada: %+v", got)
	}
}

func TestIdentityFromFallsBackToUsernameAndToleratesOddClaimShapes(t *testing.T) {
	cases := map[string]map[string]any{
		"amr como texto":          {"preferred_username": "maria", "amr": "mfa", "auth_time": "1760434200"},
		"auth_time decimal":       {"preferred_username": "maria", "amr": []string{"mfa"}, "auth_time": 1760434200.0},
		"amr com tipo estranho":   {"preferred_username": "maria", "amr": 42, "auth_time": true},
		"sem os claims opcionais": {"preferred_username": "maria"},
	}
	for name, extra := range cases {
		got, err := identityFrom(idToken(t, extra))
		if err != nil {
			t.Errorf("%s: um claim estranho nao pode derrubar o login: %v", name, err)
			continue
		}
		if got.DisplayName != "maria" {
			t.Errorf("%s: sem name deve usar preferred_username, veio %q", name, got.DisplayName)
		}
	}
	got, _ := identityFrom(idToken(t, map[string]any{"amr": 42, "auth_time": "lixo"}))
	if got.MFA || !got.AuthTime.IsZero() {
		t.Fatalf("valor estranho nao prova 2 fatores nem define horario: %+v", got)
	}
}
