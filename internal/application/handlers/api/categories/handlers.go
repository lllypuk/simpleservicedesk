package categories

import (
	"errors"
	"net/http"

	"simpleservicedesk/generated/openapi"
	"simpleservicedesk/internal/application/handlers/api"
	"simpleservicedesk/internal/application/services"
	"simpleservicedesk/internal/domain/categories"
	"simpleservicedesk/internal/domain/tickets"
	"simpleservicedesk/internal/queries"

	"github.com/labstack/echo/v4"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

type CategoryHandlers struct {
	categoryService services.CategoryService
}

func SetupHandlers(categoryService services.CategoryService) CategoryHandlers {
	return CategoryHandlers{
		categoryService: categoryService,
	}
}

func (h CategoryHandlers) GetCategories(c echo.Context, params openapi.GetCategoriesParams) error {
	filter := queries.CategoryFilter{}

	// Default limit
	filter.Limit = 50

	if params.OrganizationId != nil {
		filter.OrganizationID = params.OrganizationId
	}

	if params.ParentId != nil {
		filter.ParentID = params.ParentId
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

	cats, _, err := h.categoryService.ListCategories(c.Request().Context(), filter)
	if err != nil {
		errMsg := err.Error()
		return c.JSON(http.StatusInternalServerError, openapi.ErrorResponse{
			Message: &errMsg,
		})
	}

	response := make([]openapi.GetCategoryResponse, len(cats))
	for i, cat := range cats {
		response[i] = mapDomainToAPICategory(cat)
	}

	return c.JSON(http.StatusOK, openapi.ListCategoriesResponse{
		Categories: &response,
	})
}

func (h CategoryHandlers) PostCategories(c echo.Context) error {
	var req openapi.CreateCategoryRequest
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

	serviceReq := services.CreateCategoryRequest{
		Name:           req.Name,
		Description:    req.Description,
		OrganizationID: req.OrganizationId,
		ParentID:       req.ParentId,
	}

	cat, err := h.categoryService.CreateCategory(
		c.Request().Context(),
		serviceReq,
	)
	if err != nil {
		if errors.Is(err, categories.ErrCategoryValidation) {
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

	catID := cat.ID()
	response := openapi.CreateCategoryResponse{
		Id: &catID,
	}

	return c.JSON(http.StatusCreated, response)
}

func (h CategoryHandlers) GetCategoriesID(c echo.Context, id openapi_types.UUID) error {
	cat, err := h.categoryService.GetCategory(c.Request().Context(), id)
	if err != nil {
		if errors.Is(err, categories.ErrCategoryNotFound) {
			errMsg := api.MsgCategoryNotFound
			return c.JSON(http.StatusNotFound, openapi.ErrorResponse{
				Message: &errMsg,
			})
		}
		errMsg := err.Error()
		return c.JSON(http.StatusInternalServerError, openapi.ErrorResponse{
			Message: &errMsg,
		})
	}

	response := mapDomainToAPICategory(cat)
	return c.JSON(http.StatusOK, response)
}

func (h CategoryHandlers) PutCategoriesID(c echo.Context, id openapi_types.UUID) error {
	var req openapi.UpdateCategoryRequest
	if err := c.Bind(&req); err != nil {
		errMsg := api.MsgInvalidRequestFormat
		return c.JSON(http.StatusBadRequest, openapi.ErrorResponse{
			Message: &errMsg,
		})
	}

	serviceReq := services.UpdateCategoryRequest{
		Name:        req.Name,
		Description: req.Description,
		ParentID:    req.ParentId,
		IsActive:    req.IsActive,
	}

	cat, err := h.categoryService.UpdateCategory(
		c.Request().Context(),
		id,
		serviceReq,
	)
	if err != nil {
		if errors.Is(err, categories.ErrCategoryNotFound) {
			errMsg := api.MsgCategoryNotFound
			return c.JSON(http.StatusNotFound, openapi.ErrorResponse{
				Message: &errMsg,
			})
		}
		if errors.Is(err, categories.ErrCategoryValidation) {
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

	response := mapDomainToAPICategory(cat)
	return c.JSON(http.StatusOK, response)
}

func (h CategoryHandlers) GetCategoriesIDTickets(
	c echo.Context,
	id openapi_types.UUID,
	params openapi.GetCategoriesIDTicketsParams,
) error {
	filter := queries.TicketFilter{
		CategoryID: &id,
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

	tickets, total, err := h.categoryService.GetCategoryTickets(c.Request().Context(), id, filter)
	if err != nil {
		if errors.Is(err, categories.ErrCategoryNotFound) {
			errMsg := api.MsgCategoryNotFound
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

// Helper functions to map domain objects to API objects

func mapDomainToAPICategory(cat *categories.Category) openapi.GetCategoryResponse {
	catID := cat.ID()
	catName := cat.Name()
	catDescription := cat.Description()
	catOrgID := cat.OrganizationID()
	catIsActive := cat.IsActive()
	catCreatedAt := cat.CreatedAt()
	catUpdatedAt := cat.UpdatedAt()

	apiCat := openapi.GetCategoryResponse{
		Id:             &catID,
		Name:           &catName,
		OrganizationId: &catOrgID,
		IsActive:       &catIsActive,
		CreatedAt:      &catCreatedAt,
		UpdatedAt:      &catUpdatedAt,
	}

	if catDescription != "" {
		apiCat.Description = &catDescription
	}

	if cat.ParentID() != nil {
		parentID := *cat.ParentID()
		apiCat.ParentId = &parentID
	}

	return apiCat
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
