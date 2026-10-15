package httpapi

import (
	"encoding/json"
	"net/http"
)

// Respostas com Set-Cookie: o servidor gerado so declara um cabecalho por resposta, entao estas implementam
// diretamente os metodos Visit...Response da interface strict.

// redirect e o 302 de StartLogin e FinishLogin.
type redirect struct {
	location string
	cookies  []*http.Cookie
}

func (r redirect) write(w http.ResponseWriter) error {
	for _, c := range r.cookies {
		http.SetCookie(w, c)
	}
	w.Header().Set("Location", r.location)
	w.WriteHeader(http.StatusFound)
	return nil
}

func (r redirect) VisitStartLoginResponse(w http.ResponseWriter) error  { return r.write(w) }
func (r redirect) VisitFinishLoginResponse(w http.ResponseWriter) error { return r.write(w) }

type loginFailed struct {
	code    string
	cookies []*http.Cookie
}

func (r loginFailed) VisitFinishLoginResponse(w http.ResponseWriter) error {
	for _, c := range r.cookies {
		http.SetCookie(w, c)
	}
	w.Header().Set("Location", "/login?error="+r.code)
	w.WriteHeader(http.StatusFound)
	return nil
}

// loggedOut e o 204 do logout; apaga o cookie da sessao.
type loggedOut struct{ cookies []*http.Cookie }

func (r loggedOut) VisitLogoutResponse(w http.ResponseWriter) error {
	for _, c := range r.cookies {
		http.SetCookie(w, c)
	}
	w.WriteHeader(http.StatusNoContent)
	return nil
}

type sessionEnded struct{ cookies []*http.Cookie }

func (r sessionEnded) write(w http.ResponseWriter) error {
	for _, c := range r.cookies {
		http.SetCookie(w, c)
	}
	w.WriteHeader(http.StatusNoContent)
	return nil
}

func (r sessionEnded) VisitRevokeSessionResponse(w http.ResponseWriter) error     { return r.write(w) }
func (r sessionEnded) VisitRevokeAllSessionsResponse(w http.ResponseWriter) error { return r.write(w) }

type jsonReply struct {
	status  int
	body    any
	cookies []*http.Cookie
}

func (r jsonReply) write(w http.ResponseWriter) error {
	for _, c := range r.cookies {
		http.SetCookie(w, c)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(r.status)
	return json.NewEncoder(w).Encode(r.body)
}

func (r jsonReply) VisitBeginPasskeyRegistrationResponse(w http.ResponseWriter) error {
	return r.write(w)
}
func (r jsonReply) VisitFinishPasskeyRegistrationResponse(w http.ResponseWriter) error {
	return r.write(w)
}
func (r jsonReply) VisitBeginPasskeyLoginResponse(w http.ResponseWriter) error { return r.write(w) }

type statusReply struct {
	status  int
	cookies []*http.Cookie
}

func (r statusReply) write(w http.ResponseWriter) error {
	for _, c := range r.cookies {
		http.SetCookie(w, c)
	}
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(r.status)
	return nil
}

func (r statusReply) VisitFinishPasskeyRegistrationResponse(w http.ResponseWriter) error {
	return r.write(w)
}
func (r statusReply) VisitFinishPasskeyLoginResponse(w http.ResponseWriter) error { return r.write(w) }
