package users

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
	openapitypes "github.com/oapi-codegen/runtime/types"

	"simpleservicedesk/generated/openapi"
	"simpleservicedesk/internal/domain/users"
)

func (h UserHandlers) GetUsersID(c echo.Context, id openapitypes.UUID) error {
	user, err := h.repo.GetUser(c.Request().Context(), id)
	if err != nil {
		msg := err.Error()
		if errors.Is(err, users.ErrUserNotFound) {
			return c.JSON(http.StatusNotFound, openapi.ErrorResponse{Message: &msg})
		}
		return c.JSON(http.StatusInternalServerError, openapi.ErrorResponse{Message: &msg})
	}

	name := user.Name()
	email := openapitypes.Email(user.Email())
	return c.JSON(http.StatusOK, openapi.GetUserResponse{
		Id:    &id,
		Name:  &name,
		Email: &email,
	})
}
