package tickets

import (
	"errors"
	"net/http"

	"simpleservicedesk/generated/openapi"
	"simpleservicedesk/internal/domain/tickets"

	"github.com/labstack/echo/v4"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

func (h TicketHandlers) PutTicketsId(c echo.Context, id openapi_types.UUID) error {
	ctx := c.Request().Context()
	var req openapi.UpdateTicketRequest
	if err := c.Bind(&req); err != nil {
		return err
	}

	ticket, err := h.repo.UpdateTicket(ctx, id, func(ticket *tickets.Ticket) (bool, error) {
		return h.applyTicketUpdates(ticket, req)
	})

	if err != nil {
		return h.handleUpdateError(c, err)
	}

	response := convertTicketToResponse(ticket)
	return c.JSON(http.StatusOK, response)
}

func (h TicketHandlers) applyTicketUpdates(ticket *tickets.Ticket, req openapi.UpdateTicketRequest) (bool, error) {
	updated := false

	// Update title if provided
	if req.Title != nil {
		if err := ticket.UpdateTitle(*req.Title); err != nil {
			return false, err
		}
		updated = true
	}

	// Update description if provided
	if req.Description != nil {
		if err := ticket.UpdateDescription(*req.Description); err != nil {
			return false, err
		}
		updated = true
	}

	// Update priority if provided
	if req.Priority != nil {
		if err := h.updateTicketPriority(ticket, *req.Priority); err != nil {
			return false, err
		}
		updated = true
	}

	// Update category if provided
	if req.CategoryId != nil {
		categoryID := *req.CategoryId
		ticket.SetCategory(&categoryID)
		updated = true
	}

	return updated, nil
}

func (h TicketHandlers) updateTicketPriority(ticket *tickets.Ticket, priorityStr openapi.TicketPriority) error {
	priority, err := tickets.ParsePriority(string(priorityStr))
	if err != nil {
		return err
	}
	return ticket.UpdatePriority(priority)
}

func (h TicketHandlers) handleUpdateError(c echo.Context, err error) error {
	msg := err.Error()
	if errors.Is(err, tickets.ErrTicketNotFound) {
		return c.JSON(http.StatusNotFound, openapi.ErrorResponse{Message: &msg})
	}
	if errors.Is(err, tickets.ErrTicketValidation) || errors.Is(err, tickets.ErrInvalidTicket) {
		return c.JSON(http.StatusBadRequest, openapi.ErrorResponse{Message: &msg})
	}
	return c.JSON(http.StatusInternalServerError, openapi.ErrorResponse{Message: &msg})
}
