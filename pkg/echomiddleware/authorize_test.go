package echomiddleware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	authdomain "simpleservicedesk/internal/domain/auth"
	"simpleservicedesk/internal/domain/users"
	"simpleservicedesk/pkg/contextkeys"
	echoMw "simpleservicedesk/pkg/echomiddleware"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRequireRole(t *testing.T) {
	t.Run("missing claims returns 401", func(t *testing.T) {
		e := echo.New()
		e.Use(echoMw.RequireRole(users.RoleCustomer))
		e.GET("/protected", func(c echo.Context) error {
			return c.NoContent(http.StatusNoContent)
		})

		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		rec := httptest.NewRecorder()

		e.ServeHTTP(rec, req)

		require.Equal(t, http.StatusUnauthorized, rec.Code)
	})

	t.Run("insufficient role returns 403", func(t *testing.T) {
		e := echo.New()
		e.Use(echoMw.RequireRole(users.RoleAgent))
		e.GET("/protected", func(c echo.Context) error {
			return c.NoContent(http.StatusNoContent)
		})

		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		req = withClaims(req, &authdomain.Claims{
			UserID: "a5a8ee72-4db4-4f6c-b1a0-d6f4425908a8",
			Role:   users.RoleCustomer,
		})
		rec := httptest.NewRecorder()

		e.ServeHTTP(rec, req)

		require.Equal(t, http.StatusForbidden, rec.Code)
	})

	t.Run("same role is allowed", func(t *testing.T) {
		e := echo.New()
		e.Use(echoMw.RequireRole(users.RoleAgent))
		e.GET("/protected", func(c echo.Context) error {
			return c.NoContent(http.StatusNoContent)
		})

		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		req = withClaims(req, &authdomain.Claims{
			UserID: "7bb5f4a7-6ec6-4a97-b0d6-5af0609c8ec7",
			Role:   users.RoleAgent,
		})
		rec := httptest.NewRecorder()

		e.ServeHTTP(rec, req)

		require.Equal(t, http.StatusNoContent, rec.Code)
	})

	t.Run("higher role is allowed", func(t *testing.T) {
		e := echo.New()
		e.Use(echoMw.RequireRole(users.RoleAgent))
		e.GET("/protected", func(c echo.Context) error {
			return c.NoContent(http.StatusNoContent)
		})

		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		req = withClaims(req, &authdomain.Claims{
			UserID: "1260ef3a-cd5d-4f9c-a95e-b3f4d32f7af2",
			Role:   users.RoleAdmin,
		})
		rec := httptest.NewRecorder()

		e.ServeHTTP(rec, req)

		require.Equal(t, http.StatusNoContent, rec.Code)
	})
}

func TestIsOwnerOrRole(t *testing.T) {
	t.Run("owner is allowed", func(t *testing.T) {
		e := echo.New()
		req := withClaims(httptest.NewRequest(http.MethodGet, "/resource", nil), &authdomain.Claims{
			UserID: "056bb5d5-9d39-4aa6-9539-1e14cbf80a2b",
			Role:   users.RoleCustomer,
		})
		c := e.NewContext(req, httptest.NewRecorder())

		assert.True(t, echoMw.IsOwnerOrRole(c, "056bb5d5-9d39-4aa6-9539-1e14cbf80a2b", users.RoleAdmin))
	})

	t.Run("higher role is allowed for non-owner", func(t *testing.T) {
		e := echo.New()
		req := withClaims(httptest.NewRequest(http.MethodGet, "/resource", nil), &authdomain.Claims{
			UserID: "f108f7ad-dcf6-4bcd-a6eb-8f70ea5fd70a",
			Role:   users.RoleAdmin,
		})
		c := e.NewContext(req, httptest.NewRecorder())

		assert.True(t, echoMw.IsOwnerOrRole(c, "7f3b1336-3ff3-4e6a-b68d-f9e7a928ec8e", users.RoleAgent))
	})

	t.Run("non-owner with insufficient role is denied", func(t *testing.T) {
		e := echo.New()
		req := withClaims(httptest.NewRequest(http.MethodGet, "/resource", nil), &authdomain.Claims{
			UserID: "16cf0f26-bcf0-4655-ab35-e4ca919ad6e7",
			Role:   users.RoleCustomer,
		})
		c := e.NewContext(req, httptest.NewRecorder())

		assert.False(t, echoMw.IsOwnerOrRole(c, "fe538365-1a0f-4165-9da7-d6ab53ed1558", users.RoleAgent))
	})

	t.Run("missing claims is denied", func(t *testing.T) {
		e := echo.New()
		c := e.NewContext(httptest.NewRequest(http.MethodGet, "/resource", nil), httptest.NewRecorder())

		assert.False(t, echoMw.IsOwnerOrRole(c, "fdb9f95f-4ef4-4dcb-b9d8-020f467ee399", users.RoleCustomer))
	})
}

func withClaims(req *http.Request, claims *authdomain.Claims) *http.Request {
	ctx := context.WithValue(req.Context(), contextkeys.AuthClaimsCtxKey, claims)
	return req.WithContext(ctx)
}
