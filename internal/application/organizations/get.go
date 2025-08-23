package organizations

import (
	"errors"
	"net/http"

	"simpleservicedesk/generated/openapi"
	"simpleservicedesk/internal/domain/organizations"

	"github.com/labstack/echo/v4"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

func (h OrganizationHandlers) GetOrganizationsID(c echo.Context, id openapi_types.UUID) error {
	ctx := c.Request().Context()

	org, err := h.repo.GetOrganization(ctx, id)
	if err != nil {
		msg := err.Error()
		if errors.Is(err, organizations.ErrOrganizationNotFound) {
			return c.JSON(http.StatusNotFound, openapi.ErrorResponse{Message: &msg})
		}
		return c.JSON(http.StatusInternalServerError, openapi.ErrorResponse{Message: &msg})
	}

	return c.JSON(http.StatusOK, h.buildOrganizationResponse(org))
}

func (h OrganizationHandlers) buildOrganizationResponse(
	org *organizations.Organization,
) openapi.GetOrganizationResponse {
	orgID := org.ID()
	name := org.Name()
	domain := org.Domain()
	isActive := org.IsActive()
	createdAt := org.CreatedAt()
	updatedAt := org.UpdatedAt()

	response := openapi.GetOrganizationResponse{
		Id:        &orgID,
		Name:      &name,
		Domain:    &domain,
		IsActive:  &isActive,
		CreatedAt: &createdAt,
		UpdatedAt: &updatedAt,
	}

	if org.HasParent() {
		parentID := org.ParentID()
		response.ParentId = parentID
	}

	return response
}
