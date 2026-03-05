//go:build integration
// +build integration

package api_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"simpleservicedesk/generated/openapi"
	userdomain "simpleservicedesk/internal/domain/users"
	"simpleservicedesk/test/integration/shared"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	openapi_types "github.com/oapi-codegen/runtime/types"
	"github.com/stretchr/testify/suite"
)

type AuthAPITestSuite struct {
	shared.IntegrationSuite
}

func TestAuthAPI(t *testing.T) {
	suite.Run(t, new(AuthAPITestSuite))
}

func (s *AuthAPITestSuite) TestPostLoginIntegration() {
	createdUser := s.MustCreateTestUser(userdomain.RoleCustomer)

	tests := []struct {
		name           string
		email          string
		passphrase     string
		expectedStatus int
		expectToken    bool
	}{
		{
			name:           "success",
			email:          createdUser.Email,
			passphrase:     createdUser.Passphrase,
			expectedStatus: http.StatusOK,
			expectToken:    true,
		},
		{
			name:           "wrong password",
			email:          createdUser.Email,
			passphrase:     "wrong-passphrase",
			expectedStatus: http.StatusUnauthorized,
			expectToken:    false,
		},
		{
			name:           "nonexistent user",
			email:          "missing.user@example.com",
			passphrase:     "passphrase-123",
			expectedStatus: http.StatusUnauthorized,
			expectToken:    false,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			token, rec := s.LoginAndGetToken(tt.email, tt.passphrase)
			s.Equal(tt.expectedStatus, rec.Code, "response: %s", rec.Body.String())

			if tt.expectToken {
				s.NotEmpty(token)
				return
			}

			s.Empty(token)
			var errorResp openapi.ErrorResponse
			err := json.Unmarshal(rec.Body.Bytes(), &errorResp)
			s.Require().NoError(err)
			s.NotNil(errorResp.Message)
			s.NotEmpty(*errorResp.Message)
		})
	}
}

func (s *AuthAPITestSuite) TestProtectedEndpointWithoutToken() {
	req := httptest.NewRequest(http.MethodGet, "/tickets", nil)
	rec := httptest.NewRecorder()

	s.HTTPServer.ServeHTTP(rec, req)
	s.Equal(http.StatusUnauthorized, rec.Code)
}

func (s *AuthAPITestSuite) TestProtectedEndpointWithInsufficientRole() {
	customer := s.MustCreateAndLoginTestUser(userdomain.RoleCustomer)

	createUserReq := openapi.CreateUserRequest{
		Name:     "Denied User",
		Email:    openapi_types.Email("denied.user@example.com"),
		Password: "password123",
	}
	reqBody, err := json.Marshal(createUserReq)
	s.Require().NoError(err)

	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req.Header.Set(echo.HeaderAuthorization, "Bearer "+customer.Token)
	rec := httptest.NewRecorder()

	s.HTTPServer.ServeHTTP(rec, req)
	s.Equal(http.StatusForbidden, rec.Code)
}

func (s *AuthAPITestSuite) TestRoleProtectedEndpoints() {
	customer := s.MustCreateAndLoginTestUser(userdomain.RoleCustomer)
	agent := s.MustCreateAndLoginTestUser(userdomain.RoleAgent)
	targetUserID := uuid.New()
	targetTicketID := uuid.New()

	statusBody, err := json.Marshal(openapi.UpdateTicketStatusRequest{Status: openapi.InProgress})
	s.Require().NoError(err)
	roleBody, err := json.Marshal(openapi.UpdateUserRoleRequest{Role: openapi.Agent})
	s.Require().NoError(err)

	tests := []struct {
		name           string
		method         string
		path           string
		body           []byte
		token          string
		expectedStatus int
	}{
		{
			name:           "customer cannot update ticket status",
			method:         http.MethodPatch,
			path:           "/tickets/" + targetTicketID.String() + "/status",
			body:           statusBody,
			token:          customer.Token,
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "agent cannot change user role",
			method:         http.MethodPatch,
			path:           "/users/" + targetUserID.String() + "/role",
			body:           roleBody,
			token:          agent.Token,
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "missing token on admin endpoint returns unauthorized",
			method:         http.MethodDelete,
			path:           "/users/" + targetUserID.String(),
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			req := httptest.NewRequest(tt.method, tt.path, bytes.NewReader(tt.body))
			if len(tt.body) > 0 {
				req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			}
			if tt.token != "" {
				req.Header.Set(echo.HeaderAuthorization, "Bearer "+tt.token)
			}
			rec := httptest.NewRecorder()

			s.HTTPServer.ServeHTTP(rec, req)
			s.Equal(tt.expectedStatus, rec.Code, "response: %s", rec.Body.String())
		})
	}
}

func (s *AuthAPITestSuite) TestCustomerSeesOnlyOwnTickets() {
	customerA := s.MustCreateAndLoginTestUser(userdomain.RoleCustomer)
	customerB := s.MustCreateAndLoginTestUser(userdomain.RoleCustomer)
	organizationID := uuid.New()

	createTicket := func(title string, authorID uuid.UUID) {
		ticketReq := openapi.CreateTicketRequest{
			Title:          title,
			Description:    "auth visibility integration test",
			Priority:       openapi.TicketPriority("normal"),
			OrganizationId: organizationID,
			AuthorId:       authorID,
		}

		reqBody, err := json.Marshal(ticketReq)
		s.Require().NoError(err)

		req := httptest.NewRequest(http.MethodPost, "/tickets", bytes.NewReader(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		s.ServeAuthenticatedHTTP(rec, req)
		s.Require().Equal(http.StatusCreated, rec.Code, "response: %s", rec.Body.String())
	}

	createTicket("ticket-a-1", customerA.UserID)
	createTicket("ticket-a-2", customerA.UserID)
	createTicket("ticket-b-1", customerB.UserID)

	req := httptest.NewRequest(http.MethodGet, "/tickets?author_id="+customerB.UserID.String(), nil)
	req.Header.Set(echo.HeaderAuthorization, "Bearer "+customerA.Token)
	rec := httptest.NewRecorder()

	s.ServeAuthenticatedHTTP(rec, req)
	s.Require().Equal(http.StatusOK, rec.Code, "response: %s", rec.Body.String())

	var listResp openapi.ListTicketsResponse
	err := json.Unmarshal(rec.Body.Bytes(), &listResp)
	s.Require().NoError(err)
	s.Require().NotNil(listResp.Tickets)
	s.Len(*listResp.Tickets, 2)

	for _, ticket := range *listResp.Tickets {
		s.Require().NotNil(ticket.AuthorId)
		s.Equal(customerA.UserID, *ticket.AuthorId)
	}
}
