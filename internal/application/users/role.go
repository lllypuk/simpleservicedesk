package users

import (
	"net/http"

	"simpleservicedesk/generated/openapi"

	"github.com/labstack/echo/v4"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

func (h UserHandlers) PatchUsersIdRole(c echo.Context, _ openapi_types.UUID) error {
	msg := "user role functionality not implemented yet"
	return c.JSON(http.StatusNotImplemented, openapi.ErrorResponse{Message: &msg})
}
