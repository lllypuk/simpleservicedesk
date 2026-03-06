package tickets

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"simpleservicedesk/generated/openapi"
	"simpleservicedesk/internal/domain/tickets"
	userdomain "simpleservicedesk/internal/domain/users"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func (h TicketHandlers) PostTickets(c echo.Context) error {
	ctx := c.Request().Context()
	var req openapi.CreateTicketRequest
	if err := c.Bind(&req); err != nil {
		return err
	}

	priority := tickets.Priority(req.Priority)

	// Convert OpenAPI types to uuid.UUID
	organizationID := req.OrganizationId
	authorID := req.AuthorId

	authUserID, role, ok := authUser(c)
	if !ok {
		return nil
	}
	if role == userdomain.RoleCustomer {
		authorID = authUserID
		if err := h.validateCustomerTicketOrganization(ctx, authUserID, organizationID); err != nil {
			if errors.Is(err, tickets.ErrUnauthorizedAccess) {
				msg := err.Error()
				return c.JSON(http.StatusForbidden, openapi.ErrorResponse{Message: &msg})
			}
			msg := err.Error()
			return c.JSON(http.StatusInternalServerError, openapi.ErrorResponse{Message: &msg})
		}
	}

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
		if errors.Is(err, tickets.ErrTicketValidation) ||
			errors.Is(err, tickets.ErrInvalidTicket) ||
			errors.Is(err, tickets.ErrInvalidPriority) {
			return c.JSON(http.StatusBadRequest, openapi.ErrorResponse{Message: &msg})
		}
		return c.JSON(http.StatusInternalServerError, openapi.ErrorResponse{Message: &msg})
	}

	response := convertTicketToResponse(ticket)
	return c.JSON(http.StatusCreated, response)
}

func (h TicketHandlers) validateCustomerTicketOrganization(
	ctx context.Context,
	userID uuid.UUID,
	organizationID uuid.UUID,
) error {
	if h.userRepo == nil {
		return nil
	}

	user, err := h.userRepo.GetUser(ctx, userID)
	if err != nil {
		return fmt.Errorf("get authenticated user: %w", err)
	}

	userOrgID := user.OrganizationID()
	if userOrgID != nil && *userOrgID != organizationID {
		return fmt.Errorf("%w: customers can create tickets only in their organization", tickets.ErrUnauthorizedAccess)
	}

	return nil
}
