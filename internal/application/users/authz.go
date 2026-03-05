package users

import (
	"net/http"
	"strings"

	userdomain "simpleservicedesk/internal/domain/users"
	"simpleservicedesk/pkg/echomiddleware"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func authorizeSelfOrAdmin(c echo.Context, userID uuid.UUID) (bool, bool) {
	claims, ok := echomiddleware.GetAuthClaims(c)
	if !ok || claims == nil {
		_ = c.NoContent(http.StatusUnauthorized)
		return false, false
	}

	if claims.Role == userdomain.RoleAdmin {
		return true, true
	}

	if claims.UserID == userID.String() {
		return false, true
	}

	_ = c.NoContent(http.StatusForbidden)
	return false, false
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}
