package registration

import (
	"strings"
	"testing"
)

func TestEveryTemplateRendersInBothLanguages(t *testing.T) {
	data := view{Link: "https://x.test/reset#segredo", LoginLink: "https://x.test/login", Validity: "30 minutos"}
	for _, locale := range locales {
		for _, name := range []string{"verify", "exists", "recovery"} {
			subject, body, err := render(locale, name, data)
			if err != nil {
				t.Fatalf("%s/%s: %v", locale, name, err)
			}
			if subject == "" || strings.Contains(subject, "Subject:") || strings.Contains(body, "{{") || !strings.Contains(body, data.Link) || !strings.Contains(body, data.Validity) {
				t.Errorf("%s/%s: assunto=%q corpo=%q", locale, name, subject, body)
			}
		}
	}
}

func TestUnknownTemplatesAndLocalesAreHandled(t *testing.T) {
	if _, _, err := render("pt-BR", "inexistente", view{}); err == nil {
		t.Fatal("modelo inexistente deve falhar")
	}
	if got := normalizeLocale("EN-us"); got != "en-US" {
		t.Fatalf("normalizar: %q", got)
	}
	if got := normalizeLocale(""); got != DefaultLocale {
		t.Fatalf("vazio vira o padrao: %q", got)
	}
}

func TestValidityIsWrittenInTheReadersLanguage(t *testing.T) {
	cases := map[string]string{
		validity("pt-BR", VerifyTTL):   "24 horas",
		validity("en-US", VerifyTTL):   "24 hours",
		validity("pt-BR", RecoveryTTL): "30 minutos",
		validity("en-US", RecoveryTTL): "30 minutes",
	}
	for got, want := range cases {
		if got != want {
			t.Errorf("esperava %q, veio %q", want, got)
		}
	}
}
