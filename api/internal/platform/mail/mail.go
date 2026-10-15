package mail

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"
)

const (
	brevoURL       = "https://api.brevo.com"
	sendTimeout    = 10 * time.Second
	maxErrorBytes  = 300
	maxResponseCap = 1 << 16
)

var ErrNotConfigured = errors.New("e-mail nao configurado")

type Message struct {
	To      string
	Subject string
	Text    string
}

type Sender interface {
	Send(ctx context.Context, m Message) error
}

type Brevo struct {
	APIKey    string
	FromEmail string
	FromName  string
	BaseURL   string
	Client    *http.Client
}

func NewBrevo(apiKey, fromEmail, fromName string) (*Brevo, error) {
	if apiKey == "" || fromEmail == "" {
		return nil, ErrNotConfigured
	}
	return &Brevo{APIKey: apiKey, FromEmail: fromEmail, FromName: fromName, BaseURL: brevoURL, Client: &http.Client{Timeout: sendTimeout}}, nil
}

type brevoAddress struct {
	Email string `json:"email"`
	Name  string `json:"name,omitempty"`
}

type brevoRequest struct {
	Sender      brevoAddress   `json:"sender"`
	To          []brevoAddress `json:"to"`
	Subject     string         `json:"subject"`
	TextContent string         `json:"textContent"`
}

func (b *Brevo) Send(ctx context.Context, m Message) error {
	body, err := json.Marshal(brevoRequest{
		Sender:      brevoAddress{Email: b.FromEmail, Name: b.FromName},
		To:          []brevoAddress{{Email: m.To}},
		Subject:     m.Subject,
		TextContent: m.Text,
	})
	if err != nil {
		return fmt.Errorf("montar o e-mail: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, b.BaseURL+"/v3/smtp/email", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("montar o pedido: %w", err)
	}
	req.Header.Set("api-key", b.APIKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	res, err := b.Client.Do(req)
	if err != nil {
		return fmt.Errorf("enviar pelo Brevo: %w", err)
	}
	defer func() { _ = res.Body.Close() }()
	if res.StatusCode/100 == 2 {
		return nil
	}
	detail, _ := io.ReadAll(io.LimitReader(res.Body, maxResponseCap))
	if len(detail) > maxErrorBytes {
		detail = detail[:maxErrorBytes]
	}
	return fmt.Errorf("Brevo respondeu %d: %s", res.StatusCode, detail)
}

type Log struct{}

func (Log) Send(_ context.Context, m Message) error {
	slog.Info("e-mail (driver de log, so para desenvolvimento)", "para", m.To, "assunto", m.Subject, "texto", m.Text)
	return nil
}
