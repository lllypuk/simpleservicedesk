package categories

import (
	"net/http"

	"simpleservicedesk/generated/openapi"
	"simpleservicedesk/internal/queries"

	"github.com/labstack/echo/v4"
)

func (h CategoryHandlers) GetCategories(c echo.Context, params openapi.GetCategoriesParams) error {
	ctx := c.Request().Context()

	// Convert OpenAPI params to filter using the centralized converter
	filter, err := queries.FromOpenAPICategoryParams(params)
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

	// Get categories from repository
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
