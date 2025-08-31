package users

import (
	"net/http"
	"strconv"

	"simpleservicedesk/generated/openapi"
	"simpleservicedesk/internal/domain/tickets"
	"simpleservicedesk/internal/queries"

	"github.com/labstack/echo/v4"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

func (h UserHandlers) GetUsersIDTickets(
	c echo.Context, userID openapi_types.UUID, params openapi.GetUsersIDTicketsParams,
) error {
	ctx := c.Request().Context()

	// Build filter from parameters
	filter := queries.TicketFilter{}

	// Apply pagination
	if params.Page != nil {
		filter.Offset = (*params.Page - 1) * filter.Limit
	}
	if params.Limit != nil {
		filter.Limit = *params.Limit
	}

	// Apply filters
	if params.Status != nil {
		status := tickets.Status(*params.Status)
		filter.Status = &status
	}

	if params.Priority != nil {
		priority := tickets.Priority(*params.Priority)
		filter.Priority = &priority
	}

	// Get user tickets using service
	tickets, total, err := h.userService.GetUserTickets(ctx, userID, filter)
	if err != nil {
		return handleUserError(c, err)
	}

	// Convert tickets to response format (simplified for now)
	items := make([]openapi.GetTicketResponse, len(tickets))
	for i, ticket := range tickets {
		id := ticket.ID()
		title := ticket.Title()
		description := ticket.Description()
		status := openapi.TicketStatus(ticket.Status().String())
		priority := openapi.TicketPriority(ticket.Priority().String())
		createdAt := ticket.CreatedAt()
		updatedAt := ticket.UpdatedAt()
		authorID := ticket.AuthorID()

		items[i] = openapi.GetTicketResponse{
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
		items[i].OrganizationId = &orgID

		if categoryID := ticket.CategoryID(); categoryID != nil {
			items[i].CategoryId = categoryID
		}

		if assignedTo := ticket.AssigneeID(); assignedTo != nil {
			items[i].AssigneeId = assignedTo
		}
	}

	// Set response headers for pagination
	c.Response().Header().Set("X-Total-Count", strconv.FormatInt(total, 10))

	// Create pagination
	page := 1
	if params.Page != nil {
		page = *params.Page
	}
	limit := filter.Limit
	totalPages := int((total + int64(limit) - 1) / int64(limit))
	hasNext := page < totalPages

	totalInt := int(total)
	pagination := openapi.PaginationResponse{
		Total:   &totalInt,
		Page:    &page,
		Limit:   &limit,
		HasNext: &hasNext,
	}

	response := openapi.ListTicketsResponse{
		Tickets:    &items,
		Pagination: &pagination,
	}

	return c.JSON(http.StatusOK, response)
}
