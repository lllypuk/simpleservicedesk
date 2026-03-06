package health_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"simpleservicedesk/internal/application/health"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockPinger struct {
	err error
}

func (m *mockPinger) Ping(_ context.Context) error {
	return m.err
}

func newTestContext(e *echo.Echo, method, path string) (echo.Context, *httptest.ResponseRecorder) {
	req := httptest.NewRequest(method, path, nil)
	rec := httptest.NewRecorder()
	return e.NewContext(req, rec), rec
}

func TestLiveHandler_AlwaysHealthy(t *testing.T) {
	e := echo.New()
	h := health.NewHandlers(&mockPinger{err: errors.New("db down")})

	c, rec := newTestContext(e, http.MethodGet, "/health/live")
	require.NoError(t, h.LiveHandler(c))

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"healthy"`)
}

func TestReadyHandler_MongoUp(t *testing.T) {
	e := echo.New()
	h := health.NewHandlers(&mockPinger{})

	c, rec := newTestContext(e, http.MethodGet, "/health/ready")
	require.NoError(t, h.ReadyHandler(c))

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), `"healthy"`)
	assert.Contains(t, rec.Body.String(), `"up"`)
}

func TestReadyHandler_MongoDown(t *testing.T) {
	e := echo.New()
	h := health.NewHandlers(&mockPinger{err: errors.New("connection refused")})

	c, rec := newTestContext(e, http.MethodGet, "/health/ready")
	require.NoError(t, h.ReadyHandler(c))

	assert.Equal(t, http.StatusServiceUnavailable, rec.Code)
	assert.Contains(t, rec.Body.String(), `"unhealthy"`)
	assert.Contains(t, rec.Body.String(), `"down"`)
}
