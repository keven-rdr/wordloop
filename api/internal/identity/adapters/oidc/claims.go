package oidc

import (
	"fmt"
	"slices"
	"strconv"
	"time"

	gooidc "github.com/coreos/go-oidc/v3/oidc"

	"github.com/keven-rdr/wordloop/api/internal/identity/domain"
)

func identityFrom(idt *gooidc.IDToken) (domain.Identity, error) {
	var claims struct {
		Name              string `json:"name"`
		PreferredUsername string `json:"preferred_username"`
		AMR               any    `json:"amr"`
		AuthTime          any    `json:"auth_time"`
	}
	if err := idt.Claims(&claims); err != nil {
		return domain.Identity{}, fmt.Errorf("ler as claims: %w", err)
	}
	name := claims.Name
	if name == "" {
		name = claims.PreferredUsername
	}
	return domain.Identity{
		Issuer: idt.Issuer, Subject: idt.Subject, DisplayName: name,
		AuthTime: unixTime(claims.AuthTime), MFA: hasMFA(methods(claims.AMR)),
	}, nil
}

func hasMFA(amr []string) bool {
	if slices.Contains(amr, "mfa") {
		return true
	}
	return slices.Contains(amr, "pwd") && (slices.Contains(amr, "otp") || slices.Contains(amr, "hwk"))
}

func methods(v any) []string {
	switch m := v.(type) {
	case string:
		return []string{m}
	case []any:
		out := make([]string, 0, len(m))
		for _, e := range m {
			if s, ok := e.(string); ok {
				out = append(out, s)
			}
		}
		return out
	}
	return nil
}

func unixTime(v any) time.Time {
	var secs float64
	switch n := v.(type) {
	case float64:
		secs = n
	case string:
		f, err := strconv.ParseFloat(n, 64)
		if err != nil {
			return time.Time{}
		}
		secs = f
	default:
		return time.Time{}
	}
	if secs <= 0 {
		return time.Time{}
	}
	return time.Unix(int64(secs), 0).UTC()
}
