package users

import (
	"net/http"

	"github.com/labstack/echo/v4"
	openapitypes "github.com/oapi-codegen/runtime/types"
)

func (h UserHandlers) GetUsersID(c echo.Context, id openapitypes.UUID) error {
	user, err := h.repo.GetUser(c.Request().Context(), id)
	if err != nil {
		return handleUserError(c, err)
	}

	response := userToResponse(user)
	return c.JSON(http.StatusOK, response)
}
