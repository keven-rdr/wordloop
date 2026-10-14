package main

import (
	"bytes"
	"context"
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

func TestCreateAdminFailsBeforeTouchingTheDatabaseOnBadArguments(t *testing.T) {
	var out bytes.Buffer
	if err := createAdmin(context.Background(), nil, &out); err == nil {
		t.Fatal("sem --email deve falhar")
	}
	if out.Len() != 0 {
		t.Fatalf("nada deve ser impresso em caso de erro: %q", out.String())
	}
}

func TestCreateAdminNeedsTheDatabase(t *testing.T) {
	t.Setenv("DATABASE_URL", "")
	var out bytes.Buffer
	err := createAdmin(context.Background(), []string{"--email", "a@b.co"}, &out)
	if err == nil || !strings.Contains(err.Error(), "DATABASE_URL") {
		t.Fatalf("esperava erro pedindo a DATABASE_URL, veio %v", err)
	}
}

func TestExecuteRoutesCreateAdmin(t *testing.T) {
	if err := execute(context.Background(), []string{"create-admin"}); err == nil {
		t.Fatal("create-admin sem --email deve falhar")
	}
}
