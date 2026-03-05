package users

import (
	"net/http"

	"simpleservicedesk/generated/openapi"
	userdomain "simpleservicedesk/internal/domain/users"
	"simpleservicedesk/pkg/echomiddleware"

	"github.com/labstack/echo/v4"
)

func requireAdmin(c echo.Context) bool {
	claims, ok := echomiddleware.GetAuthClaimsFromContext(c.Request().Context())
	if !ok || claims == nil {
		msg := "unauthorized"
		_ = c.JSON(http.StatusUnauthorized, openapi.ErrorResponse{Message: &msg})
		return false
	}

	if claims.Role != userdomain.RoleAdmin {
		msg := "forbidden"
		_ = c.JSON(http.StatusForbidden, openapi.ErrorResponse{Message: &msg})
		return false
	}

	return true
}
