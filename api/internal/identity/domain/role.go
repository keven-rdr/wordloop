package domain

type Role string

const (
	RoleUser  Role = "user"
	RoleAdmin Role = "admin"
)

func ParseRole(s string) Role {
	if s == string(RoleAdmin) {
		return RoleAdmin
	}
	return RoleUser
}
