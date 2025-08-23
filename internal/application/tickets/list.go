package tickets

import (
	"errors"
	"fmt"
	"net/http"

	"simpleservicedesk/generated/openapi"
	"simpleservicedesk/internal/domain/tickets"

	"github.com/labstack/echo/v4"
)

func (h TicketHandlers) GetTickets(c echo.Context, params openapi.GetTicketsParams) error {
	ctx := c.Request().Context()

	// Build filter from query parameters
	filter, err := h.buildTicketFilter(params)
	if err != nil {
		msg := err.Error()
		return c.JSON(http.StatusBadRequest, openapi.ErrorResponse{Message: &msg})
	}

	// Parse and validate pagination
	page, limit, err := h.parsePagination(params)
	if err != nil {
		errMsg := err.Error()
		return c.JSON(http.StatusBadRequest, openapi.ErrorResponse{Message: &errMsg})
	}

	filter.Limit = limit
	filter.Offset = (page - 1) * limit

	// Get tickets from repository
	ticketList, err := h.repo.ListTickets(ctx, filter)
	if err != nil {
		msg := err.Error()
		return c.JSON(http.StatusInternalServerError, openapi.ErrorResponse{Message: &msg})
	}

	// Build response
	response := h.buildListResponse(ticketList, page, limit)
	return c.JSON(http.StatusOK, response)
}

func (h TicketHandlers) buildTicketFilter(params openapi.GetTicketsParams) (TicketFilter, error) {
	filter := TicketFilter{}

	// Parse status filter
	if params.Status != nil {
		status, err := tickets.ParseStatus(string(*params.Status))
		if err != nil {
			return filter, fmt.Errorf("invalid status: %w", err)
		}
		filter.Status = &status
	}

	// Parse priority filter
	if params.Priority != nil {
		priority, err := tickets.ParsePriority(string(*params.Priority))
		if err != nil {
			return filter, fmt.Errorf("invalid priority: %w", err)
		}
		filter.Priority = &priority
	}

	// Convert UUID filters
	if params.CategoryId != nil {
		categoryID := *params.CategoryId
		filter.CategoryID = &categoryID
	}

	if params.AssigneeId != nil {
		assigneeID := *params.AssigneeId
		filter.AssigneeID = &assigneeID
	}

	if params.OrganizationId != nil {
		organizationID := *params.OrganizationId
		filter.OrganizationID = &organizationID
	}

	if params.AuthorId != nil {
		authorID := *params.AuthorId
		filter.AuthorID = &authorID
	}

	return filter, nil
}

func (h TicketHandlers) parsePagination(params openapi.GetTicketsParams) (int, int, error) {
	page := 1
	limit := 20

	if params.Page != nil {
		page = *params.Page
	}
	if params.Limit != nil {
		limit = *params.Limit
	}

	// Validate pagination
	if page < 1 {
		return 0, 0, errors.New("page must be greater than 0")
	}
	if limit < 1 || limit > 100 {
		return 0, 0, errors.New("limit must be between 1 and 100")
	}

	return page, limit, nil
}

func (h TicketHandlers) buildListResponse(ticketList []*tickets.Ticket, page, limit int) openapi.ListTicketsResponse {
	// Convert domain tickets to OpenAPI responses
	ticketResponses := make([]openapi.GetTicketResponse, len(ticketList))
	for i, ticket := range ticketList {
		ticketResponses[i] = convertTicketToResponse(ticket)
	}

	// Build pagination response
	total := len(ticketResponses) // This is simplified - in real implementation, you'd need total count from repo
	hasNext := len(ticketResponses) == limit

	pagination := openapi.PaginationResponse{
		Total:   &total,
		Page:    &page,
		Limit:   &limit,
		HasNext: &hasNext,
	}

	return openapi.ListTicketsResponse{
		Tickets:    &ticketResponses,
		Pagination: &pagination,
	}
}

// convertTicketToResponse converts domain ticket to OpenAPI response
func convertTicketToResponse(ticket *tickets.Ticket) openapi.GetTicketResponse {
	id := ticket.ID()
	title := ticket.Title()
	description := ticket.Description()
	status := openapi.TicketStatus(ticket.Status().String())
	priority := openapi.TicketPriority(ticket.Priority().String())
	organizationID := ticket.OrganizationID()
	authorID := ticket.AuthorID()
	createdAt := ticket.CreatedAt()
	updatedAt := ticket.UpdatedAt()

	response := openapi.GetTicketResponse{
		Id:             &id,
		Title:          &title,
		Description:    &description,
		Status:         &status,
		Priority:       &priority,
		OrganizationId: &organizationID,
		AuthorId:       &authorID,
		CreatedAt:      &createdAt,
		UpdatedAt:      &updatedAt,
	}

	// Add optional fields if they exist
	if categoryID := ticket.CategoryID(); categoryID != nil {
		response.CategoryId = categoryID
	}

	if assigneeID := ticket.AssigneeID(); assigneeID != nil {
		response.AssigneeId = assigneeID
	}

	if resolvedAt := ticket.ResolvedAt(); resolvedAt != nil {
		response.ResolvedAt = resolvedAt
	}

	if closedAt := ticket.ClosedAt(); closedAt != nil {
		response.ClosedAt = closedAt
	}

	return response
}
