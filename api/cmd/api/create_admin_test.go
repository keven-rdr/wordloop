package main

import (
	"strings"
	"testing"
)

func TestParseCreateAdmin(t *testing.T) {
	got, err := parseCreateAdmin([]string{"--email", "a@b.co", "--name", "Keven"})
	if err != nil || got.email != "a@b.co" || got.name != "Keven" {
		t.Fatalf("parse = %+v, %v", got, err)
	}
	for name, args := range map[string][]string{
		"sem argumentos":    nil,
		"sem e-mail":        {"--name", "Keven"},
		"flag desconhecida": {"--email", "a@b.co", "--senha", "x"},
		"e-mail sem valor":  {"--email"},
	} {
		_, err := parseCreateAdmin(args)
		if err == nil {
			t.Errorf("%s: esperava erro", name)
			continue
		}
		if !strings.Contains(err.Error(), "create-admin --email") {
			t.Errorf("%s: o erro deve mostrar como usar: %v", name, err)
		}
	}
}
