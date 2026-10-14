package httpapi_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/keven-rdr/wordloop/api/internal/api"
	httpapi "github.com/keven-rdr/wordloop/api/internal/identity/adapters/http"
	"github.com/keven-rdr/wordloop/api/internal/identity/adapters/webauthn"
	"github.com/keven-rdr/wordloop/api/internal/identity/app"
	"github.com/keven-rdr/wordloop/api/internal/identity/domain"
	"github.com/keven-rdr/wordloop/api/internal/identity/identitytest"
	"github.com/keven-rdr/wordloop/api/internal/identity/passkeys"
	"github.com/keven-rdr/wordloop/api/internal/platform/ratelimit"
	"github.com/keven-rdr/wordloop/api/internal/platform/seal"
)

const (
	pkOrigin = "https://tst.wordloop.com.br"
	pkRPID   = "tst.wordloop.com.br"
	pkInvite = "convite-http-0123456789abcdefghijklmnopqrstuvw"
)

type pkFixture struct {
	h     http.Handler
	store *identitytest.PasskeyStore
	users *identitytest.Users
	owner passkeys.Owner
}

func newPasskeyFixture(t *testing.T, role domain.Role, enabled bool, limiter *ratelimit.Limiter) pkFixture {
	t.Helper()
	cer, err := webauthn.New(webauthn.Config{RPID: pkRPID, RPName: "wordloop", Origins: []string{pkOrigin}})
	if err != nil {
		t.Fatal(err)
	}
	sealer, _ := seal.New(bytes.Repeat([]byte{9}, seal.KeySize))
	users, sessions := &identitytest.Users{}, &identitytest.Sessions{}
	user, _ := users.Upsert(context.Background(), domain.Identity{Issuer: "i", Subject: "s", DisplayName: "Voce"})
	users.SetRole(user.ID, role)
	store, challenges := identitytest.NewPasskeyStore(), identitytest.NewMemChallenges()
	owner := passkeys.Owner{ID: user.ID, Email: "voce@exemplo.com", DisplayName: "Voce"}
	store.AddOwner(owner)
	store.AddInvite(pkInvite, owner.ID, time.Now().Add(15*time.Minute))
	opener := app.SessionOpener{Sessions: sessions, Sealer: sealer}
	auth := app.Authenticate{Sessions: sessions, Users: users}
	h := httpapi.Handler{
		Auth: auth, Out: app.Logout{Sessions: sessions}, Secure: true,
		List:   app.ListSessions{Auth: auth, Sessions: sessions},
		Revoke: app.RevokeSession{Auth: auth, Sessions: sessions}, RevokeAll: app.RevokeAllSessions{Auth: auth, Sessions: sessions},
	}
	if enabled {
		h.Passkeys = &httpapi.Passkeys{
			BeginRegistration:  passkeys.BeginRegistration{Ceremony: cer, Store: store, Challenges: challenges},
			FinishRegistration: passkeys.FinishRegistration{Ceremony: cer, Store: store, Challenges: challenges, Users: users, Opener: opener},
			BeginLogin:         passkeys.BeginLogin{Ceremony: cer, Challenges: challenges},
			FinishLogin:        passkeys.FinishLogin{Ceremony: cer, Store: store, Challenges: challenges, Users: users, Opener: opener},
			List:               passkeys.List{Store: store},
			Delete:             passkeys.Delete{Store: store},
		}
	}
	return pkFixture{h: api.Handler(api.Server{Identity: h, AuthLimiter: limiter}), store: store, users: users, owner: owner}
}

func post(h http.Handler, path string, body any, headers map[string]string, cookies ...*http.Cookie) *http.Response {
	var rd *bytes.Reader
	switch b := body.(type) {
	case nil:
		rd = bytes.NewReader(nil)
	case []byte:
		rd = bytes.NewReader(b)
	default:
		raw, _ := json.Marshal(b)
		rd = bytes.NewReader(raw)
	}
	req := httptest.NewRequest(http.MethodPost, path, rd)
	req.Header.Set("Content-Type", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	for _, c := range cookies {
		req.AddCookie(c)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec.Result()
}

func options(t *testing.T, res *http.Response) []byte {
	t.Helper()
	var out struct {
		Options json.RawMessage `json:"options"`
	}
	if err := json.NewDecoder(res.Body).Decode(&out); err != nil || res.StatusCode != http.StatusOK || len(out.Options) == 0 {
		t.Fatalf("opcoes: status %d (%v)", res.StatusCode, err)
	}
	return out.Options
}

func (f pkFixture) registerWithInvite(t *testing.T, a *identitytest.SoftAuthenticator) (session *http.Cookie) {
	t.Helper()
	begin := post(f.h, "/api/v1/auth/passkeys/register/begin", map[string]string{"invite": pkInvite}, csrf)
	opts := options(t, begin)
	challenge := cookie(t, begin, httpapi.ChallengeCookieSecure)
	if !challenge.HttpOnly || !challenge.Secure || challenge.Path != "/api/v1/auth/passkeys" || challenge.MaxAge <= 0 || challenge.MaxAge > 300 {
		t.Fatalf("cookie do desafio sem os atributos de seguranca: %+v", challenge)
	}
	cred, _ := a.Create(opts)
	fin := post(f.h, "/api/v1/auth/passkeys/register/finish", map[string]any{"credential": json.RawMessage(cred), "name": "Meu celular"}, csrf, challenge)
	if fin.StatusCode != http.StatusCreated {
		t.Fatalf("finish: %d", fin.StatusCode)
	}
	if cleared := cookie(t, fin, httpapi.ChallengeCookieSecure); cleared.MaxAge >= 0 {
		t.Fatal("o cookie do desafio deve ser apagado")
	}
	return cookie(t, fin, httpapi.SessionCookieSecure)
}

func TestPasskeyEndToEndInviteThenLoginOverHTTP(t *testing.T) {
	f := newPasskeyFixture(t, domain.RoleAdmin, true, nil)
	a := identitytest.NewSoftAuthenticator(pkOrigin, pkRPID)
	session := f.registerWithInvite(t, a)
	if !session.HttpOnly || !session.Secure || session.Name != "__Host-wl_session" {
		t.Fatalf("a sessao do convite deve ser o cookie __Host-: %+v", session)
	}
	if code, me := getMe(t, f.h, session); code != http.StatusOK || me.Role != "admin" {
		t.Fatalf("/me depois do convite: %d %+v", code, me)
	}

	begin := post(f.h, "/api/v1/auth/passkeys/login/begin", nil, csrf)
	opts := options(t, begin)
	resp, _ := a.Get(opts, f.owner.ID[:])
	fin := post(f.h, "/api/v1/auth/passkeys/login/finish", map[string]any{"credential": json.RawMessage(resp)}, csrf, cookie(t, begin, httpapi.ChallengeCookieSecure))
	if fin.StatusCode != http.StatusNoContent {
		t.Fatalf("login finish: %d", fin.StatusCode)
	}
	if code, _ := getMe(t, f.h, cookie(t, fin, httpapi.SessionCookieSecure)); code != http.StatusOK {
		t.Fatalf("/me depois do login por passkey: %d", code)
	}
}

func TestPasskeyPostsRequireTheAntiCSRFHeader(t *testing.T) {
	f := newPasskeyFixture(t, domain.RoleUser, true, nil)
	for _, path := range []string{"register/begin", "register/finish", "login/begin", "login/finish"} {
		if res := post(f.h, "/api/v1/auth/passkeys/"+path, map[string]any{"credential": map[string]any{}}, nil); res.StatusCode != http.StatusBadRequest {
			t.Errorf("%s sem o cabecalho: esperava 400, veio %d", path, res.StatusCode)
		}
	}
}

func TestRegistrationNeedsAValidInviteOrSession(t *testing.T) {
	f := newPasskeyFixture(t, domain.RoleUser, true, nil)
	for name, body := range map[string]any{"sem corpo": nil, "convite inventado": map[string]string{"invite": "inventado"}, "convite vazio": map[string]string{"invite": ""}} {
		if res := post(f.h, "/api/v1/auth/passkeys/register/begin", body, csrf); res.StatusCode != http.StatusUnauthorized {
			t.Errorf("%s: esperava 401, veio %d", name, res.StatusCode)
		}
	}
	if res := post(f.h, "/api/v1/auth/passkeys/register/begin", map[string]string{"invite": pkInvite}, csrf, &http.Cookie{Name: httpapi.SessionCookieSecure, Value: "falso"}); res.StatusCode != http.StatusOK {
		t.Fatalf("cookie de sessao falso nao pode atrapalhar um convite valido: %d", res.StatusCode)
	}
}

func TestFinishRejectsMissingReusedAndForeignChallenges(t *testing.T) {
	f := newPasskeyFixture(t, domain.RoleUser, true, nil)
	a := identitytest.NewSoftAuthenticator(pkOrigin, pkRPID)
	begin := post(f.h, "/api/v1/auth/passkeys/register/begin", map[string]string{"invite": pkInvite}, csrf)
	cred, _ := a.Create(options(t, begin))
	body := map[string]any{"credential": json.RawMessage(cred)}
	if res := post(f.h, "/api/v1/auth/passkeys/register/finish", body, csrf); res.StatusCode != http.StatusBadRequest {
		t.Fatalf("sem cookie do desafio: %d", res.StatusCode)
	}
	if res := post(f.h, "/api/v1/auth/passkeys/register/finish", body, csrf, &http.Cookie{Name: httpapi.ChallengeCookieSecure, Value: "inventado"}); res.StatusCode != http.StatusBadRequest {
		t.Fatalf("desafio inventado: %d", res.StatusCode)
	}
	challenge := cookie(t, begin, httpapi.ChallengeCookieSecure)
	if res := post(f.h, "/api/v1/auth/passkeys/register/finish", body, csrf, challenge); res.StatusCode != http.StatusCreated {
		t.Fatalf("o primeiro uso deve funcionar: %d", res.StatusCode)
	}
	if res := post(f.h, "/api/v1/auth/passkeys/register/finish", body, csrf, challenge); res.StatusCode != http.StatusBadRequest {
		t.Fatalf("reaproveitar o desafio: %d", res.StatusCode)
	}
	if res := post(f.h, "/api/v1/auth/passkeys/register/finish", map[string]any{}, csrf, challenge); res.StatusCode != http.StatusBadRequest {
		t.Fatalf("sem credencial: %d", res.StatusCode)
	}
}

func TestRegistrationFromAnotherOriginIsRefused(t *testing.T) {
	f := newPasskeyFixture(t, domain.RoleUser, true, nil)
	phisher := identitytest.NewSoftAuthenticator("https://tst-wordloop.evil.example", "tst-wordloop.evil.example")
	begin := post(f.h, "/api/v1/auth/passkeys/register/begin", map[string]string{"invite": pkInvite}, csrf)
	cred, _ := phisher.Create(options(t, begin))
	res := post(f.h, "/api/v1/auth/passkeys/register/finish", map[string]any{"credential": json.RawMessage(cred)}, csrf, cookie(t, begin, httpapi.ChallengeCookieSecure))
	if res.StatusCode != http.StatusBadRequest || len(f.store.Credentials()) != 0 {
		t.Fatalf("origem de outro site deve ser recusada e nada gravado: %d", res.StatusCode)
	}
}

func TestLoginWithAnUnknownPasskeyIs401AndCreatesNoSession(t *testing.T) {
	f := newPasskeyFixture(t, domain.RoleUser, true, nil)
	begin := post(f.h, "/api/v1/auth/passkeys/login/begin", nil, csrf)
	resp, _ := identitytest.NewSoftAuthenticator(pkOrigin, pkRPID).Get(options(t, begin), f.owner.ID[:])
	res := post(f.h, "/api/v1/auth/passkeys/login/finish", map[string]any{"credential": json.RawMessage(resp)}, csrf, cookie(t, begin, httpapi.ChallengeCookieSecure))
	if res.StatusCode != http.StatusUnauthorized {
		t.Fatalf("passkey desconhecida: %d", res.StatusCode)
	}
	for _, c := range res.Cookies() {
		if c.Name == httpapi.SessionCookieSecure && c.MaxAge >= 0 {
			t.Fatal("nenhuma sessao pode ser criada")
		}
	}
}

func TestManagingPasskeysRequiresASessionAndOwnership(t *testing.T) {
	f := newPasskeyFixture(t, domain.RoleAdmin, true, nil)
	session := f.registerWithInvite(t, identitytest.NewSoftAuthenticator(pkOrigin, pkRPID))
	if res := do(f.h, http.MethodGet, "/api/v1/me/passkeys", nil); res.StatusCode != http.StatusUnauthorized {
		t.Fatalf("lista sem sessao: %d", res.StatusCode)
	}
	res := do(f.h, http.MethodGet, "/api/v1/me/passkeys", nil, session)
	var list []struct {
		ID, Name string
	}
	if err := json.NewDecoder(res.Body).Decode(&list); err != nil || res.StatusCode != http.StatusOK || len(list) != 1 || list[0].Name != "Meu celular" {
		t.Fatalf("lista: %d %v %v", res.StatusCode, list, err)
	}
	path := "/api/v1/me/passkeys/"
	if res := do(f.h, http.MethodDelete, path+list[0].ID, nil, session); res.StatusCode != http.StatusBadRequest {
		t.Fatalf("apagar sem o cabecalho anti-CSRF: %d", res.StatusCode)
	}
	if res := do(f.h, http.MethodDelete, path+uuid.NewString(), csrf, session); res.StatusCode != http.StatusNotFound {
		t.Fatalf("passkey de outra pessoa ou inexistente: %d", res.StatusCode)
	}
	if res := do(f.h, http.MethodDelete, path+list[0].ID, csrf, session); res.StatusCode != http.StatusConflict {
		t.Fatalf("o admin nao pode ficar sem passkey: %d", res.StatusCode)
	}
	if res := do(f.h, http.MethodDelete, path+list[0].ID, csrf, &http.Cookie{Name: httpapi.SessionCookieSecure, Value: "falso"}); res.StatusCode != http.StatusUnauthorized {
		t.Fatalf("apagar com cookie falso: %d", res.StatusCode)
	}
}

func TestSecondPasskeyViaSessionAddsOneWithoutANewSession(t *testing.T) {
	f := newPasskeyFixture(t, domain.RoleUser, true, nil)
	session := f.registerWithInvite(t, identitytest.NewSoftAuthenticator(pkOrigin, pkRPID))
	begin := post(f.h, "/api/v1/auth/passkeys/register/begin", nil, csrf, session)
	cred, _ := identitytest.NewSoftAuthenticator(pkOrigin, pkRPID).Create(options(t, begin))
	fin := post(f.h, "/api/v1/auth/passkeys/register/finish", map[string]any{"credential": json.RawMessage(cred)}, csrf, session, cookie(t, begin, httpapi.ChallengeCookieSecure))
	if fin.StatusCode != http.StatusCreated {
		t.Fatalf("segunda passkey: %d", fin.StatusCode)
	}
	for _, c := range fin.Cookies() {
		if c.Name == httpapi.SessionCookieSecure && c.MaxAge >= 0 {
			t.Fatal("adicionar passkey nao deve abrir outra sessao")
		}
	}
	if len(f.store.Credentials()) != 2 {
		t.Fatal("a conta deve ter 2 passkeys")
	}
}
