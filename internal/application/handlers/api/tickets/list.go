package tickets

import (
	"net/http"

	"simpleservicedesk/generated/openapi"
	"simpleservicedesk/internal/application/handlers/api"
	"simpleservicedesk/internal/domain/tickets"
	"simpleservicedesk/internal/queries"

	"github.com/labstack/echo/v4"
)

func (h TicketHandlers) GetTickets(c echo.Context, params openapi.GetTicketsParams) error {
	ctx := c.Request().Context()

	// Set up filter with defaults
	filter := queries.TicketFilter{
		BaseFilter: queries.BaseFilter{
			Limit:  api.DefaultLimit,
			Offset: 0,
		},
	}

	// Apply pagination parameters
	if params.Limit != nil {
		if *params.Limit < 1 {
			msg := "limit must be at least 1"
			return c.JSON(http.StatusBadRequest, openapi.ErrorResponse{Message: &msg})
		}
		filter.Limit = *params.Limit
	}
	if params.Page != nil {
		if *params.Page < 1 {
			msg := "page must be at least 1"
			return c.JSON(http.StatusBadRequest, openapi.ErrorResponse{Message: &msg})
		}
		filter.Offset = (*params.Page - 1) * filter.Limit
	}

	// Apply status filter
	if params.Status != nil {
		status, err := tickets.ParseStatus(string(*params.Status))
		if err != nil {
			msg := "invalid status: " + err.Error()
			return c.JSON(http.StatusBadRequest, openapi.ErrorResponse{Message: &msg})
		}
		filter.Status = &status
	}

	// Apply priority filter
	if params.Priority != nil {
		priority, err := tickets.ParsePriority(string(*params.Priority))
		if err != nil {
			msg := "invalid priority: " + err.Error()
			return c.JSON(http.StatusBadRequest, openapi.ErrorResponse{Message: &msg})
		}
		filter.Priority = &priority
	}

	// Apply assignee filter
	if params.AssigneeId != nil {
		filter.AssigneeID = params.AssigneeId
	}

	// Apply category filter
	if params.CategoryId != nil {
		filter.CategoryID = params.CategoryId
	}

	// Validate and set defaults for filter
	filter, err := filter.ValidateAndSetDefaults()
	if err != nil {
		msg := err.Error()
		return c.JSON(http.StatusBadRequest, openapi.ErrorResponse{Message: &msg})
	}

	// Get tickets using service
	ticketList, total, err := h.ticketService.ListTickets(ctx, filter)
	if err != nil {
		msg := err.Error()
		return c.JSON(http.StatusInternalServerError, openapi.ErrorResponse{Message: &msg})
	}

	// Convert tickets to response format
	response := make([]openapi.GetTicketResponse, len(ticketList))
	for i, ticket := range ticketList {
		response[i] = convertTicketToResponse(ticket)
	}

	// Prepare pagination response
	totalInt := int(total)
	pagination := &openapi.PaginationResponse{
		Page:  params.Page,
		Limit: &filter.Limit,
		Total: &totalInt,
	}

	return c.JSON(http.StatusOK, openapi.ListTicketsResponse{
		Tickets:    &response,
		Pagination: pagination,
	})
}
