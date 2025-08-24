package categories

import (
	"net/http"

	"github.com/labstack/echo/v4"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

func (h CategoryHandlers) DeleteCategoriesId(c echo.Context, id openapi_types.UUID) error {
	ctx := c.Request().Context()

	err := h.repo.DeleteCategory(ctx, id)
	if err != nil {
		return h.handleCategoryError(c, err)
	}

	return c.NoContent(http.StatusNoContent)
}
