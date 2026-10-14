package passkeys_test

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/keven-rdr/wordloop/api/internal/identity/adapters/webauthn"
	"github.com/keven-rdr/wordloop/api/internal/identity/app"
	"github.com/keven-rdr/wordloop/api/internal/identity/domain"
	"github.com/keven-rdr/wordloop/api/internal/identity/identitytest"
	"github.com/keven-rdr/wordloop/api/internal/identity/passkeys"
	"github.com/keven-rdr/wordloop/api/internal/platform/seal"
)

const (
	origin = "https://tst.wordloop.com.br"
	rpID   = "tst.wordloop.com.br"
	invite = "convite-secreto-de-teste-0123456789abcdefghijklmn"
)

var (
	now = time.Date(2026, 10, 15, 12, 0, 0, 0, time.UTC)
	bg  = context.Background()
)

type env struct {
	store      *identitytest.PasskeyStore
	challenges *identitytest.MemChallenges
	users      *identitytest.Users
	sessions   *identitytest.Sessions
	ceremony   passkeys.Ceremony
	owner      passkeys.Owner
}

func newEnv(t *testing.T, role domain.Role) *env {
	t.Helper()
	cer, err := webauthn.New(webauthn.Config{RPID: rpID, RPName: "wordloop", Origins: []string{origin}})
	if err != nil {
		t.Fatal(err)
	}
	users := &identitytest.Users{}
	user, _ := users.Upsert(bg, domain.Identity{Issuer: "i", Subject: "s", DisplayName: "Voce"})
	users.SetRole(user.ID, role)
	e := &env{
		store: identitytest.NewPasskeyStore(), challenges: identitytest.NewMemChallenges(), users: users,
		sessions: &identitytest.Sessions{}, ceremony: cer,
		owner: passkeys.Owner{ID: user.ID, Email: "voce@exemplo.com", DisplayName: "Voce"},
	}
	e.store.AddOwner(e.owner)
	e.store.AddInvite(invite, e.owner.ID, now.Add(15*time.Minute))
	return e
}

func (e *env) clock() func() time.Time { return func() time.Time { return now } }

func (e *env) opener() app.SessionOpener {
	sealer, _ := seal.New(bytes.Repeat([]byte{7}, seal.KeySize))
	return app.SessionOpener{Sessions: e.sessions, Sealer: sealer, Now: e.clock()}
}

func (e *env) begin() passkeys.BeginRegistration {
	return passkeys.BeginRegistration{Ceremony: e.ceremony, Store: e.store, Challenges: e.challenges, Now: e.clock()}
}

func (e *env) finish() passkeys.FinishRegistration {
	return passkeys.FinishRegistration{
		Ceremony: e.ceremony, Store: e.store, Challenges: e.challenges, Users: e.users, Opener: e.opener(), Now: e.clock(),
	}
}

func (e *env) loginBegin() passkeys.BeginLogin {
	return passkeys.BeginLogin{Ceremony: e.ceremony, Challenges: e.challenges, Now: e.clock()}
}

func (e *env) loginFinish() passkeys.FinishLogin {
	return passkeys.FinishLogin{
		Ceremony: e.ceremony, Store: e.store, Challenges: e.challenges, Users: e.users, Opener: e.opener(), Now: e.clock(),
	}
}

func (e *env) registerWithInvite(t *testing.T, a *identitytest.SoftAuthenticator) passkeys.Registered {
	t.Helper()
	st, err := e.begin().Run(bg, passkeys.Subject{InviteToken: invite})
	if err != nil {
		t.Fatal(err)
	}
	resp, _ := a.Create(st.Options)
	got, err := e.finish().Run(bg, passkeys.FinishRegistrationInput{ChallengeID: st.ChallengeID, Credential: resp, Name: "Meu celular", UserAgent: "ua"})
	if err != nil {
		t.Fatal(err)
	}
	return got
}

func (e *env) registerSecond(t *testing.T) {
	t.Helper()
	st, _ := e.begin().Run(bg, passkeys.Subject{SessionUserID: e.owner.ID})
	resp, _ := identitytest.NewSoftAuthenticator(origin, rpID).Create(st.Options)
	if _, err := e.finish().Run(bg, passkeys.FinishRegistrationInput{ChallengeID: st.ChallengeID, Credential: resp}); err != nil {
		t.Fatal(err)
	}
}

func (e *env) uid() []byte {
	id := e.owner.ID
	return id[:]
}
