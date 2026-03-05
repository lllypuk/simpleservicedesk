package auth_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"

	"simpleservicedesk/generated/openapi"

	"github.com/labstack/echo/v4"
)

func (s *AuthSuite) TestRequestValidationMiddleware() {
	s.Run("invalid API query parameter returns structured error response", func() {
		req := httptest.NewRequest(http.MethodGet, "/users?page=invalid-page", nil)
		req.Header.Set(testBypassHeaderKey, "true")
		rec := httptest.NewRecorder()

		s.HTTPServer.ServeHTTP(rec, req)

		s.Require().Equal(http.StatusBadRequest, rec.Code)

		var errorResp openapi.ErrorResponse
		err := json.Unmarshal(rec.Body.Bytes(), &errorResp)
		s.Require().NoError(err)
		s.Require().NotNil(errorResp.Message)
		s.Require().NotEmpty(strings.TrimSpace(*errorResp.Message))
	})

	s.Run("login is skipped by OpenAPI validation middleware", func() {
		loginBody, err := json.Marshal(map[string]string{"email": "agent@example.com"})
		s.Require().NoError(err)

		req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(loginBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		req.Header.Set(testBypassHeaderKey, "true")
		rec := httptest.NewRecorder()

		s.HTTPServer.ServeHTTP(rec, req)

		s.Require().Equal(http.StatusBadRequest, rec.Code)

		var errorResp openapi.ErrorResponse
		err = json.Unmarshal(rec.Body.Bytes(), &errorResp)
		s.Require().NoError(err)
		s.Require().NotNil(errorResp.Message)
		s.Require().Equal("email and password are required", *errorResp.Message)
	})
}
