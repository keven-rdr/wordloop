package accounts

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/mail"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/keven-rdr/wordloop/api/internal/identity/domain"
	"github.com/keven-rdr/wordloop/api/internal/platform/token"
)

const InviteTTL = 15 * time.Minute

type CreateAdmin struct {
	Store     Store
	PublicURL string
	Rand      io.Reader
	Now       func() time.Time
}

type Invite struct {
	URL       string
	Email     string
	ExpiresAt time.Time
	Reissued  bool
}

func (u CreateAdmin) Run(ctx context.Context, email, name string) (Invite, error) {
	base, err := publicBase(u.PublicURL)
	if err != nil {
		return Invite{}, err
	}
	addr, err := NormalizeEmail(email)
	if err != nil {
		return Invite{}, err
	}
	now := u.now()
	id, reissued, err := u.account(ctx, addr, name, now)
	if err != nil {
		return Invite{}, err
	}
	tok, hash, err := token.New(u.Rand)
	if err != nil {
		return Invite{}, err
	}
	expires := now.Add(InviteTTL)
	if err := u.Store.ReplaceInvite(ctx, id, hash, now, expires); err != nil {
		return Invite{}, fmt.Errorf("gravar o convite: %w", err)
	}
	return Invite{URL: base + "/invite#" + tok, Email: addr, ExpiresAt: expires, Reissued: reissued}, nil
}

func (u CreateAdmin) account(ctx context.Context, email, name string, now time.Time) (uuid.UUID, bool, error) {
	acc, err := u.Store.FindByEmail(ctx, email)
	switch {
	case errors.Is(err, ErrNotFound):
		id, err := u.Store.CreateAdmin(ctx, email, displayName(name, email), now)
		if err != nil {
			return uuid.Nil, false, fmt.Errorf("criar o admin: %w", err)
		}
		return id, false, nil
	case err != nil:
		return uuid.Nil, false, fmt.Errorf("buscar a conta: %w", err)
	case acc.Role == domain.RoleAdmin && !acc.HasCredentials:
		return acc.ID, true, nil
	}
	return uuid.Nil, false, ErrAccountExists
}

func (u CreateAdmin) now() time.Time {
	if u.Now != nil {
		return u.Now()
	}
	return time.Now()
}

func NormalizeEmail(s string) (string, error) {
	s = strings.ToLower(strings.TrimSpace(s))
	addr, err := mail.ParseAddress(s)
	if err != nil || addr.Address != s || len(s) > 254 {
		return "", ErrInvalidEmail
	}
	_, host, _ := strings.Cut(s, "@")
	if !strings.Contains(host, ".") {
		return "", ErrInvalidEmail
	}
	return s, nil
}

func displayName(name, email string) string {
	if name = strings.TrimSpace(name); name != "" {
		return name
	}
	local, _, _ := strings.Cut(email, "@")
	return local
}

func publicBase(raw string) (string, error) {
	u, err := url.Parse(strings.TrimRight(strings.TrimSpace(raw), "/"))
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" || u.Path != "" || u.Fragment != "" || u.RawQuery != "" {
		return "", ErrInvalidURL
	}
	return u.String(), nil
}
