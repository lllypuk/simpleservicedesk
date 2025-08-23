package categories

import (
	"net/http"

	"simpleservicedesk/generated/openapi"

	"github.com/labstack/echo/v4"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

const notImplementedMsg = "categories functionality not implemented yet"

type CategoryHandlers struct{}

func SetupHandlers() CategoryHandlers {
	return CategoryHandlers{}
}

func (h CategoryHandlers) GetCategories(c echo.Context, _ openapi.GetCategoriesParams) error {
	msg := notImplementedMsg
	return c.JSON(http.StatusNotImplemented, openapi.ErrorResponse{Message: &msg})
}

func (h CategoryHandlers) PostCategories(c echo.Context) error {
	msg := notImplementedMsg
	return c.JSON(http.StatusNotImplemented, openapi.ErrorResponse{Message: &msg})
}

func (h CategoryHandlers) GetCategoriesID(c echo.Context, _ openapi_types.UUID) error {
	msg := notImplementedMsg
	return c.JSON(http.StatusNotImplemented, openapi.ErrorResponse{Message: &msg})
}

func (h CategoryHandlers) GetCategoriesIdTickets(c echo.Context, _ openapi_types.UUID, _ openapi.GetCategoriesIdTicketsParams) error {
	msg := notImplementedMsg
	return c.JSON(http.StatusNotImplemented, openapi.ErrorResponse{Message: &msg})
}

func (h CategoryHandlers) PutCategoriesId(c echo.Context, _ openapi_types.UUID) error {
	msg := notImplementedMsg
	return c.JSON(http.StatusNotImplemented, openapi.ErrorResponse{Message: &msg})
}

func (h CategoryHandlers) DeleteCategoriesId(c echo.Context, _ openapi_types.UUID) error {
	msg := notImplementedMsg
	return c.JSON(http.StatusNotImplemented, openapi.ErrorResponse{Message: &msg})
}

func (h CategoryHandlers) GetCategoriesIdHierarchy(c echo.Context, _ openapi_types.UUID) error {
	msg := notImplementedMsg
	return c.JSON(http.StatusNotImplemented, openapi.ErrorResponse{Message: &msg})
}
