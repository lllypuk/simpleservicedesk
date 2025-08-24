package categories

import (
	"errors"
	"net/http"
	"strings"

	"simpleservicedesk/generated/openapi"
	"simpleservicedesk/internal/domain/categories"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

func (h CategoryHandlers) categoryToResponse(category *categories.Category) openapi.GetCategoryResponse {
	var parentID *openapi_types.UUID
	if category.ParentID() != nil {
		pid := *category.ParentID()
		parentID = &pid
	}

	var description *string
	if category.Description() != "" {
		desc := category.Description()
		description = &desc
	}

	// Convert all fields to pointers as required by OpenAPI types
	categoryID := category.ID()
	name := category.Name()
	organizationID := category.OrganizationID()
	isActive := category.IsActive()
	createdAt := category.CreatedAt()
	updatedAt := category.UpdatedAt()

	return openapi.GetCategoryResponse{
		Id:             &categoryID,
		Name:           &name,
		Description:    description,
		OrganizationId: &organizationID,
		ParentId:       parentID,
		IsActive:       &isActive,
		CreatedAt:      &createdAt,
		UpdatedAt:      &updatedAt,
	}
}

func (h CategoryHandlers) handleCategoryError(c echo.Context, err error) error {
	switch {
	case errors.Is(err, categories.ErrCategoryNotFound):
		msg := "category not found"
		return c.JSON(http.StatusNotFound, openapi.ErrorResponse{Message: &msg})
	case errors.Is(err, categories.ErrCategoryAlreadyExist):
		msg := "category already exists"
		return c.JSON(http.StatusConflict, openapi.ErrorResponse{Message: &msg})
	case errors.Is(err, categories.ErrCategoryValidation):
		msg := err.Error()
		return c.JSON(http.StatusBadRequest, openapi.ErrorResponse{Message: &msg})
	case errors.Is(err, categories.ErrCircularReference):
		msg := "circular reference detected"
		return c.JSON(http.StatusBadRequest, openapi.ErrorResponse{Message: &msg})
	default:
		// Check if error message contains validation keywords
		if err != nil && (strings.Contains(err.Error(), "validation") ||
			strings.Contains(err.Error(), "required") ||
			strings.Contains(err.Error(), "invalid") ||
			strings.Contains(err.Error(), "must be")) {
			msg := err.Error()
			return c.JSON(http.StatusBadRequest, openapi.ErrorResponse{Message: &msg})
		}
		msg := "internal server error"
		return c.JSON(http.StatusInternalServerError, openapi.ErrorResponse{Message: &msg})
	}
}

func stringValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func (h CategoryHandlers) applyCategoryUpdates(
	req *openapi.UpdateCategoryRequest,
	cat *categories.Category,
) (bool, error) {
	var hasChanges bool

	// Update name if provided
	if changed, err := h.updateCategoryName(req.Name, cat); err != nil {
		return false, err
	} else if changed {
		hasChanges = true
	}

	// Update description if provided
	if changed, err := h.updateCategoryDescription(req.Description, cat); err != nil {
		return false, err
	} else if changed {
		hasChanges = true
	}

	// Update parent if provided
	if changed, err := h.updateCategoryParent(req.ParentId, cat); err != nil {
		return false, err
	} else if changed {
		hasChanges = true
	}

	// Update active status if provided
	if changed := h.updateCategoryActiveStatus(req.IsActive, cat); changed {
		hasChanges = true
	}

	return hasChanges, nil
}

func (h CategoryHandlers) updateCategoryName(name *string, cat *categories.Category) (bool, error) {
	if name != nil && *name != cat.Name() {
		if err := cat.ChangeName(*name); err != nil {
			return false, err
		}
		return true, nil
	}
	return false, nil
}

func (h CategoryHandlers) updateCategoryDescription(description *string, cat *categories.Category) (bool, error) {
	if description != nil && *description != cat.Description() {
		if err := cat.ChangeDescription(*description); err != nil {
			return false, err
		}
		return true, nil
	}
	return false, nil
}

func (h CategoryHandlers) updateCategoryParent(parentID *openapi_types.UUID, cat *categories.Category) (bool, error) {
	if parentID == nil {
		return false, nil
	}

	currentParent := cat.ParentID()
	if !h.parentNeedsUpdate(currentParent, parentID) {
		return false, nil
	}

	pid := *parentID
	newParentID := &pid

	if err := cat.ChangeParent(newParentID); err != nil {
		return false, err
	}
	return true, nil
}

func (h CategoryHandlers) parentNeedsUpdate(currentParent *uuid.UUID, newParentID *openapi_types.UUID) bool {
	if currentParent == nil && newParentID != nil {
		return true
	}
	if currentParent != nil && newParentID == nil {
		return true
	}
	if currentParent != nil && newParentID != nil && *currentParent != *newParentID {
		return true
	}
	return false
}

func (h CategoryHandlers) updateCategoryActiveStatus(isActive *bool, cat *categories.Category) bool {
	if isActive != nil && *isActive != cat.IsActive() {
		if *isActive {
			cat.Activate()
		} else {
			cat.Deactivate()
		}
		return true
	}
	return false
}
