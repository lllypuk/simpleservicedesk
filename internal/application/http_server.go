package application

import (
	"log/slog"
	"net/http"

	"simpleservicedesk/generated/openapi"
	"simpleservicedesk/internal/application/handlers/api/categories"
	"simpleservicedesk/internal/application/handlers/api/organizations"
	"simpleservicedesk/internal/application/handlers/api/tickets"
	"simpleservicedesk/internal/application/handlers/api/users"
	"simpleservicedesk/internal/application/services"
	"simpleservicedesk/internal/interfaces"
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
	userRepo interfaces.UserRepository,
	ticketRepo interfaces.TicketRepository,
	organizationRepo interfaces.OrganizationRepository,
	categoryRepo interfaces.CategoryRepository,
) *echo.Echo {
	e := echo.New()

	e.Pre(middleware.RemoveTrailingSlash())
	e.Use(echomiddleware.SlogLoggerMiddleware(slog.Default()))
	e.Use(echomiddleware.PutRequestIDContext)
	e.Use(middleware.Recover())

	e.GET("/ping", func(c echo.Context) error {
		return c.String(http.StatusOK, "pong")
	})

	// Create services
	userService := services.NewUserService(userRepo, ticketRepo)
	ticketService := services.NewTicketService(ticketRepo)
	categoryService := services.NewCategoryService(categoryRepo, ticketRepo)
	organizationService := services.NewOrganizationService(organizationRepo, userRepo, ticketRepo)

	// Setup handlers with services
	server := httpServer{}
	server.UserHandlers = users.SetupHandlers(userService)
	server.TicketHandlers = tickets.SetupHandlers(ticketService)
	server.CategoryHandlers = categories.SetupHandlers(categoryService)
	server.OrganizationHandlers = organizations.SetupHandlers(organizationService)

	// Register routes generated from OpenAPI
	openapi.RegisterHandlers(e, server)

	return e
}
