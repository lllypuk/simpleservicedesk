package organizations

import (
	"errors"
	"net/http"

	"simpleservicedesk/generated/openapi"
	"simpleservicedesk/internal/application/handlers/api"
	"simpleservicedesk/internal/application/services"
	"simpleservicedesk/internal/domain/organizations"
	"simpleservicedesk/internal/domain/tickets"
	"simpleservicedesk/internal/queries"

	"github.com/labstack/echo/v4"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

type OrganizationHandlers struct {
	organizationService services.OrganizationService
}

func SetupHandlers(organizationService services.OrganizationService) OrganizationHandlers {
	return OrganizationHandlers{
		organizationService: organizationService,
	}
}

func (h OrganizationHandlers) GetOrganizations(c echo.Context, params openapi.GetOrganizationsParams) error {
	filter := queries.OrganizationFilter{}

	if params.Limit != nil {
		filter.Limit = *params.Limit
	} else {
		filter.Limit = 50 // Default limit
	}

	// Convert page to offset
	if params.Page != nil {
		page := *params.Page
		if page > 0 {
			filter.Offset = (page - 1) * filter.Limit
		}
	}

	if params.Name != nil {
		filter.Name = params.Name
	}

	if params.Domain != nil {
		filter.Domain = params.Domain
	}

	if params.IsActive != nil {
		filter.IsActive = params.IsActive
	}

	// Validate filter
	if err := filter.Validate(); err != nil {
		errMsg := err.Error()
		return c.JSON(http.StatusBadRequest, openapi.ErrorResponse{
			Message: &errMsg,
		})
	}

	orgs, total, err := h.organizationService.ListOrganizations(c.Request().Context(), filter)
	if err != nil {
		errMsg := err.Error()
		return c.JSON(http.StatusInternalServerError, openapi.ErrorResponse{
			Message: &errMsg,
		})
	}

	response := make([]openapi.GetOrganizationResponse, len(orgs))
	for i, org := range orgs {
		response[i] = mapDomainToAPIOrganization(org)
	}

	totalInt := int(total)
	pagination := &openapi.PaginationResponse{
		Page:  params.Page,
		Limit: &filter.Limit,
		Total: &totalInt,
	}

	return c.JSON(http.StatusOK, openapi.ListOrganizationsResponse{
		Organizations: &response,
		Pagination:    pagination,
	})
}

func (h OrganizationHandlers) PostOrganizations(c echo.Context) error {
	var req openapi.CreateOrganizationRequest
	if err := c.Bind(&req); err != nil {
		errMsg := api.MsgInvalidRequestFormat
		return c.JSON(http.StatusBadRequest, openapi.ErrorResponse{
			Message: &errMsg,
		})
	}

	if req.Name == "" {
		errMsg := "name is required"
		return c.JSON(http.StatusBadRequest, openapi.ErrorResponse{
			Message: &errMsg,
		})
	}

	if req.Domain == nil || *req.Domain == "" {
		errMsg := "domain is required"
		return c.JSON(http.StatusBadRequest, openapi.ErrorResponse{
			Message: &errMsg,
		})
	}

	serviceReq := services.CreateOrganizationRequest{
		Name:     req.Name,
		Domain:   *req.Domain,
		ParentID: req.ParentId,
	}

	org, err := h.organizationService.CreateOrganization(
		c.Request().Context(),
		serviceReq,
	)
	if err != nil {
		if errors.Is(err, organizations.ErrOrganizationAlreadyExist) {
			errMsg := err.Error()
			return c.JSON(http.StatusConflict, openapi.ErrorResponse{
				Message: &errMsg,
			})
		}
		if errors.Is(err, organizations.ErrOrganizationValidation) {
			errMsg := err.Error()
			return c.JSON(http.StatusBadRequest, openapi.ErrorResponse{
				Message: &errMsg,
			})
		}
		errMsg := err.Error()
		return c.JSON(http.StatusInternalServerError, openapi.ErrorResponse{
			Message: &errMsg,
		})
	}

	orgID := org.ID()
	response := openapi.CreateOrganizationResponse{
		Id: &orgID,
	}

	return c.JSON(http.StatusCreated, response)
}

func (h OrganizationHandlers) GetOrganizationsID(c echo.Context, id openapi_types.UUID) error {
	org, err := h.organizationService.GetOrganization(c.Request().Context(), id)
	if err != nil {
		if errors.Is(err, organizations.ErrOrganizationNotFound) {
			errMsg := api.MsgOrganizationNotFound
			return c.JSON(http.StatusNotFound, openapi.ErrorResponse{
				Message: &errMsg,
			})
		}
		errMsg := err.Error()
		return c.JSON(http.StatusInternalServerError, openapi.ErrorResponse{
			Message: &errMsg,
		})
	}

	response := mapDomainToAPIOrganization(org)
	return c.JSON(http.StatusOK, response)
}

func (h OrganizationHandlers) PutOrganizationsID(c echo.Context, id openapi_types.UUID) error {
	var req openapi.UpdateOrganizationRequest
	if err := c.Bind(&req); err != nil {
		errMsg := api.MsgInvalidRequestFormat
		return c.JSON(http.StatusBadRequest, openapi.ErrorResponse{
			Message: &errMsg,
		})
	}

	serviceReq := services.UpdateOrganizationRequest{
		Name:     req.Name,
		Domain:   req.Domain,
		ParentID: req.ParentId,
		IsActive: req.IsActive,
	}

	org, err := h.organizationService.UpdateOrganization(
		c.Request().Context(),
		id,
		serviceReq,
	)
	if err != nil {
		if errors.Is(err, organizations.ErrOrganizationNotFound) {
			errMsg := api.MsgOrganizationNotFound
			return c.JSON(http.StatusNotFound, openapi.ErrorResponse{
				Message: &errMsg,
			})
		}
		if errors.Is(err, organizations.ErrOrganizationValidation) {
			errMsg := err.Error()
			return c.JSON(http.StatusBadRequest, openapi.ErrorResponse{
				Message: &errMsg,
			})
		}
		errMsg := err.Error()
		return c.JSON(http.StatusInternalServerError, openapi.ErrorResponse{
			Message: &errMsg,
		})
	}

	response := mapDomainToAPIOrganization(org)
	return c.JSON(http.StatusOK, response)
}

func (h OrganizationHandlers) DeleteOrganizationsID(c echo.Context, id openapi_types.UUID) error {
	err := h.organizationService.DeleteOrganization(c.Request().Context(), id)
	if err != nil {
		if errors.Is(err, organizations.ErrOrganizationNotFound) {
			errMsg := api.MsgOrganizationNotFound
			return c.JSON(http.StatusNotFound, openapi.ErrorResponse{
				Message: &errMsg,
			})
		}
		errMsg := err.Error()
		return c.JSON(http.StatusInternalServerError, openapi.ErrorResponse{
			Message: &errMsg,
		})
	}

	return c.NoContent(http.StatusNoContent)
}

func (h OrganizationHandlers) GetOrganizationsIDTickets(
	c echo.Context,
	id openapi_types.UUID,
	params openapi.GetOrganizationsIDTicketsParams,
) error {
	filter := queries.TicketFilter{
		OrganizationID: &id,
	}

	if params.Limit != nil {
		filter.Limit = *params.Limit
	} else {
		filter.Limit = 50 // Default limit
	}

	// Convert page to offset
	if params.Page != nil {
		page := *params.Page
		if page > 0 {
			filter.Offset = (page - 1) * filter.Limit
		}
	}

	if params.Status != nil {
		status := tickets.Status(*params.Status)
		filter.Status = &status
	}

	if params.Priority != nil {
		priority := tickets.Priority(*params.Priority)
		filter.Priority = &priority
	}

	// Validate filter
	filter, err := filter.ValidateAndSetDefaults()
	if err != nil {
		errMsg := err.Error()
		return c.JSON(http.StatusBadRequest, openapi.ErrorResponse{
			Message: &errMsg,
		})
	}

	tickets, total, err := h.organizationService.GetOrganizationTickets(c.Request().Context(), id, filter)
	if err != nil {
		if errors.Is(err, organizations.ErrOrganizationNotFound) {
			errMsg := api.MsgOrganizationNotFound
			return c.JSON(http.StatusNotFound, openapi.ErrorResponse{
				Message: &errMsg,
			})
		}
		errMsg := err.Error()
		return c.JSON(http.StatusInternalServerError, openapi.ErrorResponse{
			Message: &errMsg,
		})
	}

	response := make([]openapi.GetTicketResponse, len(tickets))
	for i, ticket := range tickets {
		response[i] = mapDomainToAPITicket(ticket)
	}

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

func (h OrganizationHandlers) GetOrganizationsIDUsers(
	c echo.Context,
	id openapi_types.UUID,
	params openapi.GetOrganizationsIDUsersParams,
) error {
	filter := queries.UserFilter{
		OrganizationID: &id,
	}

	if params.Limit != nil {
		filter.Limit = *params.Limit
	} else {
		filter.Limit = 50 // Default limit
	}

	// Convert page to offset
	if params.Page != nil {
		page := *params.Page
		if page > 0 {
			filter.Offset = (page - 1) * filter.Limit
		}
	}

	// Validate filter
	if err := filter.Validate(); err != nil {
		errMsg := err.Error()
		return c.JSON(http.StatusBadRequest, openapi.ErrorResponse{
			Message: &errMsg,
		})
	}

	users, total, err := h.organizationService.GetOrganizationUsers(c.Request().Context(), id, filter)
	if err != nil {
		if errors.Is(err, organizations.ErrOrganizationNotFound) {
			errMsg := api.MsgOrganizationNotFound
			return c.JSON(http.StatusNotFound, openapi.ErrorResponse{
				Message: &errMsg,
			})
		}
		errMsg := err.Error()
		return c.JSON(http.StatusInternalServerError, openapi.ErrorResponse{
			Message: &errMsg,
		})
	}

	response := make([]openapi.GetUserResponse, len(users))
	for i, user := range users {
		response[i] = mapDomainToAPIUser(user)
	}

	totalInt := int(total)
	pagination := &openapi.PaginationResponse{
		Page:  params.Page,
		Limit: &filter.Limit,
		Total: &totalInt,
	}

	return c.JSON(http.StatusOK, openapi.ListUsersResponse{
		Users:      &response,
		Pagination: pagination,
	})
}

// Helper functions to map domain objects to API objects

func mapDomainToAPIOrganization(org *organizations.Organization) openapi.GetOrganizationResponse {
	orgID := org.ID()
	orgName := org.Name()
	orgDomain := org.Domain()
	orgIsActive := org.IsActive()
	orgCreatedAt := org.CreatedAt()
	orgUpdatedAt := org.UpdatedAt()

	apiOrg := openapi.GetOrganizationResponse{
		Id:        &orgID,
		Name:      &orgName,
		Domain:    &orgDomain,
		IsActive:  &orgIsActive,
		CreatedAt: &orgCreatedAt,
		UpdatedAt: &orgUpdatedAt,
	}

	if org.ParentID() != nil {
		parentID := *org.ParentID()
		apiOrg.ParentId = &parentID
	}

	return apiOrg
}

func mapDomainToAPITicket(ticket *tickets.Ticket) openapi.GetTicketResponse {
	id := ticket.ID()
	title := ticket.Title()
	description := ticket.Description()
	status := openapi.TicketStatus(ticket.Status())
	priority := openapi.TicketPriority(ticket.Priority())
	organizationID := ticket.OrganizationID()
	authorID := ticket.AuthorID()
	assigneeID := ticket.AssigneeID()
	categoryID := ticket.CategoryID()
	createdAt := ticket.CreatedAt()
	updatedAt := ticket.UpdatedAt()

	return openapi.GetTicketResponse{
		Id:             &id,
		Title:          &title,
		Description:    &description,
		Status:         &status,
		Priority:       &priority,
		OrganizationId: &organizationID,
		AuthorId:       &authorID,
		AssigneeId:     assigneeID,
		CategoryId:     categoryID,
		CreatedAt:      &createdAt,
		UpdatedAt:      &updatedAt,
	}
}

func mapDomainToAPIUser(_ interface{}) openapi.GetUserResponse {
	// This is a placeholder - the actual implementation would depend on the user domain structure
	// Since we don't have the users domain imported here, this is a basic structure
	return openapi.GetUserResponse{}
}
