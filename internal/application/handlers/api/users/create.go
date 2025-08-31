package users

import (
	"errors"
	"net/http"

	"simpleservicedesk/generated/openapi"
	"simpleservicedesk/internal/application/services"
	domain "simpleservicedesk/internal/domain/users"

	"github.com/labstack/echo/v4"
)

func (h UserHandlers) PostUsers(c echo.Context) error {
	ctx := c.Request().Context()
	var req openapi.CreateUserRequest
	if err := c.Bind(&req); err != nil {
		return err
	}

	// Create user using service
	user, err := h.userService.CreateUser(ctx, services.CreateUserRequest{
		Name:     req.Name,
		Email:    string(req.Email),
		Password: req.Password,
	})
	if err != nil {
		msg := err.Error()
		if errors.Is(err, domain.ErrInvalidUser) ||
			errors.Is(err, domain.ErrUserValidation) ||
			err.Error() == "password is required" ||
			err.Error() == "password must be at least 6 characters long" {
			return c.JSON(http.StatusBadRequest, openapi.ErrorResponse{Message: &msg})
		}
		if errors.Is(err, domain.ErrUserAlreadyExist) {
			return c.JSON(http.StatusConflict, openapi.ErrorResponse{Message: &msg})
		}
		if err.Error() == "failed to process password" {
			return c.JSON(http.StatusInternalServerError, openapi.ErrorResponse{Message: &msg})
		}
		return c.JSON(http.StatusInternalServerError, openapi.ErrorResponse{Message: &msg})
	}

	id := user.ID()
	return c.JSON(http.StatusCreated, openapi.CreateUserResponse{Id: &id})
}
