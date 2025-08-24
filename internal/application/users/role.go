package users

import (
	"net/http"

	"simpleservicedesk/generated/openapi"
	"simpleservicedesk/internal/domain/users"

	"github.com/labstack/echo/v4"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

func (h UserHandlers) PatchUsersIdRole(c echo.Context, id openapi_types.UUID) error {
	ctx := c.Request().Context()

	var req openapi.UpdateUserRoleRequest
	if err := c.Bind(&req); err != nil {
		return err
	}

	// Парсим роль из запроса
	role, err := users.ParseRole(string(req.Role))
	if err != nil {
		msg := "invalid role"
		return c.JSON(http.StatusBadRequest, openapi.ErrorResponse{Message: &msg})
	}

	user, err := h.repo.UpdateUser(ctx, id, func(user *users.User) (bool, error) {
		if user.Role() == role {
			return false, nil // Роль уже установлена
		}

		if changeErr := user.ChangeRole(role); changeErr != nil {
			return false, changeErr
		}
		return true, nil
	})

	if err != nil {
		return handleUserError(c, err)
	}

	response := userToResponse(user)
	return c.JSON(http.StatusOK, response)
}
