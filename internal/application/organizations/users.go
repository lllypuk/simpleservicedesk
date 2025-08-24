package organizations

import (
	"net/http"

	"simpleservicedesk/generated/openapi"

	"github.com/labstack/echo/v4"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

func (h OrganizationHandlers) GetOrganizationsIDUsers(
	c echo.Context, _ openapi_types.UUID, _ openapi.GetOrganizationsIDUsersParams,
) error {
	// This endpoint returns users belonging to an organization
	// For now, return not implemented as it requires user repository integration
	msg := "organization users functionality not implemented yet"
	return c.JSON(http.StatusNotImplemented, openapi.ErrorResponse{Message: &msg})
}
