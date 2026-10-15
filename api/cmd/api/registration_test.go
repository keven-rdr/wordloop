package main

import (
	"testing"

	httpapi "github.com/keven-rdr/wordloop/api/internal/identity/adapters/http"
	"github.com/keven-rdr/wordloop/api/internal/platform/mail"
)

func clearMailEnv(t *testing.T) {
	t.Helper()
	for _, key := range []string{"MAIL_DRIVER", "BREVO_API_KEY", "MAIL_FROM", "MAIL_FROM_NAME", "APP_PUBLIC_URL"} {
		t.Setenv(key, "")
	}
}

func TestBuildMailPicksTheDriverFromTheEnvironment(t *testing.T) {
	clearMailEnv(t)
	if sender, err := buildMail(); sender != nil || err != nil {
		t.Fatalf("sem configuracao o e-mail fica desligado: %v %v", sender, err)
	}
	t.Setenv("MAIL_DRIVER", "log")
	t.Setenv("APP_PUBLIC_URL", "https://tst.wordloop.com.br")
	if _, err := buildMail(); err == nil {
		t.Fatal("o driver de log nao pode ligar com uma URL publica https")
	}
	t.Setenv("APP_PUBLIC_URL", "http://localhost:5173")
	if sender, err := buildMail(); err != nil || sender == nil {
		t.Fatalf("driver de log: %v %v", sender, err)
	} else if _, ok := sender.(mail.Log); !ok {
		t.Fatalf("esperava o driver de log, veio %T", sender)
	}
	t.Setenv("MAIL_DRIVER", "")
	t.Setenv("APP_PUBLIC_URL", "")
	t.Setenv("BREVO_API_KEY", "chave-de-teste")
	if _, err := buildMail(); err == nil {
		t.Fatal("com a chave e sem o remetente a API nao sobe")
	}
	t.Setenv("MAIL_FROM", "nao-responda@wordloop.com.br")
	if sender, err := buildMail(); err != nil {
		t.Fatal(err)
	} else if _, ok := sender.(*mail.Brevo); !ok {
		t.Fatalf("esperava o Brevo, veio %T", sender)
	}
}

func TestBuildRegistrationNeedsBothTheMailServiceAndThePublicURL(t *testing.T) {
	clearMailEnv(t)
	h := httpapi.Handler{}
	if err := buildRegistration(&h, nil); err != nil || h.Registration != nil {
		t.Fatalf("sem e-mail: %v %v", err, h.Registration)
	}
	t.Setenv("MAIL_DRIVER", "log")
	if err := buildRegistration(&h, nil); err != nil || h.Registration != nil {
		t.Fatalf("sem a URL publica os links nao teriam para onde apontar: %v %v", err, h.Registration)
	}
	t.Setenv("APP_PUBLIC_URL", "http://localhost:5173")
	if err := buildRegistration(&h, nil); err != nil || h.Registration == nil {
		t.Fatalf("com tudo: %v %v", err, h.Registration)
	}
	t.Setenv("MAIL_DRIVER", "")
	t.Setenv("BREVO_API_KEY", "k")
	if err := buildRegistration(&httpapi.Handler{}, nil); err == nil {
		t.Fatal("chave sem remetente e erro de configuracao")
	}
}
