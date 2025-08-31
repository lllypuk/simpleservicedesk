package categories

import (
	"errors"
	"net/http"
	"strings"

	"simpleservicedesk/generated/openapi"
	"simpleservicedesk/internal/domain/categories"

	"github.com/labstack/echo/v4"

	"simpleservicedesk/internal/application/handlers/api"
)

func (h CategoryHandlers) handleCategoryError(c echo.Context, err error) error {
	switch {
	case errors.Is(err, categories.ErrCategoryNotFound):
		msg := api.MsgCategoryNotFound
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
