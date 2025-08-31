package users

import (
	"net/http"

	"github.com/labstack/echo/v4"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

func (h UserHandlers) DeleteUsersID(c echo.Context, id openapi_types.UUID) error {
	ctx := c.Request().Context()

	// Check if user exists first
	_, err := h.userService.GetUser(ctx, id)
	if err != nil {
		return handleUserError(c, err)
	}

	// Delete user using service
	err = h.userService.DeleteUser(ctx, id)
	if err != nil {
		return handleUserError(c, err)
	}

	return c.NoContent(http.StatusNoContent)
}
