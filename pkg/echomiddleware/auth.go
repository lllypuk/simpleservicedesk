package echomiddleware

import (
	"context"
	"errors"
	"net/http"
	"strings"

	authdomain "simpleservicedesk/internal/domain/auth"
	"simpleservicedesk/pkg/contextkeys"

	"github.com/labstack/echo/v4"
)

const (
	authClaimsEchoKey = "auth_claims"
	bearerScheme      = "Bearer"
)

var errInvalidAuthorizationHeader = errors.New("invalid authorization header")

type TokenValidator interface {
	ValidateToken(ctx context.Context, tokenString string) (*authdomain.Claims, error)
}

func Auth(validator TokenValidator) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			tokenString, err := extractBearerToken(c.Request().Header.Get(echo.HeaderAuthorization))
			if err != nil {
				return c.NoContent(http.StatusUnauthorized)
			}

			if validator == nil {
				return c.NoContent(http.StatusUnauthorized)
			}

			claims, err := validator.ValidateToken(c.Request().Context(), tokenString)
			if err != nil || claims == nil {
				return c.NoContent(http.StatusUnauthorized)
			}

			ctx := context.WithValue(c.Request().Context(), contextkeys.AuthClaimsCtxKey, claims)
			c.SetRequest(c.Request().WithContext(ctx))
			c.Set(authClaimsEchoKey, claims)

			return next(c)
		}
	}
}

func GetAuthClaims(c echo.Context) (*authdomain.Claims, bool) {
	if claims, ok := c.Get(authClaimsEchoKey).(*authdomain.Claims); ok && claims != nil {
		return claims, true
	}

	return GetAuthClaimsFromContext(c.Request().Context())
}

func GetAuthClaimsFromContext(ctx context.Context) (*authdomain.Claims, bool) {
	claims, ok := ctx.Value(contextkeys.AuthClaimsCtxKey).(*authdomain.Claims)
	if !ok || claims == nil {
		return nil, false
	}

	return claims, true
}

func extractBearerToken(authorization string) (string, error) {
	scheme, token, found := strings.Cut(strings.TrimSpace(authorization), " ")
	if !found {
		return "", errInvalidAuthorizationHeader
	}
	if !strings.EqualFold(scheme, bearerScheme) {
		return "", errInvalidAuthorizationHeader
	}
	token = strings.TrimSpace(token)
	if token == "" {
		return "", errInvalidAuthorizationHeader
	}

	return token, nil
}
