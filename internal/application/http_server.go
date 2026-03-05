package application

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"simpleservicedesk/generated/openapi"
	"simpleservicedesk/internal/application/auth"
	"simpleservicedesk/internal/application/categories"
	"simpleservicedesk/internal/application/organizations"
	"simpleservicedesk/internal/application/tickets"
	"simpleservicedesk/internal/application/users"
	userdomain "simpleservicedesk/internal/domain/users"
	appmiddleware "simpleservicedesk/pkg/echomiddleware"

	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	oapimiddleware "github.com/oapi-codegen/echo-middleware"
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
	corsAllowedOrigins []string,
) (*echo.Echo, error) {
	e := echo.New()

	e.Pre(middleware.RemoveTrailingSlash())
	e.Use(appmiddleware.SlogLoggerMiddleware(slog.Default()))
	e.Use(appmiddleware.PutRequestIDContext)
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: corsAllowedOrigins,
		AllowMethods: []string{
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodDelete,
			http.MethodOptions,
		},
		AllowHeaders: []string{
			echo.HeaderContentType,
			echo.HeaderAuthorization,
		},
		ExposeHeaders: []string{
			echo.HeaderXRequestID,
		},
	}))
	e.Use(middleware.Recover())

	swagger, err := openapi.GetSwagger()
	if err != nil {
		return nil, err
	}
	e.Use(oapimiddleware.OapiRequestValidatorWithOptions(swagger, &oapimiddleware.Options{
		Skipper:      shouldSkipOpenAPIValidation,
		ErrorHandler: openAPIValidationErrorHandler,
		Options: openapi3filter.Options{
			AuthenticationFunc: func(_ context.Context, _ *openapi3filter.AuthenticationInput) error {
				return nil
			},
		},
	}))

	e.GET("/ping", func(c echo.Context) error {
		return c.String(http.StatusOK, "pong")
	})

	server := httpServer{}
	authService, err := auth.NewService(userRepo, jwtSigningKey, jwtExpiration)
	if err != nil {
		return nil, err
	}
	server.Handlers = auth.SetupHandlers(authService)

	server.UserHandlers = users.SetupHandlers(userRepo)
	server.TicketHandlers = tickets.SetupHandlers(ticketRepo)
	server.CategoryHandlers = categories.SetupHandlers(categoryRepo, ticketRepo)
	server.OrganizationHandlers = organizations.SetupHandlers(organizationRepo)

	registerRoutes(e, server, authService)

	return e, nil
}

func registerRoutes(e *echo.Echo, server httpServer, tokenValidator appmiddleware.TokenValidator) {
	wrapper := openapi.ServerInterfaceWrapper{Handler: server}

	authMiddleware := appmiddleware.Auth(tokenValidator)
	requireAgent := appmiddleware.RequireRole(userdomain.RoleAgent)
	requireAdmin := appmiddleware.RequireRole(userdomain.RoleAdmin)

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

func shouldSkipOpenAPIValidation(c echo.Context) bool {
	path := strings.TrimSpace(c.Request().URL.Path)
	path = strings.TrimSuffix(path, "/")
	if path == "" {
		path = "/"
	}

	return path == "/ping" || path == "/login"
}

func openAPIValidationErrorHandler(c echo.Context, err *echo.HTTPError) error {
	statusCode := http.StatusBadRequest
	if err != nil && err.Code > 0 {
		statusCode = err.Code
	}

	message := extractErrorMessage(err)

	return c.JSON(statusCode, openapi.ErrorResponse{Message: &message})
}

func extractErrorMessage(err *echo.HTTPError) string {
	if err == nil {
		return "request validation failed"
	}

	switch msg := err.Message.(type) {
	case string:
		if trimmed := strings.TrimSpace(msg); trimmed != "" {
			return trimmed
		}
	case error:
		if trimmed := strings.TrimSpace(msg.Error()); trimmed != "" {
			return trimmed
		}
	default:
		if msg != nil {
			if trimmed := strings.TrimSpace(fmt.Sprint(msg)); trimmed != "" {
				return trimmed
			}
		}
	}

	if err.Internal != nil {
		if trimmed := strings.TrimSpace(err.Internal.Error()); trimmed != "" {
			return trimmed
		}
	}

	return "request validation failed"
}
