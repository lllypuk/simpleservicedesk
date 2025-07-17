package application

import (
	"log/slog"
	"net/http"
	"simpleservicedesk/internal/application/users"
	"simpleservicedesk/pkg/echomiddleware"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type httpServer struct {
	users.UserHandlers
}

func SetupHTTPServer(userRepo UserRepository) *echo.Echo {
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

	return e
}
