package passwords

import (
	"strings"
	"unicode/utf8"
)

func CheckPolicy(password, email string) error {
	n := utf8.RuneCountInString(password)
	if n < MinLength || n > MaxLength {
		return ErrWeak
	}
	local, _, _ := strings.Cut(strings.ToLower(email), "@")
	lowered := strings.ToLower(password)
	if lowered == strings.ToLower(email) || (local != "" && lowered == local) {
		return ErrWeak
	}
	distinct := map[rune]struct{}{}
	for _, r := range password {
		distinct[r] = struct{}{}
	}
	if len(distinct) < 4 {
		return ErrWeak
	}
	return nil
}
