package organizations

import (
	"net/http"

	"simpleservicedesk/generated/openapi"

	"github.com/labstack/echo/v4"
)

func (h OrganizationHandlers) GetOrganizations(c echo.Context, params openapi.GetOrganizationsParams) error {
	ctx := c.Request().Context()

	filter := OrganizationFilter{
		Limit:  DefaultPageLimit,
		Offset: 0,
	}

	if params.Page != nil && *params.Page > 0 {
		filter.Offset = (*params.Page - 1) * filter.Limit
	}
	if params.Limit != nil && *params.Limit > 0 && *params.Limit <= 100 {
		filter.Limit = *params.Limit
	}

	orgs, err := h.repo.ListOrganizations(ctx, filter)
	if err != nil {
		msg := err.Error()
		return c.JSON(http.StatusInternalServerError, openapi.ErrorResponse{Message: &msg})
	}

	var orgResponses []openapi.GetOrganizationResponse
	for _, org := range orgs {
		response := h.buildOrganizationResponse(org)
		orgResponses = append(orgResponses, response)
	}

	// TODO: Add proper pagination response
	return c.JSON(http.StatusOK, openapi.ListOrganizationsResponse{
		Organizations: &orgResponses,
	})
}
