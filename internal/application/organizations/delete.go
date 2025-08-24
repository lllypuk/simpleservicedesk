package organizations

import (
	"errors"
	"net/http"

	"simpleservicedesk/generated/openapi"
	"simpleservicedesk/internal/domain/organizations"

	"github.com/labstack/echo/v4"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

func (h OrganizationHandlers) DeleteOrganizationsID(c echo.Context, id openapi_types.UUID) error {
	ctx := c.Request().Context()

	err := h.repo.DeleteOrganization(ctx, id)
	if err != nil {
		msg := err.Error()
		if errors.Is(err, organizations.ErrOrganizationNotFound) {
			return c.JSON(http.StatusNotFound, openapi.ErrorResponse{Message: &msg})
		}
		return c.JSON(http.StatusInternalServerError, openapi.ErrorResponse{Message: &msg})
	}

	return c.NoContent(http.StatusNoContent)
}
