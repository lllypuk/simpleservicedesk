package users

import (
	"net/http"

	"simpleservicedesk/internal/domain/users"

	"github.com/labstack/echo/v4"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

func (h UserHandlers) DeleteUsersId(c echo.Context, id openapi_types.UUID) error {
	ctx := c.Request().Context()

	// Сначала проверим, существует ли пользователь
	_, err := h.repo.GetUser(ctx, id)
	if err != nil {
		return handleUserError(c, err)
	}

	// Выполняем мягкое удаление - деактивируем пользователя
	// Вместо полного удаления из базы данных
	_, err = h.repo.UpdateUser(ctx, id, func(user *users.User) (bool, error) {
		if !user.IsActive() {
			return false, nil // Уже деактивирован
		}
		user.Deactivate()
		return true, nil
	})

	if err != nil {
		return handleUserError(c, err)
	}

	return c.NoContent(http.StatusNoContent)
}
