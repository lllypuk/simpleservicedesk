package tickets

import (
	"errors"
	"net/http"

	"simpleservicedesk/generated/openapi"
	"simpleservicedesk/internal/domain/tickets"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func (h TicketHandlers) PostTickets(c echo.Context) error {
	ctx := c.Request().Context()
	var req openapi.CreateTicketRequest
	if err := c.Bind(&req); err != nil {
		return err
	}

	// Parse and validate priority
	priority, err := tickets.ParsePriority(string(req.Priority))
	if err != nil {
		msg := "invalid priority: " + err.Error()
		return c.JSON(http.StatusBadRequest, openapi.ErrorResponse{Message: &msg})
	}

	// Convert OpenAPI types to uuid.UUID
	organizationID := req.OrganizationId
	authorID := req.AuthorId

	// Convert optional category ID
	var categoryID *uuid.UUID
	if req.CategoryId != nil {
		cid := *req.CategoryId
		categoryID = &cid
	}

	ticket, err := h.repo.CreateTicket(ctx, func() (*tickets.Ticket, error) {
		return tickets.NewTicket(
			uuid.New(),
			req.Title,
			req.Description,
			priority,
			organizationID,
			authorID,
			categoryID,
		)
	})
	if err != nil {
		msg := err.Error()
		if errors.Is(err, tickets.ErrTicketValidation) || errors.Is(err, tickets.ErrInvalidTicket) {
			return c.JSON(http.StatusBadRequest, openapi.ErrorResponse{Message: &msg})
		}
		return c.JSON(http.StatusInternalServerError, openapi.ErrorResponse{Message: &msg})
	}

	response := convertTicketToResponse(ticket)
	return c.JSON(http.StatusCreated, response)
}
