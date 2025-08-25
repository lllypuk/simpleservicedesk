package application

import (
	"log/slog"
	"net/http"

	"simpleservicedesk/generated/openapi"
	"simpleservicedesk/internal/application/categories"
	"simpleservicedesk/internal/application/organizations"
	"simpleservicedesk/internal/application/tickets"
	"simpleservicedesk/internal/application/users"
	"simpleservicedesk/pkg/echomiddleware"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

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
	server.TicketHandlers = tickets.SetupHandlers(ticketRepo)
	server.CategoryHandlers = categories.SetupHandlers(categoryRepo)
	server.OrganizationHandlers = organizations.SetupHandlers(organizationRepo)

	// Register routes generated from OpenAPI
	openapi.RegisterHandlers(e, server)

	return e
}
