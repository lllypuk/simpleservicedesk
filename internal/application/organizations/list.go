package organizations

import (
	"net/http"

	"simpleservicedesk/generated/openapi"
	"simpleservicedesk/internal/queries"

	"github.com/labstack/echo/v4"
)

func (h OrganizationHandlers) GetOrganizations(c echo.Context, params openapi.GetOrganizationsParams) error {
	ctx := c.Request().Context()

	// Convert OpenAPI params to filter using the centralized converter
	filter, err := queries.FromOpenAPIOrganizationParams(params)
	if err != nil {
		msg := err.Error()
		return c.JSON(http.StatusBadRequest, openapi.ErrorResponse{Message: &msg})
	}

	// Validate filter with business rules
	filter, validateErr := filter.ValidateAndSetDefaults()
	if validateErr != nil {
		msg := validateErr.Error()
		return c.JSON(http.StatusBadRequest, openapi.ErrorResponse{Message: &msg})
	}

	orgs, err := h.repo.ListOrganizations(ctx, filter)
	if err != nil {
		msg := err.Error()
		return c.JSON(http.StatusInternalServerError, openapi.ErrorResponse{Message: &msg})
	}

	total, err := h.repo.CountOrganizations(ctx, filter)
	if err != nil {
		msg := err.Error()
		return c.JSON(http.StatusInternalServerError, openapi.ErrorResponse{Message: &msg})
	}

	var orgResponses []openapi.GetOrganizationResponse
	for _, org := range orgs {
		response := h.buildOrganizationResponse(org)
		orgResponses = append(orgResponses, response)
	}

	page := 1
	if params.Page != nil {
		page = *params.Page
	}

	limit := filter.Limit
	totalInt := int(total)
	totalPages := 0
	if limit > 0 {
		totalPages = (totalInt + limit - 1) / limit
	}
	hasNext := page < totalPages

	pagination := openapi.PaginationResponse{
		Total:   &totalInt,
		Page:    &page,
		Limit:   &limit,
		HasNext: &hasNext,
	}

	return c.JSON(http.StatusOK, openapi.ListOrganizationsResponse{
		Organizations: &orgResponses,
		Pagination:    &pagination,
	})
}
