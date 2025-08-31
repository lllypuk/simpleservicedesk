package tickets

import (
	"errors"
	"net/http"

	"simpleservicedesk/generated/openapi"
	"simpleservicedesk/internal/application/handlers/api"
	"simpleservicedesk/internal/domain/tickets"

	"github.com/labstack/echo/v4"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

func (h TicketHandlers) GetTicketsID(c echo.Context, id openapi_types.UUID) error {
	ctx := c.Request().Context()

	ticket, err := h.ticketService.GetTicket(ctx, id)
	if err != nil {
		if errors.Is(err, tickets.ErrTicketNotFound) {
			msg := api.MsgTicketNotFound
			return c.JSON(http.StatusNotFound, openapi.ErrorResponse{Message: &msg})
		}
		msg := err.Error()
		return c.JSON(http.StatusInternalServerError, openapi.ErrorResponse{Message: &msg})
	}

	response := convertTicketToResponse(ticket)
	return c.JSON(http.StatusOK, response)
}
