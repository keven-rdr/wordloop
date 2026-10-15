package github

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"golang.org/x/oauth2"
	githuboauth "golang.org/x/oauth2/github"

	"github.com/keven-rdr/wordloop/api/internal/identity/app"
	"github.com/keven-rdr/wordloop/api/internal/identity/domain"
)

const (
	apiBase     = "https://api.github.com"
	httpTimeout = 10 * time.Second
	maxBody     = 1 << 20
)

type Config struct {
	ClientID, ClientSecret, RedirectURL string
}

type Provider struct {
	oauth   *oauth2.Config
	apiBase string
	client  *http.Client
}

var _ app.Provider = (*Provider)(nil)

func New(cfg Config) *Provider {
	return &Provider{
		oauth: &oauth2.Config{
			ClientID: cfg.ClientID, ClientSecret: cfg.ClientSecret, RedirectURL: cfg.RedirectURL,
			Endpoint: githuboauth.Endpoint, Scopes: []string{"read:user", "user:email"},
		},
		apiBase: apiBase,
		client:  &http.Client{Timeout: httpTimeout},
	}
}

func (p *Provider) AuthURL(_ context.Context, state, _, verifier string) (string, error) {
	return p.oauth.AuthCodeURL(state, oauth2.S256ChallengeOption(verifier)), nil
}

func (p *Provider) Exchange(ctx context.Context, code, verifier, _ string) (app.Tokens, error) {
	ctx = context.WithValue(ctx, oauth2.HTTPClient, p.client)
	tok, err := p.oauth.Exchange(ctx, code, oauth2.VerifierOption(verifier))
	if err != nil {
		return app.Tokens{}, fmt.Errorf("trocar o code: %w", err)
	}
	client := p.oauth.Client(ctx, tok)
	var user struct {
		ID    int64  `json:"id"`
		Login string `json:"login"`
		Name  string `json:"name"`
	}
	if err := p.get(ctx, client, "/user", &user); err != nil {
		return app.Tokens{}, err
	}
	if user.ID == 0 {
		return app.Tokens{}, errors.New("o GitHub nao informou o id do usuario")
	}
	var emails []struct {
		Email    string `json:"email"`
		Primary  bool   `json:"primary"`
		Verified bool   `json:"verified"`
	}
	if err := p.get(ctx, client, "/user/emails", &emails); err != nil {
		return app.Tokens{}, err
	}
	name := strings.TrimSpace(user.Name)
	if name == "" {
		name = user.Login
	}
	identity := domain.Identity{Issuer: "github", Subject: strconv.FormatInt(user.ID, 10), DisplayName: name, AuthTime: time.Now().UTC()}
	for _, e := range emails {
		if e.Verified && (identity.Email == "" || e.Primary) {
			identity.Email, identity.EmailVerified = e.Email, true
		}
	}
	return app.Tokens{Identity: identity}, nil
}

func (p *Provider) get(ctx context.Context, client *http.Client, path string, into any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, p.apiBase+path, nil)
	if err != nil {
		return fmt.Errorf("montar o pedido ao GitHub: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "wordloop")
	res, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("falar com o GitHub: %w", err)
	}
	defer func() { _ = res.Body.Close() }()
	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("o GitHub respondeu %d em %s", res.StatusCode, path)
	}
	if err := json.NewDecoder(io.LimitReader(res.Body, maxBody)).Decode(into); err != nil {
		return fmt.Errorf("ler a resposta do GitHub: %w", err)
	}
	return nil
}
