package users

import (
	"net/http"

	"simpleservicedesk/generated/openapi"

	"github.com/labstack/echo/v4"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

func (h UserHandlers) GetUsersIdTickets(c echo.Context, _ openapi_types.UUID, _ openapi.GetUsersIdTicketsParams) error {
	msg := "user tickets functionality not implemented yet"
	return c.JSON(http.StatusNotImplemented, openapi.ErrorResponse{Message: &msg})
}
