package tickets

import (
	"net/http"

	"simpleservicedesk/generated/openapi"
	"simpleservicedesk/internal/domain/tickets"
	"simpleservicedesk/internal/queries"

	"github.com/labstack/echo/v4"
)

func (h TicketHandlers) GetTickets(c echo.Context, params openapi.GetTicketsParams) error {
	ctx := c.Request().Context()

	// Convert OpenAPI params to filter using the centralized converter
	filter, err := queries.FromOpenAPITicketParams(params)
	if err != nil {
		msg := err.Error()
		return c.JSON(http.StatusBadRequest, openapi.ErrorResponse{Message: &msg})
	}

	// Validate filter with business rules
	filter, validateErr := filter.ValidateAndSetDefaults()
	if validateErr != nil {
		msg := validateErr.Error()
		return c.JSON(http.StatusBadRequest, openapi.ErrorResponse{Message: &msg})
	}

	// Get tickets from repository
	ticketList, err := h.repo.ListTickets(ctx, filter)
	if err != nil {
		msg := err.Error()
		return c.JSON(http.StatusInternalServerError, openapi.ErrorResponse{Message: &msg})
	}

	// Build response
	page := 1
	if params.Page != nil {
		page = *params.Page
	}
	response := h.buildListResponse(ticketList, filter.Limit, page)
	return c.JSON(http.StatusOK, response)
}

func (h TicketHandlers) buildListResponse(
	ticketList []*tickets.Ticket,
	limit int,
	page int,
) openapi.ListTicketsResponse {
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
