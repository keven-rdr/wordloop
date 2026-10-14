package accounts_test

import (
	"bytes"
	"context"
	"crypto/sha256"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/keven-rdr/wordloop/api/internal/identity/accounts"
	"github.com/keven-rdr/wordloop/api/internal/identity/domain"
)

var now = time.Date(2026, 10, 14, 12, 0, 0, 0, time.UTC)

type fakeStore struct {
	byEmail map[string]accounts.Account
	names   map[uuid.UUID]string
	invites map[uuid.UUID][][]byte
	expires map[uuid.UUID]time.Time
}

func newStore() *fakeStore {
	return &fakeStore{
		byEmail: map[string]accounts.Account{}, names: map[uuid.UUID]string{},
		invites: map[uuid.UUID][][]byte{}, expires: map[uuid.UUID]time.Time{},
	}
}

func (s *fakeStore) FindByEmail(_ context.Context, email string) (accounts.Account, error) {
	if a, ok := s.byEmail[email]; ok {
		return a, nil
	}
	return accounts.Account{}, accounts.ErrNotFound
}

func (s *fakeStore) CreateAdmin(_ context.Context, email, name string, _ time.Time) (uuid.UUID, error) {
	if _, ok := s.byEmail[email]; ok {
		return uuid.Nil, accounts.ErrEmailTaken
	}
	id := uuid.New()
	s.byEmail[email] = accounts.Account{ID: id, Role: domain.RoleAdmin}
	s.names[id] = name
	return id, nil
}

func (s *fakeStore) ReplaceInvite(_ context.Context, userID uuid.UUID, hash []byte, _, expiresAt time.Time) error {
	s.invites[userID] = [][]byte{hash}
	s.expires[userID] = expiresAt
	return nil
}

func useCase(s *fakeStore) accounts.CreateAdmin {
	return accounts.CreateAdmin{Store: s, PublicURL: "https://tst.wordloop.com.br/", Now: func() time.Time { return now }}
}

func tokenOf(t *testing.T, inv accounts.Invite) string {
	t.Helper()
	prefix := "https://tst.wordloop.com.br/invite#"
	if !strings.HasPrefix(inv.URL, prefix) {
		t.Fatalf("o link deve abrir /invite com o segredo no fragmento: %q", inv.URL)
	}
	return strings.TrimPrefix(inv.URL, prefix)
}

func TestCreateAdminCreatesAccountAndStoresOnlyTheTokenHash(t *testing.T) {
	s := newStore()
	inv, err := useCase(s).Run(context.Background(), "  Voce@Exemplo.COM ", "")
	if err != nil {
		t.Fatal(err)
	}
	acc, ok := s.byEmail["voce@exemplo.com"]
	if !ok || acc.Role != domain.RoleAdmin || inv.Email != "voce@exemplo.com" || inv.Reissued {
		t.Fatalf("conta/convite inesperados: %+v %+v", acc, inv)
	}
	if s.names[acc.ID] != "voce" {
		t.Fatalf("nome padrao deve ser a parte antes do @, veio %q", s.names[acc.ID])
	}
	token := tokenOf(t, inv)
	sum := sha256.Sum256([]byte(token))
	if len(token) < 43 || len(s.invites[acc.ID]) != 1 || !bytes.Equal(s.invites[acc.ID][0], sum[:]) {
		t.Fatal("o banco deve guardar o SHA-256 do token (256 bits), nunca o token")
	}
	if !inv.ExpiresAt.Equal(now.Add(accounts.InviteTTL)) || !s.expires[acc.ID].Equal(inv.ExpiresAt) {
		t.Fatalf("o convite deve valer %v: %v", accounts.InviteTTL, inv.ExpiresAt)
	}
}

func TestCreateAdminUsesTheGivenName(t *testing.T) {
	s := newStore()
	if _, err := useCase(s).Run(context.Background(), "a@b.co", " Keven "); err != nil {
		t.Fatal(err)
	}
	if got := s.names[s.byEmail["a@b.co"].ID]; got != "Keven" {
		t.Fatalf("nome = %q", got)
	}
}

func TestCreateAdminReissuesForAnAdminWithoutCredentialsAndReplacesTheOldInvite(t *testing.T) {
	s := newStore()
	first, _ := useCase(s).Run(context.Background(), "a@b.co", "")
	second, err := useCase(s).Run(context.Background(), "A@B.co", "")
	if err != nil || !second.Reissued {
		t.Fatalf("esperava reemissao: %+v %v", second, err)
	}
	if tokenOf(t, first) == tokenOf(t, second) {
		t.Fatal("o convite novo precisa ser outro token")
	}
	id := s.byEmail["a@b.co"].ID
	sum := sha256.Sum256([]byte(tokenOf(t, second)))
	if len(s.invites[id]) != 1 || !bytes.Equal(s.invites[id][0], sum[:]) {
		t.Fatal("so o convite novo pode valer")
	}
}

func TestCreateAdminNeverTakesOverAnExistingAccount(t *testing.T) {
	cases := map[string]accounts.Account{
		"aluno comum":                  {ID: uuid.New(), Role: domain.RoleUser},
		"admin que ja tem como entrar": {ID: uuid.New(), Role: domain.RoleAdmin, HasCredentials: true},
		"aluno com credencial":         {ID: uuid.New(), Role: domain.RoleUser, HasCredentials: true},
	}
	for name, acc := range cases {
		s := newStore()
		s.byEmail["a@b.co"] = acc
		if _, err := useCase(s).Run(context.Background(), "a@b.co", ""); !errors.Is(err, accounts.ErrAccountExists) {
			t.Errorf("%s: esperava ErrAccountExists, veio %v", name, err)
		}
		if len(s.invites) != 0 {
			t.Errorf("%s: nenhum convite pode ser emitido", name)
		}
	}
}

func TestCreateAdminRejectsInvalidEmails(t *testing.T) {
	for _, in := range []string{"", "sem-arroba", "a@b", "Nome <a@b.co>", "a@b.co, c@d.co", "a b@c.co", "<a@b.co>", strings.Repeat("a", 250) + "@b.co"} {
		s := newStore()
		if _, err := useCase(s).Run(context.Background(), in, ""); !errors.Is(err, accounts.ErrInvalidEmail) {
			t.Errorf("%q: esperava ErrInvalidEmail, veio %v", in, err)
		}
		if len(s.byEmail) != 0 {
			t.Errorf("%q: nenhuma conta pode ser criada", in)
		}
	}
}

func TestCreateAdminRequiresAValidPublicURL(t *testing.T) {
	for _, in := range []string{"", "tst.wordloop.com.br", "ftp://x.co", "https://x.co/app", "https://x.co?a=1", "https://x.co#f", "https://"} {
		uc := accounts.CreateAdmin{Store: newStore(), PublicURL: in}
		if _, err := uc.Run(context.Background(), "a@b.co", ""); !errors.Is(err, accounts.ErrInvalidURL) {
			t.Errorf("%q: esperava ErrInvalidURL, veio %v", in, err)
		}
	}
	uc := accounts.CreateAdmin{Store: newStore(), PublicURL: "http://localhost:5173"}
	inv, err := uc.Run(context.Background(), "a@b.co", "")
	if err != nil || !strings.HasPrefix(inv.URL, "http://localhost:5173/invite#") {
		t.Fatalf("dev local: %+v %v", inv, err)
	}
}

func TestTwoInvitesNeverShareAToken(t *testing.T) {
	seen := map[string]bool{}
	for i := range 20 {
		s := newStore()
		inv, err := useCase(s).Run(context.Background(), "a"+string(rune('a'+i))+"@b.co", "")
		if err != nil {
			t.Fatal(err)
		}
		tok := tokenOf(t, inv)
		if seen[tok] {
			t.Fatal("token repetido")
		}
		seen[tok] = true
	}
}
