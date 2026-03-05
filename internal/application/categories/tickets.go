package categories

import (
	"context"
	"net/http"

	"simpleservicedesk/generated/openapi"
	ticketdomain "simpleservicedesk/internal/domain/tickets"
	userdomain "simpleservicedesk/internal/domain/users"
	"simpleservicedesk/internal/queries"
	"simpleservicedesk/pkg/echomiddleware"

	"github.com/google/uuid"
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

	claims, ok := echomiddleware.GetAuthClaimsFromContext(ctx)
	if !ok || claims == nil {
		msg := "unauthorized"
		return c.JSON(http.StatusUnauthorized, openapi.ErrorResponse{Message: &msg})
	}
	if claims.Role == userdomain.RoleCustomer {
		authorID, parseErr := uuid.Parse(claims.UserID)
		if parseErr != nil {
			msg := "unauthorized"
			return c.JSON(http.StatusUnauthorized, openapi.ErrorResponse{Message: &msg})
		}
		filter.AuthorID = &authorID
	}

	categoryIDs, idsErr := h.getRequestedCategoryIDs(ctx, id, params)
	if idsErr != nil {
		msg := idsErr.Error()
		return c.JSON(http.StatusInternalServerError, openapi.ErrorResponse{Message: &msg})
	}
	if len(categoryIDs) == 1 {
		filter.CategoryID = &categoryIDs[0]
	} else {
		filter.CategoryID = nil
		filter.CategoryIDs = categoryIDs
	}

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

func (h CategoryHandlers) getRequestedCategoryIDs(
	ctx context.Context,
	rootID uuid.UUID,
	params openapi.GetCategoriesIDTicketsParams,
) ([]uuid.UUID, error) {
	if params.IncludeSubcategories == nil || !*params.IncludeSubcategories {
		return []uuid.UUID{rootID}, nil
	}

	ids := []uuid.UUID{rootID}
	visited := map[uuid.UUID]struct{}{rootID: {}}
	queue := []uuid.UUID{rootID}

	for len(queue) > 0 {
		currentID := queue[0]
		queue = queue[1:]

		childFilter := queries.CategoryFilter{ParentID: &currentID}
		children, err := h.repo.ListCategories(ctx, childFilter)
		if err != nil {
			return nil, err
		}

		for _, child := range children {
			childID := child.ID()
			if _, seen := visited[childID]; seen {
				continue
			}
			visited[childID] = struct{}{}
			ids = append(ids, childID)
			queue = append(queue, childID)
		}
	}

	return ids, nil
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
