package auth

import (
	"errors"
	"net/http"

	"simpleservicedesk/generated/openapi"

	"github.com/labstack/echo/v4"
)

func (h Handlers) PostLogin(c echo.Context) error {
	if h.service == nil {
		msg := "internal server error"
		return c.JSON(http.StatusInternalServerError, openapi.ErrorResponse{Message: &msg})
	}

	var req openapi.LoginRequest
	if err := c.Bind(&req); err != nil {
		return err
	}

	token, err := h.service.Login(c.Request().Context(), string(req.Email), req.Password)
	if err != nil {
		if errors.Is(err, ErrInvalidCredentials) {
			msg := "invalid credentials"
			return c.JSON(http.StatusUnauthorized, openapi.ErrorResponse{Message: &msg})
		}

		msg := "internal server error"
		return c.JSON(http.StatusInternalServerError, openapi.ErrorResponse{Message: &msg})
	}

	return c.JSON(http.StatusOK, openapi.LoginResponse{Token: token})
}
