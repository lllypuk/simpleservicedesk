package categories

import (
	"net/http"

	"simpleservicedesk/generated/openapi"
	ticketdomain "simpleservicedesk/internal/domain/tickets"
	"simpleservicedesk/internal/queries"

	"github.com/labstack/echo/v4"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

func (h CategoryHandlers) GetCategoriesIDTickets(
	c echo.Context, id openapi_types.UUID, params openapi.GetCategoriesIDTicketsParams,
) error {
	ctx := c.Request().Context()

	// First verify the category exists.
	_, err := h.repo.GetCategory(ctx, id)
	if err != nil {
		return h.handleCategoryError(c, err)
	}

	filter, err := queries.FromOpenAPITicketParams(openapi.GetTicketsParams{
		Status:   params.Status,
		Priority: params.Priority,
		Page:     params.Page,
		Limit:    params.Limit,
	})
	if err != nil {
		msg := err.Error()
		return c.JSON(http.StatusBadRequest, openapi.ErrorResponse{Message: &msg})
	}
	filter.CategoryID = &id

	ticketList, err := h.ticketRepo.ListTickets(ctx, filter)
	if err != nil {
		msg := err.Error()
		return c.JSON(http.StatusInternalServerError, openapi.ErrorResponse{Message: &msg})
	}

	ticketResponses := make([]openapi.GetTicketResponse, len(ticketList))
	for i, ticket := range ticketList {
		ticketResponses[i] = categoryTicketToResponse(ticket)
	}

	page := 1
	if params.Page != nil {
		page = *params.Page
	}
	limit := filter.Limit
	total := len(ticketResponses)
	hasNext := len(ticketResponses) == limit

	return c.JSON(http.StatusOK, openapi.ListTicketsResponse{
		Tickets: &ticketResponses,
		Pagination: &openapi.PaginationResponse{
			Total:   &total,
			Page:    &page,
			Limit:   &limit,
			HasNext: &hasNext,
		},
	})
}

func categoryTicketToResponse(ticket *ticketdomain.Ticket) openapi.GetTicketResponse {
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
