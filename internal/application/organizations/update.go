package organizations

import (
	"errors"
	"net/http"

	"simpleservicedesk/generated/openapi"
	"simpleservicedesk/internal/domain/organizations"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

func (h OrganizationHandlers) PutOrganizationsId(c echo.Context, id openapi_types.UUID) error {
	ctx := c.Request().Context()
	var req openapi.UpdateOrganizationRequest
	if err := c.Bind(&req); err != nil {
		return err
	}

	org, err := h.repo.UpdateOrganization(ctx, id, func(org *organizations.Organization) (bool, error) {
		return h.updateOrganizationFields(req, org)
	})

	if err != nil {
		return h.handleUpdateError(c, err)
	}

	return c.JSON(http.StatusOK, h.buildOrganizationResponse(org))
}

func (h OrganizationHandlers) updateOrganizationFields(
	req openapi.UpdateOrganizationRequest,
	org *organizations.Organization,
) (bool, error) {
	changed := false

	// Update name if provided
	if req.Name != nil && *req.Name != org.Name() {
		if err := org.ChangeName(*req.Name); err != nil {
			return false, err
		}
		changed = true
	}

	// Update domain if provided
	if req.Domain != nil && *req.Domain != org.Domain() {
		if err := org.ChangeDomain(*req.Domain); err != nil {
			return false, err
		}
		changed = true
	}

	// Update parent if provided
	if req.ParentId != nil {
		if h.shouldUpdateParent(req.ParentId, org.ParentID()) {
			if err := org.ChangeParent(req.ParentId); err != nil {
				return false, err
			}
			changed = true
		}
	}

	// Update active status if provided
	if req.IsActive != nil && *req.IsActive != org.IsActive() {
		if *req.IsActive {
			org.Activate()
		} else {
			org.Deactivate()
		}
		changed = true
	}

	return changed, nil
}

func (h OrganizationHandlers) shouldUpdateParent(newParent *uuid.UUID, currentParent *uuid.UUID) bool {
	if currentParent == nil && newParent != nil {
		return true
	}
	if currentParent != nil && newParent == nil {
		return true
	}
	if currentParent != nil && newParent != nil && *currentParent != *newParent {
		return true
	}
	return false
}

func (h OrganizationHandlers) handleUpdateError(c echo.Context, err error) error {
	msg := err.Error()
	if errors.Is(err, organizations.ErrOrganizationNotFound) {
		return c.JSON(http.StatusNotFound, openapi.ErrorResponse{Message: &msg})
	}
	if errors.Is(err, organizations.ErrInvalidOrganization) ||
		errors.Is(err, organizations.ErrOrganizationValidation) {
		return c.JSON(http.StatusBadRequest, openapi.ErrorResponse{Message: &msg})
	}
	if errors.Is(err, organizations.ErrCircularReference) {
		return c.JSON(http.StatusBadRequest, openapi.ErrorResponse{Message: &msg})
	}
	return c.JSON(http.StatusInternalServerError, openapi.ErrorResponse{Message: &msg})
}
