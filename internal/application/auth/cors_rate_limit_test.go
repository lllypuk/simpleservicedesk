package auth_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"

	"simpleservicedesk/generated/openapi"
	"simpleservicedesk/internal/application"

	"github.com/labstack/echo/v4"
)

const (
	testSingleRequestPerSecond = 1
	testHighRequestPerSecond   = 1000
	testLoginBurstLimit        = 5
	testLoginRetryAfterSeconds = "12"
	testPingPath               = "/ping"
	testLoginPath              = "/login"
	testTicketsPath            = "/tickets"
)

func (s *AuthSuite) TestCORSPreflightIncludesExpectedHeaders() {
	server := s.setupServerWithGlobalRateLimit(testHighRequestPerSecond)

	req := httptest.NewRequest(http.MethodOptions, testTicketsPath, nil)
	req.Header.Set(echo.HeaderOrigin, "https://frontend.example.com")
	req.Header.Set(echo.HeaderAccessControlRequestMethod, http.MethodPost)
	req.Header.Set(echo.HeaderAccessControlRequestHeaders, echo.HeaderContentType+","+echo.HeaderAuthorization)
	rec := httptest.NewRecorder()

	server.ServeHTTP(rec, req)

	s.Require().Equal(http.StatusNoContent, rec.Code)
	s.Require().Equal("*", rec.Header().Get(echo.HeaderAccessControlAllowOrigin))

	allowedMethods := rec.Header().Get(echo.HeaderAccessControlAllowMethods)
	s.Require().Contains(allowedMethods, http.MethodGet)
	s.Require().Contains(allowedMethods, http.MethodPost)
	s.Require().Contains(allowedMethods, http.MethodPut)
	s.Require().Contains(allowedMethods, http.MethodDelete)
	s.Require().Contains(allowedMethods, http.MethodOptions)

	allowedHeaders := strings.ToLower(rec.Header().Get(echo.HeaderAccessControlAllowHeaders))
	s.Require().Contains(allowedHeaders, strings.ToLower(echo.HeaderContentType))
	s.Require().Contains(allowedHeaders, strings.ToLower(echo.HeaderAuthorization))
}

func (s *AuthSuite) TestGlobalRateLimiterReturnsTooManyRequestsWhenExceeded() {
	server := s.setupServerWithGlobalRateLimit(testSingleRequestPerSecond)

	firstReq := httptest.NewRequest(http.MethodGet, testPingPath, nil)
	firstReq.RemoteAddr = "203.0.113.10:12345"
	firstRec := httptest.NewRecorder()
	server.ServeHTTP(firstRec, firstReq)
	s.Require().Equal(http.StatusOK, firstRec.Code)

	secondReq := httptest.NewRequest(http.MethodGet, testPingPath, nil)
	secondReq.RemoteAddr = "203.0.113.10:12345"
	secondRec := httptest.NewRecorder()
	server.ServeHTTP(secondRec, secondReq)

	s.Require().Equal(http.StatusTooManyRequests, secondRec.Code)
	s.Require().Equal("1", secondRec.Header().Get(echo.HeaderRetryAfter))

	var response openapi.ErrorResponse
	err := json.Unmarshal(secondRec.Body.Bytes(), &response)
	s.Require().NoError(err)
	s.Require().NotNil(response.Message)
	s.Require().Equal("rate limit exceeded", *response.Message)
}

func (s *AuthSuite) TestLoginEndpointHasStricterRateLimit() {
	server := s.setupServerWithGlobalRateLimit(testHighRequestPerSecond)

	loginPayload, err := json.Marshal(openapi.LoginRequest{
		Email:    "test-auth-admin@example.com",
		Password: "password123",
	})
	s.Require().NoError(err)

	for range testLoginBurstLimit {
		req := httptest.NewRequest(http.MethodPost, testLoginPath, bytes.NewBuffer(loginPayload))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		req.RemoteAddr = "198.51.100.12:54321"
		rec := httptest.NewRecorder()
		server.ServeHTTP(rec, req)
		s.Require().Equal(http.StatusOK, rec.Code)
	}

	exceededReq := httptest.NewRequest(http.MethodPost, testLoginPath, bytes.NewBuffer(loginPayload))
	exceededReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	exceededReq.RemoteAddr = "198.51.100.12:54321"
	exceededRec := httptest.NewRecorder()
	server.ServeHTTP(exceededRec, exceededReq)

	s.Require().Equal(http.StatusTooManyRequests, exceededRec.Code)
	s.Require().Equal(testLoginRetryAfterSeconds, exceededRec.Header().Get(echo.HeaderRetryAfter))
}

func (s *AuthSuite) setupServerWithGlobalRateLimit(requestsPerSecond int) *echo.Echo {
	server, err := application.SetupHTTPServer(
		s.UsersRepo,
		s.TicketsRepo,
		s.OrganizationsRepo,
		s.CategoriesRepo,
		"test-jwt-signing-key",
		time.Hour,
		[]string{"*"},
		requestsPerSecond,
	)
	s.Require().NoError(err)

	return server
}
