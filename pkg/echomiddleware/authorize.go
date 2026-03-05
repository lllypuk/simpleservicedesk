package echomiddleware

import (
	"net/http"

	"simpleservicedesk/internal/domain/users"

	"github.com/labstack/echo/v4"
)

// RequireRole allows request execution only for users with role level >= minRole.
func RequireRole(minRole users.Role) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			claims, ok := GetAuthClaims(c)
			if !ok || claims == nil {
				return c.NoContent(http.StatusUnauthorized)
			}

			if !claims.Role.IsValid() || !minRole.IsValid() {
				return c.NoContent(http.StatusForbidden)
			}

			if !claims.Role.HasHigherOrEqualLevel(minRole) {
				return c.NoContent(http.StatusForbidden)
			}

			return next(c)
		}
	}
}

// IsOwnerOrRole checks resource ownership or minimum role requirement.
func IsOwnerOrRole(c echo.Context, userID string, minRole users.Role) bool {
	claims, ok := GetAuthClaims(c)
	if !ok || claims == nil {
		return false
	}

	if userID != "" && claims.UserID == userID {
		return true
	}

	if !claims.Role.IsValid() || !minRole.IsValid() {
		return false
	}

	return claims.Role.HasHigherOrEqualLevel(minRole)
}
