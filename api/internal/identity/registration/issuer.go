package registration

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"time"

	"github.com/google/uuid"

	"github.com/keven-rdr/wordloop/api/internal/identity/accounts"
	"github.com/keven-rdr/wordloop/api/internal/identity/domain"
	"github.com/keven-rdr/wordloop/api/internal/platform/mail"
	"github.com/keven-rdr/wordloop/api/internal/platform/token"
)

type kind struct {
	purpose  domain.TokenPurpose
	ttl      time.Duration
	path     string
	template string
}

var (
	kindVerify   = kind{domain.TokenVerifyEmail, VerifyTTL, "/verify", "verify"}
	kindExists   = kind{domain.TokenRecovery, RecoveryTTL, "/reset", "exists"}
	kindRecovery = kind{domain.TokenRecovery, RecoveryTTL, "/reset", "recovery"}
)

type Issuer struct {
	Store     Store
	Mail      mail.Sender
	PublicURL string
	Rand      io.Reader
	Now       func() time.Time
	Async     bool
}

func (i Issuer) now() time.Time {
	if i.Now != nil {
		return i.Now()
	}
	return time.Now()
}

func (i Issuer) throttled(ctx context.Context, userID uuid.UUID, now time.Time) (bool, error) {
	recent, err := i.Store.RecentTokens(ctx, userID, now.Add(-cooldown))
	if err != nil || recent > 0 {
		return true, err
	}
	hourly, err := i.Store.RecentTokens(ctx, userID, now.Add(-time.Hour))
	return hourly >= hourlyLimit, err
}

func (i Issuer) issue(ctx context.Context, userID uuid.UUID, email string, k kind, locale string) error {
	base, err := accounts.PublicBase(i.PublicURL)
	if err != nil {
		return err
	}
	now := i.now()
	if skip, err := i.throttled(ctx, userID, now); err != nil || skip {
		return err
	}
	secret, hash, err := token.New(i.Rand)
	if err != nil {
		return err
	}
	if err := i.Store.ReplaceToken(ctx, userID, k.purpose, hash, now, now.Add(k.ttl)); err != nil {
		return fmt.Errorf("gravar o token: %w", err)
	}
	locale = normalizeLocale(locale)
	subject, body, err := render(locale, k.template, view{Link: base + k.path + "#" + secret, LoginLink: base + "/login", Validity: validity(locale, k.ttl)})
	if err != nil {
		return err
	}
	message := mail.Message{To: email, Subject: subject, Text: body}
	if !i.Async {
		return i.Mail.Send(ctx, message)
	}
	go i.send(context.WithoutCancel(ctx), userID, message)
	return nil
}

func (i Issuer) send(ctx context.Context, userID uuid.UUID, m mail.Message) {
	ctx, cancel := context.WithTimeout(ctx, sendTimeout)
	defer cancel()
	if err := i.Mail.Send(ctx, m); err != nil {
		slog.Error("nao foi possivel enviar o e-mail", "user_id", userID, "err", err)
	}
}
