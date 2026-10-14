package app_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/keven-rdr/wordloop/api/internal/identity/app"
	"github.com/keven-rdr/wordloop/api/internal/identity/domain"
	"github.com/keven-rdr/wordloop/api/internal/identity/identitytest"
)

var bg = context.Background()

func TestAuthenticateReturnsProfileAndRoleForValidSession(t *testing.T) {
	e := newEnv(t)
	done, _ := login(t, e, "/")
	who, err := e.auth(now.Add(time.Hour)).Run(bg, done.SessionID)
	if err != nil || who.User.DisplayName != "Aluno" || who.User.Role != domain.RoleUser {
		t.Fatalf("Authenticate = %+v, %v", who, err)
	}
	if who.SessionID == uuid.Nil || e.sessions.Touched != 1 {
		t.Fatalf("esperava id publico e 1 atualizacao do ultimo uso (veio %d)", e.sessions.Touched)
	}
	for name, id := range map[string]string{"vazio": "", "desconhecido": "naoexiste"} {
		if _, err := e.auth(now).Run(bg, id); !errors.Is(err, app.ErrNoSession) {
			t.Errorf("%s: esperava ErrNoSession, veio %v", name, err)
		}
	}
}

func TestAuthenticateTouchesAtMostOncePerMinute(t *testing.T) {
	e := newEnv(t)
	done, _ := login(t, e, "/")
	if _, err := e.auth(now.Add(30*time.Second)).Run(bg, done.SessionID); err != nil {
		t.Fatal(err)
	}
	if e.sessions.Touched != 0 {
		t.Fatalf("dentro de 1 minuto nao deve gravar o ultimo uso, gravou %d", e.sessions.Touched)
	}
}

func TestUserSessionExpiresByIdleAndByAbsoluteLimit(t *testing.T) {
	p := domain.DefaultSessionPolicy().User
	cases := map[string]struct {
		at   time.Duration
		keep bool
	}{
		"um minuto antes do ocioso": {p.Idle - time.Minute, true},
		"ocioso":                    {p.Idle + time.Second, false},
		"depois do teto absoluto":   {p.Absolute + time.Second, false},
	}
	for name, c := range cases {
		e := newEnv(t)
		done, _ := login(t, e, "/")
		_, err := e.auth(now.Add(c.at)).Run(bg, done.SessionID)
		if c.keep != (err == nil) {
			t.Errorf("%s: err = %v, esperava valida=%v", name, err, c.keep)
		}
		if !c.keep && e.sessions.Len() != 0 {
			t.Errorf("%s: a sessao vencida deve ser apagada", name)
		}
	}
}

func TestActiveUseKeepsUserSessionPastIdleLimitButNotPastAbsolute(t *testing.T) {
	e := newEnv(t)
	done, _ := login(t, e, "/")
	step := 5 * 24 * time.Hour
	for i := 1; i <= 5; i++ {
		if _, err := e.auth(now.Add(time.Duration(i)*step)).Run(bg, done.SessionID); err != nil {
			t.Fatalf("uso no dia %d deveria manter a sessao: %v", i*5, err)
		}
	}
	if _, err := e.auth(now.Add(31*24*time.Hour)).Run(bg, done.SessionID); !errors.Is(err, app.ErrNoSession) {
		t.Fatalf("passou de 30 dias: esperava ErrNoSession, veio %v", err)
	}
}

func TestPromotingToAdminShortensTheExistingSessionAtOnce(t *testing.T) {
	e := newEnv(t)
	done, _ := login(t, e, "/")
	stored, _ := e.sessions.Last()
	e.users.SetRole(stored.UserID, domain.RoleAdmin)
	idle := domain.DefaultSessionPolicy().Admin.Idle
	if _, err := e.auth(now.Add(idle+time.Minute)).Run(bg, done.SessionID); !errors.Is(err, app.ErrNoSession) {
		t.Fatalf("sessao de quem virou admin deve seguir o prazo curto, veio %v", err)
	}
}

func TestLoginRequiresProvenSecondFactorOnlyWhenTheAccountTurnedItOn(t *testing.T) {
	e := newEnv(t)
	login(t, e, "/")
	stored, _ := e.sessions.Last()
	e.users.SetRole(stored.UserID, domain.RoleAdmin)
	_ = e.sessions.DeleteAllByUser(bg, stored.UserID)

	started, _ := e.start().Run(bg, "/")
	off := app.FinishInput{Code: identitytest.GoodCode, State: e.flow(t, started.FlowCookie).State, FlowCookie: started.FlowCookie}
	if _, err := e.finish().Run(bg, off); err != nil {
		t.Fatalf("com o 2FA desligado o admin entra: %v", err)
	}
	_ = e.sessions.DeleteAllByUser(bg, stored.UserID)
	_ = e.users.SetMFARequired(bg, stored.UserID, true)

	started, _ = e.start().Run(bg, "/")
	in := app.FinishInput{Code: identitytest.GoodCode, State: e.flow(t, started.FlowCookie).State, FlowCookie: started.FlowCookie}
	if _, err := e.finish().Run(bg, in); !errors.Is(err, app.ErrMFARequired) {
		t.Fatalf("admin sem 2 fatores: esperava ErrMFARequired, veio %v", err)
	}
	if e.sessions.Len() != 0 {
		t.Fatal("nenhuma sessao pode nascer quando o admin nao provou 2 fatores")
	}

	e.provider.Identity.MFA = true
	started, _ = e.start().Run(bg, "/")
	in = app.FinishInput{Code: identitytest.GoodCode, State: e.flow(t, started.FlowCookie).State, FlowCookie: started.FlowCookie}
	done, err := e.finish().Run(bg, in)
	if err != nil {
		t.Fatalf("admin com 2 fatores deveria entrar: %v", err)
	}
	if want := now.Add(domain.DefaultSessionPolicy().Admin.Absolute); !done.ExpiresAt.Equal(want) {
		t.Fatalf("a sessao de admin expira em %v, esperava %v", done.ExpiresAt, want)
	}
}

func TestAuthenticatedAtNeverComesFromTheFuture(t *testing.T) {
	e := newEnv(t)
	e.provider.Identity.AuthTime = now.Add(48 * time.Hour)
	e.provider.Identity.MFA = true
	login(t, e, "/")
	stored, _ := e.sessions.Last()
	if !stored.AuthenticatedAt.Equal(now) || !stored.MFA {
		t.Fatalf("authenticated_at = %v (esperava %v), mfa = %v", stored.AuthenticatedAt, now, stored.MFA)
	}
}

func TestSessionOfRemovedUserIsNoSessionAndIsDeleted(t *testing.T) {
	e := newEnv(t)
	done, _ := login(t, e, "/")
	stored, _ := e.sessions.Last()
	e.users.Remove(stored.UserID)
	if _, err := e.auth(now).Run(bg, done.SessionID); !errors.Is(err, app.ErrNoSession) {
		t.Fatalf("esperava ErrNoSession, veio %v", err)
	}
	if e.sessions.Len() != 0 {
		t.Fatal("a sessao de usuario removido deve ser apagada")
	}
}

func otherUserSession(t *testing.T, e env) string {
	t.Helper()
	e.provider.Identity = domain.Identity{Issuer: "https://idp.test", Subject: "outra-pessoa", DisplayName: "Outra"}
	done, _ := login(t, e, "/")
	return done.SessionID
}

func TestListSessionsShowsOnlyMineAndMarksCurrent(t *testing.T) {
	e := newEnv(t)
	first, _ := login(t, e, "/")
	second, _ := login(t, e, "/")
	otherUserSession(t, e)
	list := app.ListSessions{Auth: e.auth(now), Sessions: e.sessions}
	views, err := list.Run(bg, second.SessionID)
	if err != nil || len(views) != 2 {
		t.Fatalf("esperava 2 sessoes minhas, veio %d (%v)", len(views), err)
	}
	current := 0
	for _, v := range views {
		if v.Current {
			current++
		}
	}
	if current != 1 {
		t.Fatalf("exatamente uma sessao deve vir como atual, vieram %d", current)
	}
	if _, err := list.Run(bg, first.SessionID); err != nil {
		t.Fatal(err)
	}
	if _, err := list.Run(bg, "forjado"); !errors.Is(err, app.ErrNoSession) {
		t.Fatalf("cookie forjado: esperava ErrNoSession, veio %v", err)
	}
}

func TestRevokeSessionCannotTouchAnotherPersonsSession(t *testing.T) {
	e := newEnv(t)
	mine, _ := login(t, e, "/")
	theirs := otherUserSession(t, e)
	theirList, err := app.ListSessions{Auth: e.auth(now), Sessions: e.sessions}.Run(bg, theirs)
	if err != nil || len(theirList) != 1 {
		t.Fatal("preparacao: esperava 1 sessao da outra pessoa")
	}
	revoke := app.RevokeSession{Auth: e.auth(now), Sessions: e.sessions}
	if _, err := revoke.Run(bg, mine.SessionID, theirList[0].ID); !errors.Is(err, app.ErrNotFound) {
		t.Fatalf("revogar sessao alheia: esperava ErrNotFound (igual a inexistente), veio %v", err)
	}
	if _, err := revoke.Run(bg, mine.SessionID, uuid.New()); !errors.Is(err, app.ErrNotFound) {
		t.Fatalf("id inexistente: esperava ErrNotFound, veio %v", err)
	}
	if _, err := (app.ListSessions{Auth: e.auth(now), Sessions: e.sessions}).Run(bg, theirs); err != nil {
		t.Fatalf("a sessao da outra pessoa deveria continuar valida: %v", err)
	}
}

func TestRevokeSessionEndsOtherDeviceAndReportsCurrent(t *testing.T) {
	e := newEnv(t)
	phone, _ := login(t, e, "/")
	laptop, _ := login(t, e, "/")
	revoke := app.RevokeSession{Auth: e.auth(now), Sessions: e.sessions}
	views, _ := app.ListSessions{Auth: e.auth(now), Sessions: e.sessions}.Run(bg, laptop.SessionID)
	var phoneID, laptopID uuid.UUID
	for _, v := range views {
		if v.Current {
			laptopID = v.ID
		} else {
			phoneID = v.ID
		}
	}
	if current, err := revoke.Run(bg, laptop.SessionID, phoneID); err != nil || current {
		t.Fatalf("derrubar o celular: current=%v err=%v", current, err)
	}
	if _, err := e.auth(now).Run(bg, phone.SessionID); !errors.Is(err, app.ErrNoSession) {
		t.Fatalf("a sessao derrubada nao pode mais autenticar: %v", err)
	}
	if current, err := revoke.Run(bg, laptop.SessionID, laptopID); err != nil || !current {
		t.Fatalf("derrubar a propria sessao: current=%v err=%v", current, err)
	}
}

func TestRevokeAllSessionsEndsEverythingOfMineOnly(t *testing.T) {
	e := newEnv(t)
	a, _ := login(t, e, "/")
	b, _ := login(t, e, "/")
	other := otherUserSession(t, e)
	if err := (app.RevokeAllSessions{Auth: e.auth(now), Sessions: e.sessions}).Run(bg, a.SessionID); err != nil {
		t.Fatal(err)
	}
	for _, id := range []string{a.SessionID, b.SessionID} {
		if _, err := e.auth(now).Run(bg, id); !errors.Is(err, app.ErrNoSession) {
			t.Fatalf("sessao minha deveria ter sido encerrada: %v", err)
		}
	}
	if _, err := e.auth(now).Run(bg, other); err != nil {
		t.Fatalf("a sessao da outra pessoa deve continuar: %v", err)
	}
	if err := (app.RevokeAllSessions{Auth: e.auth(now), Sessions: e.sessions}).Run(bg, "forjado"); !errors.Is(err, app.ErrNoSession) {
		t.Fatalf("cookie forjado: esperava ErrNoSession, veio %v", err)
	}
}
