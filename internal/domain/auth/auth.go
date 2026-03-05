package auth

import (
	"github.com/golang-jwt/jwt/v5"

	"simpleservicedesk/internal/domain/users"
)

// Claims describes authenticated user identity inside JWT payload.
type Claims struct {
	jwt.RegisteredClaims

	UserID string     `json:"user_id"`
	Role   users.Role `json:"role"`
}
