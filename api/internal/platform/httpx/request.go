// Package httpx reune ajudas de HTTP que o servidor gerado (strict) nao oferece, como ler cookies do pedido.
package httpx

import (
	"context"
	"net/http"
)

type requestKey struct{}

// WithRequest guarda o *http.Request no contexto; o handler gerado (strict) so entrega parametros declarados.
func WithRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), requestKey{}, r)))
	})
}

// Cookie devolve o valor do cookie do pedido em andamento, ou "" se ausente.
func Cookie(ctx context.Context, name string) string {
	r, _ := ctx.Value(requestKey{}).(*http.Request)
	if r == nil {
		return ""
	}
	c, err := r.Cookie(name)
	if err != nil {
		return ""
	}
	return c.Value
}

// UserAgent devolve o User-Agent do pedido em andamento (para listar sessoes), limitado a 200 caracteres.
func UserAgent(ctx context.Context) string {
	r, _ := ctx.Value(requestKey{}).(*http.Request)
	if r == nil {
		return ""
	}
	ua := r.UserAgent()
	if len(ua) > 200 {
		ua = ua[:200]
	}
	return ua
}
