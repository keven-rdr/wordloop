package httpapi

import "net/http"

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

// badLogin e o 400 do callback; limpa o cookie de estado para o login nao ser repetido.
type badLogin struct{ cookies []*http.Cookie }

func (r badLogin) VisitFinishLoginResponse(w http.ResponseWriter) error {
	for _, c := range r.cookies {
		http.SetCookie(w, c)
	}
	w.WriteHeader(http.StatusBadRequest)
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

// mfaRequired e o 403 do callback: conta de administrador sem 2 fatores; limpa o cookie de estado.
type mfaRequired struct{ cookies []*http.Cookie }

func (r mfaRequired) VisitFinishLoginResponse(w http.ResponseWriter) error {
	for _, c := range r.cookies {
		http.SetCookie(w, c)
	}
	w.WriteHeader(http.StatusForbidden)
	return nil
}

// sessionEnded e o 204 de encerrar sessao quando a sessao atual foi encerrada: apaga o cookie.
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
