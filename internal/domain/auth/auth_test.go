package auth_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/require"

	authdomain "simpleservicedesk/internal/domain/auth"
	"simpleservicedesk/internal/domain/users"
)

func TestClaimsImplementsJWTClaims(t *testing.T) {
	claims := authdomain.Claims{}

	var jwtClaims jwt.Claims = claims
	require.NotNil(t, jwtClaims)
}

func TestClaimsContainsCustomAndRegisteredFields(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	expiresAt := now.Add(2 * time.Hour)

	claims := authdomain.Claims{
		UserID: "2be568f4-50af-4658-bcef-4ce26ef48a95",
		Role:   users.RoleAdmin,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "2be568f4-50af-4658-bcef-4ce26ef48a95",
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	}

	require.Equal(t, "2be568f4-50af-4658-bcef-4ce26ef48a95", claims.UserID)
	require.Equal(t, users.RoleAdmin, claims.Role)
	require.Equal(t, "2be568f4-50af-4658-bcef-4ce26ef48a95", claims.Subject)
	require.Equal(t, now, claims.IssuedAt.Time)
	require.Equal(t, now, claims.NotBefore.Time)
	require.Equal(t, expiresAt, claims.ExpiresAt.Time)
}

func TestLoginTypesJSONTags(t *testing.T) {
	request := authdomain.LoginRequest{
		Email:      "admin@example.com",
		Passphrase: "secure-password",
	}

	requestPayload, err := json.Marshal(request)
	require.NoError(t, err)
	require.JSONEq(t, `{"email":"admin@example.com","password":"secure-password"}`, string(requestPayload))

	response := authdomain.LoginResponse{Token: "jwt-token-value"}
	responsePayload, err := json.Marshal(response)
	require.NoError(t, err)
	require.JSONEq(t, `{"token":"jwt-token-value"}`, string(responsePayload))
}
