package users

import (
	"net/http"

	"simpleservicedesk/generated/openapi"
	"simpleservicedesk/internal/application/services"

	"github.com/labstack/echo/v4"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

func (h UserHandlers) PutUsersID(c echo.Context, id openapi_types.UUID) error {
	ctx := c.Request().Context()

	var req openapi.UpdateUserRequest
	if err := c.Bind(&req); err != nil {
		return err
	}

	// Convert OpenAPI request to service request
	serviceReq := services.UpdateUserRequest{}

	if req.Name != nil {
		serviceReq.Name = req.Name
	}

	if req.Email != nil {
		email := string(*req.Email)
		serviceReq.Email = &email
	}

	if req.IsActive != nil {
		serviceReq.IsActive = req.IsActive
	}

	// Note: Password update is not included in this endpoint
	// It should be handled by a separate endpoint for security reasons

	user, err := h.userService.UpdateUser(ctx, id, serviceReq)
	if err != nil {
		return handleUserError(c, err)
	}

	response := userToResponse(user)
	return c.JSON(http.StatusOK, response)
}
