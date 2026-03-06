package health

import (
	"context"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

const pingTimeout = 5 * time.Second

// Handlers holds health check HTTP handlers.
type Handlers struct {
	pinger Pinger
}

// NewHandlers creates health Handlers with the given pinger.
func NewHandlers(pinger Pinger) Handlers {
	return Handlers{pinger: pinger}
}

// LiveHandler handles GET /health/live — always 200 if the process is running.
func (h Handlers) LiveHandler(c echo.Context) error {
	return c.JSON(http.StatusOK, Status{
		Status: "healthy",
		Checks: []CheckResult{},
	})
}

// ReadyHandler handles GET /health/ready — checks MongoDB connectivity.
func (h Handlers) ReadyHandler(c echo.Context) error {
	ctx, cancel := context.WithTimeout(c.Request().Context(), pingTimeout)
	defer cancel()

	result := CheckMongoDB(ctx, h.pinger)

	s := Status{
		Status: "healthy",
		Checks: []CheckResult{result},
	}
	code := http.StatusOK
	if result.Status != "up" {
		s.Status = "unhealthy"
		code = http.StatusServiceUnavailable
	}

	return c.JSON(code, s)
}
