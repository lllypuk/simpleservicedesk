//go:build integration
// +build integration

package api_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"simpleservicedesk/generated/openapi"
	authdomain "simpleservicedesk/internal/domain/auth"
	userdomain "simpleservicedesk/internal/domain/users"
	"simpleservicedesk/test/integration/shared"

	"github.com/golang-jwt/jwt/v5"
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

func (s *AuthAPITestSuite) TestInvalidTokenForms() {
	tests := []struct {
		name  string
		token string
	}{
		{
			name:  "malformed token",
			token: "not-a-jwt",
		},
		{
			name:  "wrong signature",
			token: signIntegrationToken(s.T(), "wrong-signing-key", time.Now().UTC().Add(time.Hour)),
		},
		{
			name:  "expired token",
			token: signIntegrationToken(s.T(), "integration-test-jwt-signing-key", time.Now().UTC().Add(-time.Minute)),
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			req := httptest.NewRequest(http.MethodGet, "/tickets", nil)
			req.Header.Set(echo.HeaderAuthorization, "Bearer "+tt.token)
			rec := httptest.NewRecorder()

			s.HTTPServer.ServeHTTP(rec, req)
			s.Equal(http.StatusUnauthorized, rec.Code, "response: %s", rec.Body.String())
		})
	}
}

func (s *AuthAPITestSuite) TestCustomerCannotImpersonateOrEscalateCommentVisibility() {
	customerA := s.MustCreateAndLoginTestUser(userdomain.RoleCustomer)
	customerB := s.MustCreateAndLoginTestUser(userdomain.RoleCustomer)

	createTicketReq := openapi.CreateTicketRequest{
		Title:          "Customer-authored ticket",
		Description:    "security test ticket",
		Priority:       openapi.Normal,
		OrganizationId: uuid.New(),
		AuthorId:       customerB.UserID,
	}
	createTicketBody, err := json.Marshal(createTicketReq)
	s.Require().NoError(err)

	createTicketHTTPReq := httptest.NewRequest(http.MethodPost, "/tickets", bytes.NewReader(createTicketBody))
	createTicketHTTPReq.Header.Set(echo.HeaderAuthorization, "Bearer "+customerA.Token)
	createTicketHTTPReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	createTicketRec := httptest.NewRecorder()

	s.HTTPServer.ServeHTTP(createTicketRec, createTicketHTTPReq)
	s.Require().Equal(http.StatusCreated, createTicketRec.Code, "response: %s", createTicketRec.Body.String())

	var createdTicket openapi.GetTicketResponse
	err = json.Unmarshal(createTicketRec.Body.Bytes(), &createdTicket)
	s.Require().NoError(err)
	s.Require().NotNil(createdTicket.AuthorId)
	s.Require().Equal(customerA.UserID, *createdTicket.AuthorId)
	s.Require().NotNil(createdTicket.Id)

	createCommentReq := openapi.CreateCommentRequest{
		AuthorId: customerB.UserID,
		Content:  "customer comment",
	}
	createCommentBody, err := json.Marshal(createCommentReq)
	s.Require().NoError(err)

	createCommentHTTPReq := httptest.NewRequest(
		http.MethodPost,
		"/tickets/"+createdTicket.Id.String()+"/comments",
		bytes.NewReader(createCommentBody),
	)
	createCommentHTTPReq.Header.Set(echo.HeaderAuthorization, "Bearer "+customerA.Token)
	createCommentHTTPReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	createCommentRec := httptest.NewRecorder()

	s.HTTPServer.ServeHTTP(createCommentRec, createCommentHTTPReq)
	s.Require().Equal(http.StatusCreated, createCommentRec.Code, "response: %s", createCommentRec.Body.String())

	var createdComment openapi.TicketComment
	err = json.Unmarshal(createCommentRec.Body.Bytes(), &createdComment)
	s.Require().NoError(err)
	s.Require().NotNil(createdComment.AuthorId)
	s.Equal(customerA.UserID, *createdComment.AuthorId)

	isInternal := true
	internalCommentReq := openapi.CreateCommentRequest{
		AuthorId:   customerA.UserID,
		Content:    "internal comment attempt",
		IsInternal: &isInternal,
	}
	internalCommentBody, err := json.Marshal(internalCommentReq)
	s.Require().NoError(err)

	internalCommentHTTPReq := httptest.NewRequest(
		http.MethodPost,
		"/tickets/"+createdTicket.Id.String()+"/comments",
		bytes.NewReader(internalCommentBody),
	)
	internalCommentHTTPReq.Header.Set(echo.HeaderAuthorization, "Bearer "+customerA.Token)
	internalCommentHTTPReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	internalCommentRec := httptest.NewRecorder()

	s.HTTPServer.ServeHTTP(internalCommentRec, internalCommentHTTPReq)
	s.Equal(http.StatusForbidden, internalCommentRec.Code, "response: %s", internalCommentRec.Body.String())

	getInternalReq := httptest.NewRequest(
		http.MethodGet,
		"/tickets/"+createdTicket.Id.String()+"/comments?include_internal=true",
		nil,
	)
	getInternalReq.Header.Set(echo.HeaderAuthorization, "Bearer "+customerA.Token)
	getInternalRec := httptest.NewRecorder()

	s.HTTPServer.ServeHTTP(getInternalRec, getInternalReq)
	s.Equal(http.StatusForbidden, getInternalRec.Code, "response: %s", getInternalRec.Body.String())
}

func (s *AuthAPITestSuite) TestUserProfileAccessControl() {
	customerA := s.MustCreateAndLoginTestUser(userdomain.RoleCustomer)
	customerB := s.MustCreateAndLoginTestUser(userdomain.RoleCustomer)

	getOtherReq := httptest.NewRequest(http.MethodGet, "/users/"+customerB.UserID.String(), nil)
	getOtherReq.Header.Set(echo.HeaderAuthorization, "Bearer "+customerA.Token)
	getOtherRec := httptest.NewRecorder()
	s.HTTPServer.ServeHTTP(getOtherRec, getOtherReq)
	s.Equal(http.StatusForbidden, getOtherRec.Code, "response: %s", getOtherRec.Body.String())

	updateOtherBody, err := json.Marshal(openapi.UpdateUserRequest{
		Name: func() *string {
			name := "Updated by other customer"
			return &name
		}(),
	})
	s.Require().NoError(err)

	updateOtherReq := httptest.NewRequest(
		http.MethodPut,
		"/users/"+customerB.UserID.String(),
		bytes.NewReader(updateOtherBody),
	)
	updateOtherReq.Header.Set(echo.HeaderAuthorization, "Bearer "+customerA.Token)
	updateOtherReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	updateOtherRec := httptest.NewRecorder()
	s.HTTPServer.ServeHTTP(updateOtherRec, updateOtherReq)
	s.Equal(http.StatusForbidden, updateOtherRec.Code, "response: %s", updateOtherRec.Body.String())

	disableSelfBody, err := json.Marshal(openapi.UpdateUserRequest{
		IsActive: func() *bool {
			isActive := false
			return &isActive
		}(),
	})
	s.Require().NoError(err)

	disableSelfReq := httptest.NewRequest(
		http.MethodPut,
		"/users/"+customerA.UserID.String(),
		bytes.NewReader(disableSelfBody),
	)
	disableSelfReq.Header.Set(echo.HeaderAuthorization, "Bearer "+customerA.Token)
	disableSelfReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	disableSelfRec := httptest.NewRecorder()
	s.HTTPServer.ServeHTTP(disableSelfRec, disableSelfReq)
	s.Equal(http.StatusForbidden, disableSelfRec.Code, "response: %s", disableSelfRec.Body.String())

	newName := "Customer Updated Name"
	newEmail := openapi_types.Email("customer.updated@example.com")
	updateSelfBody, err := json.Marshal(openapi.UpdateUserRequest{
		Name:  &newName,
		Email: &newEmail,
	})
	s.Require().NoError(err)

	updateSelfReq := httptest.NewRequest(
		http.MethodPut,
		"/users/"+customerA.UserID.String(),
		bytes.NewReader(updateSelfBody),
	)
	updateSelfReq.Header.Set(echo.HeaderAuthorization, "Bearer "+customerA.Token)
	updateSelfReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	updateSelfRec := httptest.NewRecorder()
	s.HTTPServer.ServeHTTP(updateSelfRec, updateSelfReq)
	s.Equal(http.StatusOK, updateSelfRec.Code, "response: %s", updateSelfRec.Body.String())
}

func (s *AuthAPITestSuite) TestCustomerCannotAccessOtherCustomerTicketByID() {
	customerA := s.MustCreateAndLoginTestUser(userdomain.RoleCustomer)
	customerB := s.MustCreateAndLoginTestUser(userdomain.RoleCustomer)

	createTicketReq := openapi.CreateTicketRequest{
		Title:          "Customer A ticket",
		Description:    "ownership test ticket",
		Priority:       openapi.Normal,
		OrganizationId: uuid.New(),
		AuthorId:       customerA.UserID,
	}
	createTicketBody, err := json.Marshal(createTicketReq)
	s.Require().NoError(err)

	createTicketHTTPReq := httptest.NewRequest(http.MethodPost, "/tickets", bytes.NewReader(createTicketBody))
	createTicketHTTPReq.Header.Set(echo.HeaderAuthorization, "Bearer "+customerA.Token)
	createTicketHTTPReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	createTicketRec := httptest.NewRecorder()

	s.HTTPServer.ServeHTTP(createTicketRec, createTicketHTTPReq)
	s.Require().Equal(http.StatusCreated, createTicketRec.Code, "response: %s", createTicketRec.Body.String())

	var createdTicket openapi.GetTicketResponse
	err = json.Unmarshal(createTicketRec.Body.Bytes(), &createdTicket)
	s.Require().NoError(err)
	s.Require().NotNil(createdTicket.Id)

	getReq := httptest.NewRequest(http.MethodGet, "/tickets/"+createdTicket.Id.String(), nil)
	getReq.Header.Set(echo.HeaderAuthorization, "Bearer "+customerB.Token)
	getRec := httptest.NewRecorder()
	s.HTTPServer.ServeHTTP(getRec, getReq)
	s.Equal(http.StatusForbidden, getRec.Code, "response: %s", getRec.Body.String())

	updateBody, err := json.Marshal(openapi.UpdateTicketRequest{
		Title: func() *string {
			title := "Unauthorized update"
			return &title
		}(),
	})
	s.Require().NoError(err)

	updateReq := httptest.NewRequest(http.MethodPut, "/tickets/"+createdTicket.Id.String(), bytes.NewReader(updateBody))
	updateReq.Header.Set(echo.HeaderAuthorization, "Bearer "+customerB.Token)
	updateReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	updateRec := httptest.NewRecorder()
	s.HTTPServer.ServeHTTP(updateRec, updateReq)
	s.Equal(http.StatusForbidden, updateRec.Code, "response: %s", updateRec.Body.String())

	commentBody, err := json.Marshal(openapi.CreateCommentRequest{
		AuthorId: customerB.UserID,
		Content:  "unauthorized comment",
	})
	s.Require().NoError(err)

	commentReq := httptest.NewRequest(
		http.MethodPost,
		"/tickets/"+createdTicket.Id.String()+"/comments",
		bytes.NewReader(commentBody),
	)
	commentReq.Header.Set(echo.HeaderAuthorization, "Bearer "+customerB.Token)
	commentReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	commentRec := httptest.NewRecorder()
	s.HTTPServer.ServeHTTP(commentRec, commentReq)
	s.Equal(http.StatusForbidden, commentRec.Code, "response: %s", commentRec.Body.String())

	deleteReq := httptest.NewRequest(http.MethodDelete, "/tickets/"+createdTicket.Id.String(), nil)
	deleteReq.Header.Set(echo.HeaderAuthorization, "Bearer "+customerB.Token)
	deleteRec := httptest.NewRecorder()
	s.HTTPServer.ServeHTTP(deleteRec, deleteReq)
	s.Equal(http.StatusForbidden, deleteRec.Code, "response: %s", deleteRec.Body.String())
}

func signIntegrationToken(t testing.TB, key string, exp time.Time) string {
	t.Helper()

	userID := uuid.New().String()
	claims := authdomain.Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
			ExpiresAt: jwt.NewNumericDate(exp),
		},
		UserID: userID,
		Role:   userdomain.RoleCustomer,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(key))
	if err != nil {
		t.Fatalf("failed to sign token: %v", err)
	}

	return tokenString
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
