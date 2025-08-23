package categories

import (
	"net/http"

	"simpleservicedesk/generated/openapi"
	"simpleservicedesk/internal/domain/categories"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func (h CategoryHandlers) PostCategories(c echo.Context) error {
	ctx := c.Request().Context()
	var req openapi.CreateCategoryRequest
	if err := c.Bind(&req); err != nil {
		return err
	}

	// Convert OpenAPI types to uuid.UUID
	organizationID := req.OrganizationId

	// Convert optional parent ID
	var parentID *uuid.UUID
	if req.ParentId != nil {
		pid := *req.ParentId
		parentID = &pid
	}

	// Create the category
	category, err := h.repo.CreateCategory(ctx, func() (*categories.Category, error) {
		return categories.CreateCategory(
			req.Name,
			stringValue(req.Description),
			organizationID,
			parentID,
		)
	})
	if err != nil {
		return h.handleCategoryError(c, err)
	}

	categoryID := category.ID()
	response := openapi.CreateCategoryResponse{
		Id: &categoryID,
	}
	return c.JSON(http.StatusCreated, response)
}
