package registration_test

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/keven-rdr/wordloop/api/internal/identity/domain"
	"github.com/keven-rdr/wordloop/api/internal/identity/identitytest"
	"github.com/keven-rdr/wordloop/api/internal/identity/registration"
	"github.com/keven-rdr/wordloop/api/internal/platform/token"
)

var (
	bg  = context.Background()
	t0  = time.Date(2026, 10, 16, 12, 0, 0, 0, time.UTC)
	now = t0
)

type env struct {
	store    *identitytest.RegistrationStore
	mailbox  *identitytest.MailBox
	register registration.Register
	forgot   registration.Forgot
}

func newEnv() *env {
	e := &env{store: &identitytest.RegistrationStore{}, mailbox: &identitytest.MailBox{}}
	issuer := registration.Issuer{
		Store: e.store, Mail: e.mailbox, PublicURL: "https://tst.wordloop.com.br",
		Rand: bytes.NewReader(bytes.Repeat([]byte{7}, 4096)), Now: func() time.Time { return now },
	}
	e.register, e.forgot = registration.Register{Issuer: issuer}, registration.Forgot{Issuer: issuer}
	now = t0
	return e
}

func TestNewEmailGetsAPendingAccountAndAConfirmationLink(t *testing.T) {
	e := newEnv()
	if err := e.register.Run(bg, registration.Input{Email: "  Ana@Exemplo.com ", Locale: "pt-BR"}); err != nil {
		t.Fatal(err)
	}
	if e.store.Accounts() != 1 || e.mailbox.Count() != 1 {
		t.Fatalf("contas=%d e-mails=%d", e.store.Accounts(), e.mailbox.Count())
	}
	m := e.mailbox.Last()
	if m.To != "ana@exemplo.com" || !strings.Contains(m.Subject, "Confirme") || !strings.Contains(m.Text, "https://tst.wordloop.com.br/verify#") {
		t.Fatalf("e-mail: %+v", m)
	}
	purpose, hash, expires := e.store.LastToken()
	if purpose != domain.TokenVerifyEmail || !bytes.Equal(hash, token.Hash(e.mailbox.LastSecret())) || !expires.Equal(t0.Add(registration.VerifyTTL)) {
		t.Fatalf("token: %s %x %v", purpose, hash, expires)
	}
	if strings.Contains(m.Text, string(hash)) {
		t.Fatal("o e-mail leva o segredo, nunca o hash")
	}
}

func TestEnglishLocaleAndUnknownLocalesFallBackToPortuguese(t *testing.T) {
	e := newEnv()
	_ = e.register.Run(bg, registration.Input{Email: "john@example.com", Locale: "en-US"})
	if !strings.Contains(e.mailbox.Last().Subject, "Confirm your e-mail") || !strings.Contains(e.mailbox.Last().Text, "24 hours") {
		t.Fatalf("en-US: %+v", e.mailbox.Last())
	}
	_ = e.register.Run(bg, registration.Input{Email: "maria@example.com", Locale: "fr-FR"})
	if !strings.Contains(e.mailbox.Last().Subject, "Confirme o seu e-mail") || !strings.Contains(e.mailbox.Last().Text, "24 horas") {
		t.Fatalf("fallback: %+v", e.mailbox.Last())
	}
}

func TestMalformedEmailsAreRefusedBeforeAnythingIsCreated(t *testing.T) {
	e := newEnv()
	for _, bad := range []string{"", "ana", "ana@", "@exemplo.com", "ana@exemplo", strings.Repeat("a", 260) + "@exemplo.com"} {
		if err := e.register.Run(bg, registration.Input{Email: bad}); !errors.Is(err, registration.ErrInvalidEmail) {
			t.Errorf("%q: esperava ErrInvalidEmail, veio %v", bad, err)
		}
		if err := e.forgot.Run(bg, registration.Input{Email: bad}); !errors.Is(err, registration.ErrInvalidEmail) {
			t.Errorf("forgot %q: esperava ErrInvalidEmail, veio %v", bad, err)
		}
	}
	if e.store.Accounts() != 0 || e.mailbox.Count() != 0 {
		t.Fatal("nada pode ser criado nem enviado")
	}
}

func TestResendsAreThrottledByMinuteAndByHour(t *testing.T) {
	e := newEnv()
	in := registration.Input{Email: "ana@exemplo.com"}
	_ = e.register.Run(bg, in)
	now = t0.Add(30 * time.Second)
	if err := e.register.Run(bg, in); err != nil || e.mailbox.Count() != 1 {
		t.Fatalf("dentro do minuto nao reenvia, sem erro: %v %d", err, e.mailbox.Count())
	}
	for i := 1; i <= 4; i++ {
		now = t0.Add(time.Duration(i) * 2 * time.Minute)
		_ = e.register.Run(bg, in)
	}
	if e.mailbox.Count() != 5 {
		t.Fatalf("5 e-mails na hora: %d", e.mailbox.Count())
	}
	now = t0.Add(20 * time.Minute)
	_ = e.register.Run(bg, in)
	if e.mailbox.Count() != 5 {
		t.Fatalf("o 6o na mesma hora e descartado: %d", e.mailbox.Count())
	}
	now = t0.Add(2 * time.Hour)
	_ = e.register.Run(bg, in)
	if e.mailbox.Count() != 6 {
		t.Fatalf("passada a hora volta a enviar: %d", e.mailbox.Count())
	}
}

func TestExistingAccountsGetTheSameAnswerButADifferentEmail(t *testing.T) {
	e := newEnv()
	e.store.Seed("pendente@exemplo.com", domain.RoleUser, false)
	e.store.Seed("aluna@exemplo.com", domain.RoleUser, true)
	e.store.Seed("dono@exemplo.com", domain.RoleAdmin, true)
	for _, addr := range []string{"novo@exemplo.com", "pendente@exemplo.com", "aluna@exemplo.com", "dono@exemplo.com"} {
		if err := e.register.Run(bg, registration.Input{Email: addr}); err != nil {
			t.Fatalf("%s: a resposta e sempre sucesso, veio %v", addr, err)
		}
	}
	if e.mailbox.Count() != 3 {
		t.Fatalf("admin nao recebe e-mail: %d", e.mailbox.Count())
	}
	subjects := []string{}
	for _, m := range e.mailbox.Sent {
		subjects = append(subjects, m.Subject)
	}
	if !strings.Contains(subjects[2], "já tem conta") || !strings.Contains(e.mailbox.Sent[2].Text, "/reset#") || !strings.Contains(e.mailbox.Sent[2].Text, "/login") {
		t.Fatalf("conta confirmada recebe o aviso com os caminhos: %v", subjects)
	}
	if purpose, _, _ := e.store.LastToken(); purpose != domain.TokenRecovery {
		t.Fatalf("o aviso leva um link de redefinicao: %s", purpose)
	}
}

func TestAnEmailTakenInTheMeantimeIsHandledAsExisting(t *testing.T) {
	e := newEnv()
	e.store.TakenOnce = true
	if err := e.register.Run(bg, registration.Input{Email: "corrida@exemplo.com"}); err != nil || e.mailbox.Count() != 1 {
		t.Fatalf("corrida: %v %d", err, e.mailbox.Count())
	}
}

func TestAnEmailHeldByAnInvisibleAccountNeverLoopsNorSendsAnything(t *testing.T) {
	e := newEnv()
	e.store.TakenOnce = true
	e.store.HideTaken = true
	if err := e.register.Run(bg, registration.Input{Email: "apagada@exemplo.com"}); err != nil || e.mailbox.Count() != 0 {
		t.Fatalf("e-mail de conta apagada: %v %d", err, e.mailbox.Count())
	}
}

func TestStoreAndMailFailuresAreReported(t *testing.T) {
	boom := errors.New("fora do ar")
	e := newEnv()
	e.store.FailFind = boom
	if err := e.register.Run(bg, registration.Input{Email: "a@exemplo.com"}); !errors.Is(err, boom) {
		t.Fatalf("busca: %v", err)
	}
	if err := e.forgot.Run(bg, registration.Input{Email: "a@exemplo.com"}); !errors.Is(err, boom) {
		t.Fatalf("busca (forgot): %v", err)
	}
	e = newEnv()
	e.store.FailNew = boom
	if err := e.register.Run(bg, registration.Input{Email: "a@exemplo.com"}); !errors.Is(err, boom) {
		t.Fatalf("criar: %v", err)
	}
	e = newEnv()
	e.mailbox.Err = boom
	if err := e.register.Run(bg, registration.Input{Email: "a@exemplo.com"}); !errors.Is(err, boom) {
		t.Fatalf("envio sincrono: %v", err)
	}
}

func TestAsyncSendingNeverBlocksNorFailsTheRequest(t *testing.T) {
	e := newEnv()
	e.mailbox.Done = make(chan struct{})
	e.mailbox.Err = errors.New("Brevo fora do ar")
	e.register.Async = true
	if err := e.register.Run(bg, registration.Input{Email: "ana@exemplo.com"}); err != nil {
		t.Fatalf("assincrono nao devolve erro de envio: %v", err)
	}
	select {
	case <-e.mailbox.Done:
	case <-time.After(5 * time.Second):
		t.Fatal("o envio em segundo plano nao aconteceu")
	}
}

func TestForgotOnlySendsToConfirmedStudentAccounts(t *testing.T) {
	e := newEnv()
	e.store.Seed("aluna@exemplo.com", domain.RoleUser, true)
	e.store.Seed("pendente@exemplo.com", domain.RoleUser, false)
	e.store.Seed("dono@exemplo.com", domain.RoleAdmin, true)
	for _, addr := range []string{"ninguem@exemplo.com", "pendente@exemplo.com", "dono@exemplo.com", "aluna@exemplo.com"} {
		if err := e.forgot.Run(bg, registration.Input{Email: addr, Locale: "en-US"}); err != nil {
			t.Fatalf("%s: %v", addr, err)
		}
	}
	m := e.mailbox.Last()
	if e.mailbox.Count() != 1 || m.To != "aluna@exemplo.com" || !strings.Contains(m.Subject, "Reset your wordloop password") || !strings.Contains(m.Text, "/reset#") {
		t.Fatalf("so a aluna confirmada recebe: %d %+v", e.mailbox.Count(), m)
	}
	if purpose, _, expires := e.store.LastToken(); purpose != domain.TokenRecovery || !expires.Equal(t0.Add(registration.RecoveryTTL)) {
		t.Fatalf("token de recuperacao de 30 minutos: %s %v", purpose, expires)
	}
	if e.store.Accounts() != 3 {
		t.Fatal("recuperar senha nunca cria conta")
	}
}

func TestAnInvalidPublicURLStopsBeforeAnyTokenIsStored(t *testing.T) {
	e := newEnv()
	e.store.Seed("aluna@exemplo.com", domain.RoleUser, true)
	bad := registration.Forgot{Issuer: registration.Issuer{Store: e.store, Mail: e.mailbox, PublicURL: "isto nao e uma url"}}
	if err := bad.Run(bg, registration.Input{Email: "aluna@exemplo.com"}); err == nil {
		t.Fatal("URL publica invalida deve falhar")
	}
	if e.store.OpenTokens() != 0 || e.mailbox.Count() != 0 {
		t.Fatal("nada pode ser gravado nem enviado")
	}
}
