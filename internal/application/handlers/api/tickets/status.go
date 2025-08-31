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

func (h TicketHandlers) PatchTicketsIDStatus(c echo.Context, id openapi_types.UUID) error {
	ctx := c.Request().Context()

	var req openapi.UpdateTicketStatusRequest
	if err := c.Bind(&req); err != nil {
		msg := api.MsgInvalidRequestFormat
		return c.JSON(http.StatusBadRequest, openapi.ErrorResponse{Message: &msg})
	}

	// Parse and validate status
	status, err := tickets.ParseStatus(string(req.Status))
	if err != nil {
		msg := "invalid status: " + err.Error()
		return c.JSON(http.StatusBadRequest, openapi.ErrorResponse{Message: &msg})
	}

	// Update ticket status using service
	ticket, err := h.ticketService.UpdateTicketStatus(ctx, id, status)
	if err != nil {
		if errors.Is(err, tickets.ErrTicketNotFound) {
			msg := api.MsgTicketNotFound
			return c.JSON(http.StatusNotFound, openapi.ErrorResponse{Message: &msg})
		}
		if errors.Is(err, tickets.ErrInvalidTransition) {
			msg := err.Error()
			return c.JSON(http.StatusBadRequest, openapi.ErrorResponse{Message: &msg})
		}
		msg := err.Error()
		return c.JSON(http.StatusInternalServerError, openapi.ErrorResponse{Message: &msg})
	}

	response := convertTicketToResponse(ticket)
	return c.JSON(http.StatusOK, response)
}
