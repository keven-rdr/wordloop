package main

import (
	"strings"
	"testing"

	httpapi "github.com/keven-rdr/wordloop/api/internal/identity/adapters/http"
)

func setWebAuthn(t *testing.T, rpID, origin string) {
	t.Helper()
	t.Setenv("WEBAUTHN_RP_ID", rpID)
	t.Setenv("WEBAUTHN_ORIGIN", origin)
}

func TestValidateWebAuthn(t *testing.T) {
	cases := map[string]struct {
		rpID, origin string
		ok           bool
	}{
		"tst":                       {"tst.wordloop.com.br", "https://tst.wordloop.com.br", true},
		"dev local":                 {"localhost", "http://localhost:5173", true},
		"dominio acima da origem":   {"wordloop.com.br", "https://tst.wordloop.com.br", true},
		"so o id":                   {"tst.wordloop.com.br", "", false},
		"so a origem":               {"", "https://tst.wordloop.com.br", false},
		"origem com caminho":        {"tst.wordloop.com.br", "https://tst.wordloop.com.br/app", false},
		"origem sem esquema":        {"tst.wordloop.com.br", "tst.wordloop.com.br", false},
		"origem de outro dominio":   {"wordloop.com.br", "https://wordloop.com.br.evil.example", false},
		"id que nao e da origem":    {"evil.example", "https://tst.wordloop.com.br", false},
		"origem com esquema errado": {"x.co", "ftp://x.co", false},
	}
	for name, c := range cases {
		err := validateWebAuthn(c.rpID, c.origin)
		if c.ok != (err == nil) {
			t.Errorf("%s: err = %v, esperava ok=%v", name, err, c.ok)
		}
	}
}

func TestBuildPasskeysIsOffByDefault(t *testing.T) {
	setWebAuthn(t, "", "")
	h := httpapi.Handler{}
	if err := buildPasskeys(&h, nil, nil, nil); err != nil || h.Passkeys != nil {
		t.Fatalf("sem configuracao as passkeys ficam desligadas: %v %v", err, h.Passkeys)
	}
}

func TestBuildPasskeysTurnsOnWithACompleteConfigurationAndSetsSecureOnHTTPS(t *testing.T) {
	setWebAuthn(t, "tst.wordloop.com.br", "https://tst.wordloop.com.br")
	h := httpapi.Handler{}
	if err := buildPasskeys(&h, nil, nil, nil); err != nil || h.Passkeys == nil || !h.Secure {
		t.Fatalf("passkeys=%v secure=%v err=%v", h.Passkeys, h.Secure, err)
	}
	setWebAuthn(t, "localhost", "http://localhost:5173")
	local := httpapi.Handler{}
	if err := buildPasskeys(&local, nil, nil, nil); err != nil || local.Passkeys == nil || local.Secure {
		t.Fatalf("em http local o cookie nao pode ser Secure: secure=%v err=%v", local.Secure, err)
	}
}

func TestBuildPasskeysRejectsHalfConfiguration(t *testing.T) {
	setWebAuthn(t, "tst.wordloop.com.br", "")
	if err := buildPasskeys(&httpapi.Handler{}, nil, nil, nil); err == nil || !strings.Contains(err.Error(), "WEBAUTHN_ORIGIN") {
		t.Fatalf("configuracao pela metade deve falhar dizendo o que falta: %v", err)
	}
}
