package organizations

import (
	"errors"
	"net/http"

	"simpleservicedesk/generated/openapi"
	"simpleservicedesk/internal/domain/organizations"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func (h OrganizationHandlers) PostOrganizations(c echo.Context) error {
	ctx := c.Request().Context()
	var req openapi.CreateOrganizationRequest
	if err := c.Bind(&req); err != nil {
		return err
	}

	// Validate required fields
	if req.Name == "" {
		msg := "name is required"
		return c.JSON(http.StatusBadRequest, openapi.ErrorResponse{Message: &msg})
	}
	if req.Domain == nil || *req.Domain == "" {
		msg := "domain is required"
		return c.JSON(http.StatusBadRequest, openapi.ErrorResponse{Message: &msg})
	}

	var parentID *uuid.UUID
	if req.ParentId != nil {
		parentID = req.ParentId
	}

	var org *organizations.Organization
	var err error

	if parentID == nil {
		org, err = h.repo.CreateOrganization(ctx, func() (*organizations.Organization, error) {
			return organizations.CreateRootOrganization(req.Name, *req.Domain)
		})
	} else {
		org, err = h.repo.CreateOrganization(ctx, func() (*organizations.Organization, error) {
			return organizations.CreateSubOrganization(req.Name, *req.Domain, *parentID)
		})
	}

	if err != nil {
		msg := err.Error()
		if errors.Is(err, organizations.ErrInvalidOrganization) ||
			errors.Is(err, organizations.ErrOrganizationValidation) {
			return c.JSON(http.StatusBadRequest, openapi.ErrorResponse{Message: &msg})
		}
		if errors.Is(err, organizations.ErrOrganizationAlreadyExist) {
			return c.JSON(http.StatusConflict, openapi.ErrorResponse{Message: &msg})
		}
		return c.JSON(http.StatusInternalServerError, openapi.ErrorResponse{Message: &msg})
	}

	id := org.ID()

	return c.JSON(http.StatusCreated, openapi.CreateOrganizationResponse{
		Id: &id,
	})
}
