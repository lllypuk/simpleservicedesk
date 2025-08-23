package users

import (
	"net/http"

	"simpleservicedesk/generated/openapi"

	"github.com/labstack/echo/v4"
)

func (h UserHandlers) GetUsers(c echo.Context, _ openapi.GetUsersParams) error {
	msg := "user list functionality not implemented yet"
	return c.JSON(http.StatusNotImplemented, openapi.ErrorResponse{Message: &msg})
}
