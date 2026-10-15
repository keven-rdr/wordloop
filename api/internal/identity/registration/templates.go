package registration

import (
	"bytes"
	"embed"
	"fmt"
	"strings"
	"text/template"
	"time"
)

//go:embed templates/*/*.txt
var templateFiles embed.FS

var locales = []string{"pt-BR", "en-US"}

func normalizeLocale(locale string) string {
	for _, l := range locales {
		if strings.EqualFold(locale, l) {
			return l
		}
	}
	return DefaultLocale
}

type view struct {
	Link      string
	LoginLink string
	Validity  string
}

func validity(locale string, ttl time.Duration) string {
	unit, amount := "minutos", int(ttl.Minutes())
	if locale == "en-US" {
		unit = "minutes"
	}
	if ttl >= time.Hour {
		unit, amount = "horas", int(ttl.Hours())
		if locale == "en-US" {
			unit = "hours"
		}
	}
	return fmt.Sprintf("%d %s", amount, unit)
}

func render(locale, name string, data view) (subject, body string, err error) {
	raw, err := templateFiles.ReadFile(fmt.Sprintf("templates/%s/%s.txt", normalizeLocale(locale), name))
	if err != nil {
		return "", "", fmt.Errorf("modelo %s/%s: %w", locale, name, err)
	}
	tpl, err := template.New(name).Parse(string(raw))
	if err != nil {
		return "", "", fmt.Errorf("modelo %s/%s: %w", locale, name, err)
	}
	var out bytes.Buffer
	if err := tpl.Execute(&out, data); err != nil {
		return "", "", fmt.Errorf("modelo %s/%s: %w", locale, name, err)
	}
	first, rest, _ := strings.Cut(strings.ReplaceAll(out.String(), "\r\n", "\n"), "\n")
	subject = strings.TrimSpace(strings.TrimPrefix(first, "Subject:"))
	body = strings.TrimSpace(rest)
	if subject == "" || body == "" {
		return "", "", fmt.Errorf("modelo %s/%s sem assunto ou corpo", locale, name)
	}
	return subject, body, nil
}
