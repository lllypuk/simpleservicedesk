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

type httpServer struct {
	users.UserHandlers
	tickets.TicketHandlers
	categories.CategoryHandlers
	organizations.OrganizationHandlers
}

func SetupHTTPServer(userRepo UserRepository, ticketRepo TicketRepository) *echo.Echo {
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
	server.CategoryHandlers = categories.SetupHandlers()
	server.OrganizationHandlers = organizations.SetupHandlers()

	// Register routes generated from OpenAPI
	openapi.RegisterHandlers(e, server)

	return e
}
