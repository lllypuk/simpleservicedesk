package tickets

import (
	"errors"
	"net/http"

	"simpleservicedesk/generated/openapi"
	"simpleservicedesk/internal/domain/tickets"

	"github.com/labstack/echo/v4"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

func (h TicketHandlers) PatchTicketsIdAssign(c echo.Context, id openapi_types.UUID) error {
	ctx := c.Request().Context()
	var req openapi.AssignTicketRequest
	if err := c.Bind(&req); err != nil {
		return err
	}

	ticket, err := h.repo.UpdateTicket(ctx, id, func(ticket *tickets.Ticket) (bool, error) {
		if req.AssigneeId == nil {
			// Unassign ticket
			ticket.Unassign()
		} else {
			// Convert and assign to user
			assigneeID := *req.AssigneeId
			if err := ticket.AssignTo(assigneeID); err != nil {
				return false, err
			}
		}
		return true, nil
	})

	if err != nil {
		msg := err.Error()
		if errors.Is(err, tickets.ErrTicketNotFound) {
			return c.JSON(http.StatusNotFound, openapi.ErrorResponse{Message: &msg})
		}
		if errors.Is(err, tickets.ErrTicketValidation) {
			return c.JSON(http.StatusBadRequest, openapi.ErrorResponse{Message: &msg})
		}
		return c.JSON(http.StatusInternalServerError, openapi.ErrorResponse{Message: &msg})
	}

	response := convertTicketToResponse(ticket)
	return c.JSON(http.StatusOK, response)
}
