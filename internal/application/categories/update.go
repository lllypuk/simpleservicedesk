package categories

import (
	"net/http"

	"simpleservicedesk/generated/openapi"
	"simpleservicedesk/internal/domain/categories"

	"github.com/labstack/echo/v4"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

func (h CategoryHandlers) PutCategoriesId(c echo.Context, id openapi_types.UUID) error {
	ctx := c.Request().Context()
	var req openapi.UpdateCategoryRequest
	if err := c.Bind(&req); err != nil {
		return err
	}

	category, err := h.repo.UpdateCategory(ctx, id, func(cat *categories.Category) (bool, error) {
		return h.applyCategoryUpdates(&req, cat)
	})
	if err != nil {
		return h.handleCategoryError(c, err)
	}

	response := h.categoryToResponse(category)
	return c.JSON(http.StatusOK, response)
}
