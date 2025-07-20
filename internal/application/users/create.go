package users

import (
	"errors"
	"net/http"

	"simpleservicedesk/generated/openapi"
	domain "simpleservicedesk/internal/domain/users"

	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"
)

func (h UserHandlers) PostUsers(c echo.Context) error {
	ctx := c.Request().Context()
	var req openapi.CreateUserRequest
	if err := c.Bind(&req); err != nil {
		return err
	}

	// Валидация пароля на уровне хэндлера
	if req.Password == "" {
		msg := "password is required"
		return c.JSON(http.StatusBadRequest, openapi.ErrorResponse{Message: &msg})
	}
	if len(req.Password) < 6 {
		msg := "password must be at least 6 characters long"
		return c.JSON(http.StatusBadRequest, openapi.ErrorResponse{Message: &msg})
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		msg := "failed to process password"
		return c.JSON(http.StatusInternalServerError, openapi.ErrorResponse{Message: &msg})
	}

	email := string(req.Email)
	user, err := h.repo.CreateUser(ctx, email, passwordHash, func() (*domain.User, error) {
		return domain.CreateUser(req.Name, email, passwordHash)
	})
	if err != nil {
		msg := err.Error()
		if errors.Is(err, domain.ErrInvalidUser) || errors.Is(err, domain.ErrUserValidation) {
			return c.JSON(http.StatusBadRequest, openapi.ErrorResponse{Message: &msg})
		}
		if errors.Is(err, domain.ErrUserAlreadyExist) {
			return c.JSON(http.StatusConflict, openapi.ErrorResponse{Message: &msg})
		}
		return c.JSON(http.StatusInternalServerError, openapi.ErrorResponse{Message: &msg})
	}

	id := user.ID()
	return c.JSON(http.StatusCreated, openapi.CreateUserResponse{Id: &id})
}
