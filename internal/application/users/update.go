package users

import (
	"net/http"

	"simpleservicedesk/generated/openapi"
	"simpleservicedesk/internal/domain/users"

	"github.com/labstack/echo/v4"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

func (h UserHandlers) PutUsersId(c echo.Context, id openapi_types.UUID) error {
	ctx := c.Request().Context()

	var req openapi.UpdateUserRequest
	if err := c.Bind(&req); err != nil {
		return err
	}

	user, err := h.repo.UpdateUser(ctx, id, func(user *users.User) (bool, error) {
		return h.applyUserUpdates(&req, user)
	})
	if err != nil {
		return handleUserError(c, err)
	}

	response := userToResponse(user)
	return c.JSON(http.StatusOK, response)
}

func (h UserHandlers) applyUserUpdates(req *openapi.UpdateUserRequest, user *users.User) (bool, error) {
	var hasChanges bool

	// Update name if provided
	if req.Name != nil && *req.Name != user.Name() {
		if err := user.ChangeName(*req.Name); err != nil {
			return false, err
		}
		hasChanges = true
	}

	// Update email if provided
	if req.Email != nil && string(*req.Email) != user.Email() {
		if err := user.ChangeEmail(string(*req.Email)); err != nil {
			return false, err
		}
		hasChanges = true
	}

	// Update organization if provided
	if req.OrganizationId != nil {
		currentOrgID := user.OrganizationID()
		if currentOrgID == nil || *currentOrgID != *req.OrganizationId {
			if err := user.ChangeOrganization(req.OrganizationId); err != nil {
				return false, err
			}
			hasChanges = true
		}
	}

	// Update active status if provided
	if req.IsActive != nil && *req.IsActive != user.IsActive() {
		if *req.IsActive {
			user.Activate()
		} else {
			user.Deactivate()
		}
		hasChanges = true
	}

	return hasChanges, nil
}
