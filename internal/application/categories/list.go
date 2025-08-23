package categories

import (
	"net/http"

	"simpleservicedesk/generated/openapi"

	"github.com/labstack/echo/v4"
)

func (h CategoryHandlers) GetCategories(c echo.Context, params openapi.GetCategoriesParams) error {
	ctx := c.Request().Context()

	// Build filter from query parameters
	filter := CategoryFilter{}

	if params.OrganizationId != nil {
		oid := *params.OrganizationId
		filter.OrganizationID = &oid
	}

	if params.ParentId != nil {
		pid := *params.ParentId
		filter.ParentID = &pid
	}

	if params.IsActive != nil {
		filter.IsActive = params.IsActive
	}

	categoriesList, err := h.repo.ListCategories(ctx, filter)
	if err != nil {
		return h.handleCategoryError(c, err)
	}

	// Convert to response format
	categoryResponses := make([]openapi.GetCategoryResponse, len(categoriesList))
	for i, category := range categoriesList {
		categoryResponses[i] = h.categoryToResponse(category)
	}

	response := openapi.ListCategoriesResponse{
		Categories: &categoryResponses,
	}
	return c.JSON(http.StatusOK, response)
}
