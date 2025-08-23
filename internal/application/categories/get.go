package categories

import (
	"net/http"

	"github.com/labstack/echo/v4"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

func (h CategoryHandlers) GetCategoriesID(c echo.Context, id openapi_types.UUID) error {
	ctx := c.Request().Context()

	category, err := h.repo.GetCategory(ctx, id)
	if err != nil {
		return h.handleCategoryError(c, err)
	}

	response := h.categoryToResponse(category)
	return c.JSON(http.StatusOK, response)
}
