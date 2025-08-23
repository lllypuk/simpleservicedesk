package organizations

import (
	"net/http"

	"simpleservicedesk/generated/openapi"

	"github.com/labstack/echo/v4"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

const notImplementedMsg = "organizations functionality not implemented yet"

type OrganizationHandlers struct{}

func SetupHandlers() OrganizationHandlers {
	return OrganizationHandlers{}
}

func (h OrganizationHandlers) GetOrganizations(c echo.Context, _ openapi.GetOrganizationsParams) error {
	msg := notImplementedMsg
	return c.JSON(http.StatusNotImplemented, openapi.ErrorResponse{Message: &msg})
}

func (h OrganizationHandlers) PostOrganizations(c echo.Context) error {
	msg := notImplementedMsg
	return c.JSON(http.StatusNotImplemented, openapi.ErrorResponse{Message: &msg})
}

func (h OrganizationHandlers) GetOrganizationsID(c echo.Context, _ openapi_types.UUID) error {
	msg := notImplementedMsg
	return c.JSON(http.StatusNotImplemented, openapi.ErrorResponse{Message: &msg})
}

func (h OrganizationHandlers) GetOrganizationsIdTickets(c echo.Context, _ openapi_types.UUID, _ openapi.GetOrganizationsIdTicketsParams) error {
	msg := notImplementedMsg
	return c.JSON(http.StatusNotImplemented, openapi.ErrorResponse{Message: &msg})
}

func (h OrganizationHandlers) GetOrganizationsIdUsers(c echo.Context, _ openapi_types.UUID, _ openapi.GetOrganizationsIdUsersParams) error {
	msg := notImplementedMsg
	return c.JSON(http.StatusNotImplemented, openapi.ErrorResponse{Message: &msg})
}

func (h OrganizationHandlers) PutOrganizationsId(c echo.Context, _ openapi_types.UUID) error {
	msg := notImplementedMsg
	return c.JSON(http.StatusNotImplemented, openapi.ErrorResponse{Message: &msg})
}

func (h OrganizationHandlers) DeleteOrganizationsId(c echo.Context, _ openapi_types.UUID) error {
	msg := notImplementedMsg
	return c.JSON(http.StatusNotImplemented, openapi.ErrorResponse{Message: &msg})
}

func (h OrganizationHandlers) GetOrganizationsIdHierarchy(c echo.Context, _ openapi_types.UUID) error {
	msg := notImplementedMsg
	return c.JSON(http.StatusNotImplemented, openapi.ErrorResponse{Message: &msg})
}
