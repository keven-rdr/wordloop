package domain

import (
	"strings"
	"testing"
)

func TestCleanPictureURLOnlyKeepsHTTPS(t *testing.T) {
	cases := map[string]struct{ in, want string }{
		"https":              {"https://lh3.googleusercontent.com/a/x", "https://lh3.googleusercontent.com/a/x"},
		"com espacos":        {"  https://avatars.githubusercontent.com/u/1  ", "https://avatars.githubusercontent.com/u/1"},
		"http":               {"http://exemplo.com/a.png", ""},
		"javascript":         {"javascript:alert(1)", ""},
		"data":               {"data:image/png;base64,AAAA", ""},
		"vazio":              {"", ""},
		"grande demais":      {"https://" + strings.Repeat("a", 2100), ""},
		"relativo":           {"/media/a.png", ""},
		"https em maiusculo": {"HTTPS://exemplo.com/a.png", ""},
	}
	for name, c := range cases {
		if got := CleanPictureURL(c.in); got != c.want {
			t.Errorf("%s: CleanPictureURL(%q) = %q, esperava %q", name, c.in, got, c.want)
		}
	}
}
