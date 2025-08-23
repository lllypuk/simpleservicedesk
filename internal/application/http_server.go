package application

import (
	"context"
	"log/slog"
	"net/http"

	"simpleservicedesk/generated/openapi"
	"simpleservicedesk/internal/application/categories"
	"simpleservicedesk/internal/application/organizations"
	"simpleservicedesk/internal/application/tickets"
	"simpleservicedesk/internal/application/users"
	domainCategories "simpleservicedesk/internal/domain/categories"
	domainOrganizations "simpleservicedesk/internal/domain/organizations"
	domainTickets "simpleservicedesk/internal/domain/tickets"
	"simpleservicedesk/pkg/echomiddleware"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// ticketRepoAdapter adapts application.TicketRepository to tickets.TicketRepository
type ticketRepoAdapter struct {
	appRepo TicketRepository
}

func (a *ticketRepoAdapter) CreateTicket(
	ctx context.Context,
	createFn func() (*domainTickets.Ticket, error),
) (*domainTickets.Ticket, error) {
	return a.appRepo.CreateTicket(ctx, createFn)
}

func (a *ticketRepoAdapter) UpdateTicket(
	ctx context.Context,
	id uuid.UUID,
	updateFn func(*domainTickets.Ticket) (bool, error),
) (*domainTickets.Ticket, error) {
	return a.appRepo.UpdateTicket(ctx, id, updateFn)
}

func (a *ticketRepoAdapter) GetTicket(ctx context.Context, id uuid.UUID) (*domainTickets.Ticket, error) {
	return a.appRepo.GetTicket(ctx, id)
}

func (a *ticketRepoAdapter) ListTickets(
	ctx context.Context,
	filter tickets.TicketFilter,
) ([]*domainTickets.Ticket, error) {
	// Convert tickets.TicketFilter to application.TicketFilter
	appFilter := TicketFilter{
		Status:         filter.Status,
		Priority:       filter.Priority,
		AssigneeID:     filter.AssigneeID,
		AuthorID:       filter.AuthorID,
		OrganizationID: filter.OrganizationID,
		CategoryID:     filter.CategoryID,
		Limit:          filter.Limit,
		Offset:         filter.Offset,
	}
	return a.appRepo.ListTickets(ctx, appFilter)
}

func (a *ticketRepoAdapter) DeleteTicket(ctx context.Context, id uuid.UUID) error {
	return a.appRepo.DeleteTicket(ctx, id)
}

// organizationRepoAdapter adapts application.OrganizationRepository to organizations.Repository
type organizationRepoAdapter struct {
	appRepo OrganizationRepository
}

func (a *organizationRepoAdapter) CreateOrganization(
	ctx context.Context,
	createFn func() (*domainOrganizations.Organization, error),
) (*domainOrganizations.Organization, error) {
	return a.appRepo.CreateOrganization(ctx, createFn)
}

func (a *organizationRepoAdapter) UpdateOrganization(
	ctx context.Context,
	id uuid.UUID,
	updateFn func(*domainOrganizations.Organization) (bool, error),
) (*domainOrganizations.Organization, error) {
	return a.appRepo.UpdateOrganization(ctx, id, updateFn)
}

func (a *organizationRepoAdapter) GetOrganization(
	ctx context.Context,
	id uuid.UUID,
) (*domainOrganizations.Organization, error) {
	return a.appRepo.GetOrganization(ctx, id)
}

func (a *organizationRepoAdapter) ListOrganizations(
	ctx context.Context,
	filter organizations.OrganizationFilter,
) ([]*domainOrganizations.Organization, error) {
	// Convert organizations.OrganizationFilter to application.OrganizationFilter
	appFilter := OrganizationFilter{
		ParentID:   filter.ParentID,
		IsActive:   filter.IsActive,
		Name:       filter.Name,
		Domain:     filter.Domain,
		IsRootOnly: filter.IsRootOnly,
		Limit:      filter.Limit,
		Offset:     filter.Offset,
		SortBy:     filter.SortBy,
		SortOrder:  filter.SortOrder,
	}
	return a.appRepo.ListOrganizations(ctx, appFilter)
}

func (a *organizationRepoAdapter) DeleteOrganization(ctx context.Context, id uuid.UUID) error {
	return a.appRepo.DeleteOrganization(ctx, id)
}

// categoryRepoAdapter adapts application.CategoryRepository to categories.CategoryRepository
type categoryRepoAdapter struct {
	appRepo CategoryRepository
}

func (a *categoryRepoAdapter) CreateCategory(
	ctx context.Context,
	createFn func() (*domainCategories.Category, error),
) (*domainCategories.Category, error) {
	return a.appRepo.CreateCategory(ctx, createFn)
}

func (a *categoryRepoAdapter) UpdateCategory(
	ctx context.Context,
	id uuid.UUID,
	updateFn func(*domainCategories.Category) (bool, error),
) (*domainCategories.Category, error) {
	return a.appRepo.UpdateCategory(ctx, id, updateFn)
}

func (a *categoryRepoAdapter) GetCategory(ctx context.Context, id uuid.UUID) (*domainCategories.Category, error) {
	return a.appRepo.GetCategory(ctx, id)
}

func (a *categoryRepoAdapter) ListCategories(
	ctx context.Context,
	filter categories.CategoryFilter,
) ([]*domainCategories.Category, error) {
	// Convert categories.CategoryFilter to application.CategoryFilter
	appFilter := CategoryFilter{
		OrganizationID: filter.OrganizationID,
		ParentID:       filter.ParentID,
		IsActive:       filter.IsActive,
		Name:           filter.Name,
		IsRootOnly:     filter.IsRootOnly,
		Limit:          filter.Limit,
		Offset:         filter.Offset,
		SortBy:         filter.SortBy,
		SortOrder:      filter.SortOrder,
	}
	return a.appRepo.ListCategories(ctx, appFilter)
}

func (a *categoryRepoAdapter) DeleteCategory(ctx context.Context, id uuid.UUID) error {
	return a.appRepo.DeleteCategory(ctx, id)
}

type httpServer struct {
	users.UserHandlers
	tickets.TicketHandlers
	categories.CategoryHandlers
	organizations.OrganizationHandlers
}

func SetupHTTPServer(
	userRepo UserRepository,
	ticketRepo TicketRepository,
	organizationRepo OrganizationRepository,
	categoryRepo CategoryRepository,
) *echo.Echo {
	e := echo.New()

	e.Pre(middleware.RemoveTrailingSlash())
	e.Use(echomiddleware.SlogLoggerMiddleware(slog.Default()))
	e.Use(echomiddleware.PutRequestIDContext)
	e.Use(middleware.Recover())

	e.GET("/ping", func(c echo.Context) error {
		return c.String(http.StatusOK, "pong")
	})

	server := httpServer{}
	server.UserHandlers = users.SetupHandlers(userRepo)
	server.TicketHandlers = tickets.SetupHandlers(&ticketRepoAdapter{appRepo: ticketRepo})
	server.CategoryHandlers = categories.SetupHandlers(&categoryRepoAdapter{appRepo: categoryRepo})
	server.OrganizationHandlers = organizations.SetupHandlers(&organizationRepoAdapter{appRepo: organizationRepo})

	// Register routes generated from OpenAPI
	openapi.RegisterHandlers(e, server)

	return e
}
