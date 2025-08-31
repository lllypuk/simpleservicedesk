package tickets

import (
	"errors"
	"net/http"

	"simpleservicedesk/generated/openapi"
	"simpleservicedesk/internal/application/services"
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

	// Convert optional category ID
	var categoryID *uuid.UUID
	if req.CategoryId != nil {
		cid := *req.CategoryId
		categoryID = &cid
	}

	// Create ticket using service
	ticket, err := h.ticketService.CreateTicket(ctx, req.AuthorId, services.CreateTicketRequest{
		Title:          req.Title,
		Description:    req.Description,
		Priority:       priority,
		CategoryID:     categoryID,
		OrganizationID: &req.OrganizationId,
		AssignedToID:   nil, // New tickets are not assigned by default
	})
	if err != nil {
		msg := err.Error()
		if errors.Is(err, tickets.ErrTicketValidation) ||
			errors.Is(err, tickets.ErrInvalidTicket) {
			return c.JSON(http.StatusBadRequest, openapi.ErrorResponse{Message: &msg})
		}
		return c.JSON(http.StatusInternalServerError, openapi.ErrorResponse{Message: &msg})
	}

	response := convertTicketToResponse(ticket)
	return c.JSON(http.StatusCreated, response)
}

func convertTicketToResponse(ticket *tickets.Ticket) openapi.GetTicketResponse {
	id := ticket.ID()
	title := ticket.Title()
	description := ticket.Description()
	status := openapi.TicketStatus(ticket.Status().String())
	priority := openapi.TicketPriority(ticket.Priority().String())
	createdAt := ticket.CreatedAt()
	updatedAt := ticket.UpdatedAt()
	authorID := ticket.AuthorID()

	response := openapi.GetTicketResponse{
		Id:          &id,
		Title:       &title,
		Description: &description,
		Status:      &status,
		Priority:    &priority,
		CreatedAt:   &createdAt,
		UpdatedAt:   &updatedAt,
		AuthorId:    &authorID,
	}

	orgID := ticket.OrganizationID()
	response.OrganizationId = &orgID

	if categoryID := ticket.CategoryID(); categoryID != nil {
		response.CategoryId = categoryID
	}

	if assignedTo := ticket.AssigneeID(); assignedTo != nil {
		response.AssigneeId = assignedTo
	}

	return response
}
