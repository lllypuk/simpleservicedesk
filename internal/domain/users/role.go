package users

import (
	"errors"
	"strings"
)

var (
	ErrInvalidRole = errors.New("invalid user role")
)

// Role представляет роль пользователя в системе
type Role string

const (
	RoleCustomer Role = "customer" // Клиент - создает заявки
	RoleAgent    Role = "agent"    // Агент - обрабатывает заявки
	RoleAdmin    Role = "admin"    // Администратор - полный доступ
)

// AllRoles возвращает все возможные роли
func AllRoles() []Role {
	return []Role{
		RoleCustomer,
		RoleAgent,
		RoleAdmin,
	}
}

// String возвращает строковое представление роли
func (r Role) String() string {
	return string(r)
}

// IsValid проверяет, является ли роль валидной
func (r Role) IsValid() bool {
	for _, role := range AllRoles() {
		if r == role {
			return true
		}
	}
	return false
}

// ParseRole преобразует строку в роль
func ParseRole(s string) (Role, error) {
	role := Role(strings.ToLower(strings.TrimSpace(s)))
	if !role.IsValid() {
		return "", ErrInvalidRole
	}
	return role, nil
}

// DisplayName возвращает человекочитаемое название роли
func (r Role) DisplayName() string {
	names := map[Role]string{
		RoleCustomer: "Клиент",
		RoleAgent:    "Агент",
		RoleAdmin:    "Администратор",
	}

	if name, exists := names[r]; exists {
		return name
	}
	return string(r)
}

// CanCreateTickets проверяет, может ли пользователь с данной ролью создавать заявки
func (r Role) CanCreateTickets() bool {
	return r == RoleCustomer || r == RoleAgent || r == RoleAdmin
}

// CanAssignTickets проверяет, может ли пользователь назначать заявки другим
func (r Role) CanAssignTickets() bool {
	return r == RoleAgent || r == RoleAdmin
}

// CanViewAllTickets проверяет, может ли пользователь видеть все заявки организации
func (r Role) CanViewAllTickets() bool {
	return r == RoleAgent || r == RoleAdmin
}

// CanManageUsers проверяет, может ли пользователь управлять другими пользователями
func (r Role) CanManageUsers() bool {
	return r == RoleAdmin
}

// CanManageOrganization проверяет, может ли пользователь управлять настройками организации
func (r Role) CanManageOrganization() bool {
	return r == RoleAdmin
}

// CanViewInternalComments проверяет, может ли пользователь видеть внутренние комментарии
func (r Role) CanViewInternalComments() bool {
	return r == RoleAgent || r == RoleAdmin
}

// CanCreateInternalComments проверяет, может ли пользователь создавать внутренние комментарии
func (r Role) CanCreateInternalComments() bool {
	return r == RoleAgent || r == RoleAdmin
}

// Level возвращает уровень доступа роли (для сравнения)
func (r Role) Level() int {
	levels := map[Role]int{
		RoleCustomer: 1,
		RoleAgent:    2,
		RoleAdmin:    3,
	}

	if level, exists := levels[r]; exists {
		return level
	}
	return 0
}

// HasHigherOrEqualLevel проверяет, имеет ли роль уровень доступа выше или равный указанной роли
func (r Role) HasHigherOrEqualLevel(other Role) bool {
	return r.Level() >= other.Level()
}
