package tickets

import (
	"net/http"
	"strings"

	userdomain "simpleservicedesk/internal/domain/users"
	"simpleservicedesk/pkg/echomiddleware"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func authUser(c echo.Context) (uuid.UUID, userdomain.Role, bool) {
	claims, ok := echomiddleware.GetAuthClaims(c)
	if !ok || claims == nil {
		_ = c.NoContent(http.StatusUnauthorized)
		return uuid.Nil, "", false
	}

	userID, err := uuid.Parse(strings.TrimSpace(claims.UserID))
	if err != nil {
		_ = c.NoContent(http.StatusUnauthorized)
		return uuid.Nil, "", false
	}

	return userID, claims.Role, true
}

func hasElevatedTicketAccess(role userdomain.Role) bool {
	return role == userdomain.RoleAgent || role == userdomain.RoleAdmin
}
