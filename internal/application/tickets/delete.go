package tickets

import (
	"errors"
	"net/http"

	"simpleservicedesk/generated/openapi"
	"simpleservicedesk/internal/domain/tickets"

	"github.com/labstack/echo/v4"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

func (h TicketHandlers) DeleteTicketsId(c echo.Context, id openapi_types.UUID) error {
	ctx := c.Request().Context()

	err := h.repo.DeleteTicket(ctx, id)
	if err != nil {
		msg := err.Error()
		if errors.Is(err, tickets.ErrTicketNotFound) {
			return c.JSON(http.StatusNotFound, openapi.ErrorResponse{Message: &msg})
		}
		return c.JSON(http.StatusInternalServerError, openapi.ErrorResponse{Message: &msg})
	}

	return c.NoContent(http.StatusNoContent)
}
