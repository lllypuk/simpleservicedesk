package echomiddleware_test

import (
	"context"
	"errors"
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

type stubTokenValidator struct {
	claims     *authdomain.Claims
	err        error
	calls      int
	lastCtx    context.Context
	lastToken  string
	returnsNil bool
}

func (s *stubTokenValidator) ValidateToken(ctx context.Context, tokenString string) (*authdomain.Claims, error) {
	s.calls++
	s.lastCtx = ctx
	s.lastToken = tokenString
	if s.returnsNil {
		return nil, s.err
	}

	return s.claims, s.err
}

func TestAuth(t *testing.T) {
	t.Run("missing authorization header returns 401", func(t *testing.T) {
		validator := &stubTokenValidator{}

		e := echo.New()
		e.Use(echoMw.Auth(validator))
		e.GET("/protected", func(c echo.Context) error {
			return c.NoContent(http.StatusOK)
		})

		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		rec := httptest.NewRecorder()

		e.ServeHTTP(rec, req)

		require.Equal(t, http.StatusUnauthorized, rec.Code)
		require.Equal(t, 0, validator.calls)
	})

	t.Run("invalid authorization header returns 401", func(t *testing.T) {
		validator := &stubTokenValidator{}

		e := echo.New()
		e.Use(echoMw.Auth(validator))
		e.GET("/protected", func(c echo.Context) error {
			return c.NoContent(http.StatusOK)
		})

		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		req.Header.Set(echo.HeaderAuthorization, "Token abc")
		rec := httptest.NewRecorder()

		e.ServeHTTP(rec, req)

		require.Equal(t, http.StatusUnauthorized, rec.Code)
		require.Equal(t, 0, validator.calls)
	})

	t.Run("invalid token returns 401", func(t *testing.T) {
		validator := &stubTokenValidator{
			err: errors.New("bad token"),
		}

		e := echo.New()
		e.Use(echoMw.Auth(validator))
		e.GET("/protected", func(c echo.Context) error {
			return c.NoContent(http.StatusOK)
		})

		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		req.Header.Set(echo.HeaderAuthorization, "Bearer token-123")
		rec := httptest.NewRecorder()

		e.ServeHTTP(rec, req)

		require.Equal(t, http.StatusUnauthorized, rec.Code)
		require.Equal(t, 1, validator.calls)
		require.Equal(t, "token-123", validator.lastToken)
	})

	t.Run("valid token injects claims into context", func(t *testing.T) {
		expectedClaims := &authdomain.Claims{
			UserID: "835dce37-aefd-4e24-8cc0-a50e59f07ae2",
			Role:   users.RoleAdmin,
		}
		validator := &stubTokenValidator{
			claims: expectedClaims,
		}

		e := echo.New()
		e.Use(echoMw.Auth(validator))
		e.GET("/protected", func(c echo.Context) error {
			claimsFromEcho, ok := echoMw.GetAuthClaims(c)
			require.True(t, ok)
			require.Equal(t, expectedClaims, claimsFromEcho)

			claimsFromRequestContext, ok := c.Request().Context().Value(contextkeys.AuthClaimsCtxKey).(*authdomain.Claims)
			require.True(t, ok)
			require.Equal(t, expectedClaims, claimsFromRequestContext)

			return c.NoContent(http.StatusNoContent)
		})

		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		req.Header.Set(echo.HeaderAuthorization, "Bearer token-123")
		rec := httptest.NewRecorder()

		e.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusNoContent, rec.Code)
		assert.Equal(t, 1, validator.calls)
		require.NotNil(t, validator.lastCtx)
		assert.Equal(t, "token-123", validator.lastToken)
	})

	t.Run("helper reads claims from request context fallback", func(t *testing.T) {
		expectedClaims := &authdomain.Claims{
			UserID: "cbb8adcf-3f9f-4867-b70b-76e6803a1f2c",
			Role:   users.RoleAgent,
		}
		e := echo.New()
		req := httptest.NewRequest(http.MethodGet, "/fallback", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		reqCtx := context.WithValue(c.Request().Context(), contextkeys.AuthClaimsCtxKey, expectedClaims)
		c.SetRequest(c.Request().WithContext(reqCtx))

		actualClaims, ok := echoMw.GetAuthClaims(c)
		require.True(t, ok)
		require.Equal(t, expectedClaims, actualClaims)
	})
}
