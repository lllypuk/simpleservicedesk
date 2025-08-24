package users

import (
	"errors"
	"net/http"

	"simpleservicedesk/generated/openapi"
	"simpleservicedesk/internal/domain/users"

	"github.com/labstack/echo/v4"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

const (
	userNotFoundMessage = "user not found"
)

func userToResponse(user *users.User) openapi.GetUserResponse {
	id := user.ID()
	name := user.Name()
	email := openapi_types.Email(user.Email())
	role := openapi.UserRole(user.Role().String())
	isActive := user.IsActive()
	createdAt := user.CreatedAt()
	updatedAt := user.UpdatedAt()

	response := openapi.GetUserResponse{
		Id:        &id,
		Name:      &name,
		Email:     &email,
		Role:      &role,
		IsActive:  &isActive,
		CreatedAt: &createdAt,
		UpdatedAt: &updatedAt,
	}

	if orgID := user.OrganizationID(); orgID != nil {
		response.OrganizationId = orgID
	}

	return response
}

func handleUserError(c echo.Context, err error) error {
	if errors.Is(err, users.ErrUserNotFound) {
		msg := userNotFoundMessage
		return c.JSON(http.StatusNotFound, openapi.ErrorResponse{Message: &msg})
	}
	if errors.Is(err, users.ErrUserAlreadyExist) {
		msg := "user already exists"
		return c.JSON(http.StatusConflict, openapi.ErrorResponse{Message: &msg})
	}
	if errors.Is(err, users.ErrUserValidation) || errors.Is(err, users.ErrInvalidRole) {
		msg := err.Error()
		return c.JSON(http.StatusBadRequest, openapi.ErrorResponse{Message: &msg})
	}

	msg := "internal server error"
	return c.JSON(http.StatusInternalServerError, openapi.ErrorResponse{Message: &msg})
}
