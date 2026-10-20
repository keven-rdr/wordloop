package httpx

import (
	"context"
	"net/http"
)

type requestKey struct{}

func WithRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), requestKey{}, r)))
	})
}

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
