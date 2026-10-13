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
