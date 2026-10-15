package mail_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/keven-rdr/wordloop/api/internal/platform/mail"
)

func TestBrevoSendsTheEmailWithTheKeyInTheHeaderOnly(t *testing.T) {
	var got map[string]any
	var key, path, contentType string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key, path, contentType = r.Header.Get("api-key"), r.URL.Path, r.Header.Get("Content-Type")
		_ = json.NewDecoder(r.Body).Decode(&got)
		w.WriteHeader(http.StatusCreated)
	}))
	defer srv.Close()
	b, err := mail.NewBrevo("chave-de-teste", "nao-responda@wordloop.com.br", "wordloop")
	if err != nil {
		t.Fatal(err)
	}
	b.BaseURL = srv.URL
	if err := b.Send(context.Background(), mail.Message{To: "ana@exemplo.com", Subject: "Confirme", Text: "link"}); err != nil {
		t.Fatal(err)
	}
	if path != "/v3/smtp/email" || key != "chave-de-teste" || contentType != "application/json" {
		t.Fatalf("pedido: %s %q %q", path, key, contentType)
	}
	sender := got["sender"].(map[string]any)
	to := got["to"].([]any)[0].(map[string]any)
	if sender["email"] != "nao-responda@wordloop.com.br" || to["email"] != "ana@exemplo.com" || got["subject"] != "Confirme" || got["textContent"] != "link" {
		t.Fatalf("corpo: %+v", got)
	}
	if strings.Contains(strings.ToLower(toJSON(got)), "chave-de-teste") {
		t.Fatal("a chave nunca vai no corpo")
	}
}

func toJSON(v any) string {
	b, _ := json.Marshal(v)
	return string(b)
}

func TestBrevoErrorsCarryTheStatusButNeverTheKey(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"message":"Key not found"}`))
	}))
	defer srv.Close()
	b, _ := mail.NewBrevo("chave-secreta", "a@b.co", "")
	b.BaseURL = srv.URL
	err := b.Send(context.Background(), mail.Message{To: "ana@exemplo.com"})
	if err == nil || !strings.Contains(err.Error(), "401") || strings.Contains(err.Error(), "chave-secreta") {
		t.Fatalf("erro: %v", err)
	}
}

func TestBrevoNeedsAKeyAndASender(t *testing.T) {
	for _, in := range [][2]string{{"", "a@b.co"}, {"k", ""}} {
		if _, err := mail.NewBrevo(in[0], in[1], ""); !errors.Is(err, mail.ErrNotConfigured) {
			t.Errorf("%v: esperava ErrNotConfigured, veio %v", in, err)
		}
	}
}

func TestBrevoReportsNetworkFailuresAndHonoursTheContext(t *testing.T) {
	b, _ := mail.NewBrevo("k", "a@b.co", "")
	b.BaseURL = "http://127.0.0.1:1"
	if err := b.Send(context.Background(), mail.Message{To: "x@y.co"}); err == nil {
		t.Fatal("esperava erro de rede")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := b.Send(ctx, mail.Message{To: "x@y.co"}); err == nil {
		t.Fatal("contexto cancelado deve falhar")
	}
}

func TestLogDriverNeverFails(t *testing.T) {
	if err := (mail.Log{}).Send(context.Background(), mail.Message{To: "x@y.co", Subject: "s", Text: "t"}); err != nil {
		t.Fatal(err)
	}
}
