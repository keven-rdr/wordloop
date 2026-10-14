package domain

// Role e o papel do usuario. Vem SO da coluna app_user.role; nenhum pedido, corpo ou claim o altera.
type Role string

const (
	RoleUser  Role = "user"
	RoleAdmin Role = "admin"
)

// ParseRole converte o valor do banco. Qualquer coisa diferente de "admin" e usuario comum (falha para o lado seguro).
func ParseRole(s string) Role {
	if s == string(RoleAdmin) {
		return RoleAdmin
	}
	return RoleUser
}
