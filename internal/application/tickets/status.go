package tickets

import (
	"errors"
	"net/http"

	"simpleservicedesk/generated/openapi"
	"simpleservicedesk/internal/domain/tickets"

	"github.com/labstack/echo/v4"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

func (h TicketHandlers) PatchTicketsIDStatus(c echo.Context, id openapi_types.UUID) error {
	ctx := c.Request().Context()
	var req openapi.UpdateTicketStatusRequest
	if err := c.Bind(&req); err != nil {
		return err
	}

	// Parse and validate status
	newStatus, err := tickets.ParseStatus(string(req.Status))
	if err != nil {
		msg := "invalid status: " + err.Error()
		return c.JSON(http.StatusBadRequest, openapi.ErrorResponse{Message: &msg})
	}

	ticket, err := h.repo.UpdateTicket(ctx, id, func(ticket *tickets.Ticket) (bool, error) {
		if statusErr := ticket.ChangeStatus(newStatus); statusErr != nil {
			return false, statusErr
		}
		return true, nil
	})

	if err != nil {
		msg := err.Error()
		if errors.Is(err, tickets.ErrTicketNotFound) {
			return c.JSON(http.StatusNotFound, openapi.ErrorResponse{Message: &msg})
		}
		if errors.Is(err, tickets.ErrInvalidTransition) {
			return c.JSON(http.StatusBadRequest, openapi.ErrorResponse{Message: &msg})
		}
		return c.JSON(http.StatusInternalServerError, openapi.ErrorResponse{Message: &msg})
	}

	response := convertTicketToResponse(ticket)
	return c.JSON(http.StatusOK, response)
}
