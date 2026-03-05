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

// LoginRequest contains credentials for authentication.
type LoginRequest struct {
	Email      string `json:"email"`
	Passphrase string `json:"password"` //nolint:gosec // Password is user-provided request data, not a hardcoded secret.
}

// LoginResponse contains an access token returned after successful login.
type LoginResponse struct {
	Token string `json:"token"`
}
