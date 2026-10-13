//go:build integration

package identitytest

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// Valores do realm de dev (deploy/dev/realms/wordloop-dev.json), os mesmos do Compose.
const (
	RealmName     = "wordloop-dev"
	ClientID      = "wordloop-api"
	ClientSecret  = "wordloop-api-dev-secret"
	UserName      = "aluno@wordloop.local"
	UserPassword  = "aluno"
	UserFullName  = "Aluno Teste"
	adminUser     = "admin"
	adminPassword = "admin"
)

// Keycloak e um Keycloak REAL com o realm de dev importado.
type Keycloak struct {
	BaseURL string // ex.: http://localhost:55123
	t       *testing.T
}

// Issuer e o emissor dos tokens do realm (o Keycloak o deriva do Host do pedido).
func (k Keycloak) Issuer() string { return k.BaseURL + "/realms/" + RealmName }

// StartKeycloak sobe o Keycloak em contêiner e espera o realm de dev responder. Exige Docker ligado.
func StartKeycloak(t *testing.T) Keycloak {
	t.Helper()
	_, file, _, _ := runtime.Caller(0)
	realm := filepath.Join(filepath.Dir(file), "..", "..", "..", "..", "deploy", "dev", "realms", "wordloop-dev.json")
	ctx := context.Background()
	c, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		Started: true,
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "quay.io/keycloak/keycloak:26.7.1",
			Cmd:          []string{"start-dev", "--import-realm"},
			ExposedPorts: []string{"8080/tcp"},
			Env: map[string]string{
				"KC_BOOTSTRAP_ADMIN_USERNAME": adminUser, "KC_BOOTSTRAP_ADMIN_PASSWORD": adminPassword,
				"WORDLOOP_API_CLIENT_SECRET": ClientSecret, "WORDLOOP_ADMIN_CLIENT_SECRET": "admin-secret-teste",
			},
			Files: []testcontainers.ContainerFile{{HostFilePath: realm, ContainerFilePath: "/opt/keycloak/data/import/wordloop-dev.json", FileMode: 0o644}},
			WaitingFor: wait.ForHTTP("/realms/" + RealmName + "/.well-known/openid-configuration").
				WithPort("8080/tcp").WithStartupTimeout(4 * time.Minute),
		},
	})
	if err != nil {
		t.Fatalf("subir o Keycloak (Docker ligado?): %v", err)
	}
	t.Cleanup(func() { _ = testcontainers.TerminateContainer(c) })
	base, err := c.PortEndpoint(ctx, "8080/tcp", "http")
	if err != nil {
		t.Fatal(err)
	}
	return Keycloak{BaseURL: base, t: t}
}

// AllowRedirect registra uma URL de callback no cliente wordloop-api (a porta do servidor de teste e aleatoria).
func (k Keycloak) AllowRedirect(redirectURI string) {
	k.t.Helper()
	token := k.adminToken()
	var found []map[string]any
	k.adminJSON(token, http.MethodGet, "/admin/realms/"+RealmName+"/clients?clientId="+ClientID, nil, &found)
	if len(found) != 1 {
		k.t.Fatalf("cliente %s nao encontrado", ClientID)
	}
	id, ok := found[0]["id"].(string)
	if !ok || id == "" {
		k.t.Fatalf("cliente %s sem id: %v", ClientID, found[0])
	}
	var client map[string]any
	k.adminJSON(token, http.MethodGet, "/admin/realms/"+RealmName+"/clients/"+id, nil, &client)
	uris, _ := client["redirectUris"].([]any)
	client["redirectUris"] = append(uris, redirectURI)
	k.adminJSON(token, http.MethodPut, "/admin/realms/"+RealmName+"/clients/"+id, client, nil)
}

var httpClient = &http.Client{Timeout: 30 * time.Second}

// adminToken pede o token do admin; repete porque o realm master pode demorar um instante depois do realm de dev.
func (k Keycloak) adminToken() string {
	form := url.Values{"client_id": {"admin-cli"}, "grant_type": {"password"}, "username": {adminUser}, "password": {adminPassword}}
	var lastErr error
	for attempt := 0; attempt < 10; attempt++ {
		if attempt > 0 {
			time.Sleep(2 * time.Second)
		}
		res, err := httpClient.PostForm(k.BaseURL+"/realms/master/protocol/openid-connect/token", form)
		if err != nil {
			lastErr = err
			continue
		}
		var body struct {
			AccessToken string `json:"access_token"`
		}
		err = json.NewDecoder(res.Body).Decode(&body)
		_ = res.Body.Close()
		if err == nil && body.AccessToken != "" {
			return body.AccessToken
		}
		lastErr = fmt.Errorf("status %d, %v", res.StatusCode, err)
	}
	k.t.Fatalf("token de admin do Keycloak: %v", lastErr)
	return ""
}

func (k Keycloak) adminJSON(token, method, path string, in, out any) {
	k.t.Helper()
	var buf io.Reader
	if in != nil {
		raw, _ := json.Marshal(in)
		buf = bytes.NewReader(raw)
	}
	req, _ := http.NewRequest(method, k.BaseURL+path, buf)
	req.Header.Set("Authorization", "Bearer "+token)
	if in != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	res, err := httpClient.Do(req)
	if err != nil {
		k.t.Fatal(err)
	}
	defer func() { _ = res.Body.Close() }()
	if res.StatusCode >= 300 {
		msg, _ := io.ReadAll(res.Body)
		k.t.Fatalf("%s %s: %d %s", method, path, res.StatusCode, strings.TrimSpace(string(msg)))
	}
	if out != nil {
		if err := json.NewDecoder(res.Body).Decode(out); err != nil {
			k.t.Fatal(fmt.Sprintf("%s %s: %v", method, path, err))
		}
	}
}
