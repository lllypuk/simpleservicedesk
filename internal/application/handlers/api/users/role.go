package users

import (
	"net/http"

	"simpleservicedesk/generated/openapi"
	"simpleservicedesk/internal/domain/users"

	"github.com/labstack/echo/v4"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

func (h UserHandlers) PatchUsersIDRole(c echo.Context, id openapi_types.UUID) error {
	ctx := c.Request().Context()

	var req openapi.UpdateUserRoleRequest
	if err := c.Bind(&req); err != nil {
		return err
	}

	// Parse role from request
	role, err := users.ParseRole(string(req.Role))
	if err != nil {
		msg := "invalid role: " + err.Error()
		return c.JSON(http.StatusBadRequest, openapi.ErrorResponse{Message: &msg})
	}

	// Update user role using service
	user, err := h.userService.UpdateUserRole(ctx, id, role)
	if err != nil {
		return handleUserError(c, err)
	}

	response := userToResponse(user)
	return c.JSON(http.StatusOK, response)
}
