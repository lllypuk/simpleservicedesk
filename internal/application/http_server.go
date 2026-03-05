package application

import (
	"log/slog"
	"net/http"
	"time"

	"simpleservicedesk/generated/openapi"
	"simpleservicedesk/internal/application/auth"
	"simpleservicedesk/internal/application/categories"
	"simpleservicedesk/internal/application/organizations"
	"simpleservicedesk/internal/application/tickets"
	"simpleservicedesk/internal/application/users"
	userdomain "simpleservicedesk/internal/domain/users"
	"simpleservicedesk/pkg/echomiddleware"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type httpServer struct {
	auth.Handlers
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
	jwtSigningKey string,
	jwtExpiration time.Duration,
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
	authService, err := auth.NewService(userRepo, jwtSigningKey, jwtExpiration)
	if err == nil {
		server.Handlers = auth.SetupHandlers(authService)
	}

	server.UserHandlers = users.SetupHandlers(userRepo)
	server.TicketHandlers = tickets.SetupHandlers(ticketRepo)
	server.CategoryHandlers = categories.SetupHandlers(categoryRepo)
	server.OrganizationHandlers = organizations.SetupHandlers(organizationRepo)

	registerRoutes(e, server, authService)

	return e
}

func registerRoutes(e *echo.Echo, server httpServer, tokenValidator echomiddleware.TokenValidator) {
	wrapper := openapi.ServerInterfaceWrapper{Handler: server}

	authMiddleware := echomiddleware.Auth(tokenValidator)
	requireAgent := echomiddleware.RequireRole(userdomain.RoleAgent)
	requireAdmin := echomiddleware.RequireRole(userdomain.RoleAdmin)

	// Public endpoints.
	e.POST("/login", wrapper.PostLogin)

	// Authenticated endpoints (customer and above).
	e.GET("/categories", wrapper.GetCategories, authMiddleware)
	e.POST("/categories", wrapper.PostCategories, authMiddleware)
	e.DELETE("/categories/:id", wrapper.DeleteCategoriesID, authMiddleware)
	e.GET("/categories/:id", wrapper.GetCategoriesID, authMiddleware)
	e.PUT("/categories/:id", wrapper.PutCategoriesID, authMiddleware)
	e.GET("/categories/:id/tickets", wrapper.GetCategoriesIDTickets, authMiddleware)

	e.GET("/organizations", wrapper.GetOrganizations, authMiddleware)
	e.POST("/organizations", wrapper.PostOrganizations, authMiddleware)
	e.DELETE("/organizations/:id", wrapper.DeleteOrganizationsID, authMiddleware)
	e.GET("/organizations/:id", wrapper.GetOrganizationsID, authMiddleware)
	e.PUT("/organizations/:id", wrapper.PutOrganizationsID, authMiddleware)
	e.GET("/organizations/:id/tickets", wrapper.GetOrganizationsIDTickets, authMiddleware)
	e.GET("/organizations/:id/users", wrapper.GetOrganizationsIDUsers, authMiddleware)

	e.GET("/tickets", wrapper.GetTickets, authMiddleware)
	e.POST("/tickets", wrapper.PostTickets, authMiddleware)
	e.DELETE("/tickets/:id", wrapper.DeleteTicketsID, authMiddleware)
	e.GET("/tickets/:id", wrapper.GetTicketsID, authMiddleware)
	e.PUT("/tickets/:id", wrapper.PutTicketsID, authMiddleware)
	e.GET("/tickets/:id/comments", wrapper.GetTicketsIDComments, authMiddleware)
	e.POST("/tickets/:id/comments", wrapper.PostTicketsIDComments, authMiddleware)

	e.GET("/users/:id", wrapper.GetUsersID, authMiddleware)
	e.PUT("/users/:id", wrapper.PutUsersID, authMiddleware)
	e.GET("/users/:id/tickets", wrapper.GetUsersIDTickets, authMiddleware)

	// Agent+ endpoints.
	e.PATCH("/tickets/:id/assign", wrapper.PatchTicketsIDAssign, authMiddleware, requireAgent)
	e.PATCH("/tickets/:id/status", wrapper.PatchTicketsIDStatus, authMiddleware, requireAgent)
	e.GET("/users", wrapper.GetUsers, authMiddleware, requireAgent)

	// Admin-only endpoints.
	e.POST("/users", wrapper.PostUsers, authMiddleware, requireAdmin)
	e.DELETE("/users/:id", wrapper.DeleteUsersID, authMiddleware, requireAdmin)
	e.PATCH("/users/:id/role", wrapper.PatchUsersIDRole, authMiddleware, requireAdmin)
}
