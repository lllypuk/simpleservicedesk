package auth_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"

	"simpleservicedesk/generated/openapi"

	"github.com/labstack/echo/v4"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

const testBypassHeaderKey = "X-Test-Bypass"

func (s *AuthSuite) TestRouteAccessControl() {
	s.Run("public ping works without auth", func() {
		req := httptest.NewRequest(http.MethodGet, "/ping", nil)
		req.Header.Set(testBypassHeaderKey, "true")
		rec := httptest.NewRecorder()

		s.HTTPServer.ServeHTTP(rec, req)

		s.Require().Equal(http.StatusOK, rec.Code)
	})

	s.Run("public login does not require auth", func() {
		req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBufferString(`{"email":`))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		req.Header.Set(testBypassHeaderKey, "true")
		rec := httptest.NewRecorder()

		s.HTTPServer.ServeHTTP(rec, req)

		s.Require().Equal(http.StatusBadRequest, rec.Code)
	})

	s.Run("protected route rejects missing token", func() {
		req := httptest.NewRequest(http.MethodGet, "/tickets", nil)
		req.Header.Set(testBypassHeaderKey, "true")
		rec := httptest.NewRecorder()

		s.HTTPServer.ServeHTTP(rec, req)

		s.Require().Equal(http.StatusUnauthorized, rec.Code)
	})

	s.Run("agent route rejects customer", func() {
		customerToken := s.createAndLoginUser("customer-user@example.com", openapi.Customer)

		req := httptest.NewRequest(http.MethodGet, "/users", nil)
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+customerToken)
		rec := httptest.NewRecorder()

		s.HTTPServer.ServeHTTP(rec, req)

		s.Require().Equal(http.StatusForbidden, rec.Code)
	})

	s.Run("agent route allows agent", func() {
		agentToken := s.createAndLoginUser("agent-user@example.com", openapi.Agent)

		req := httptest.NewRequest(http.MethodGet, "/users", nil)
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+agentToken)
		rec := httptest.NewRecorder()

		s.HTTPServer.ServeHTTP(rec, req)

		s.Require().Equal(http.StatusOK, rec.Code)
	})

	s.Run("admin route rejects non-admin", func() {
		agentToken := s.createAndLoginUser("agent-for-admin-check@example.com", openapi.Agent)

		createUserRequest := openapi.CreateUserRequest{
			Name:     "Denied User",
			Email:    "denied-user@example.com",
			Password: "password123",
		}
		reqBody, err := json.Marshal(createUserRequest)
		s.Require().NoError(err)

		req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBuffer(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		req.Header.Set(echo.HeaderAuthorization, "Bearer "+agentToken)
		rec := httptest.NewRecorder()

		s.HTTPServer.ServeHTTP(rec, req)

		s.Require().Equal(http.StatusForbidden, rec.Code)
	})
}

func (s *AuthSuite) createAndLoginUser(email string, role openapi.UserRole) string {
	createUserRequest := openapi.CreateUserRequest{
		Name:     "Route Access User",
		Email:    openapi_types.Email(email),
		Password: "password123",
	}
	createUserBody, err := json.Marshal(createUserRequest)
	s.Require().NoError(err)

	createReq := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBuffer(createUserBody))
	createReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	createRec := httptest.NewRecorder()

	s.HTTPServer.ServeHTTP(createRec, createReq)
	s.Require().Equal(http.StatusCreated, createRec.Code)

	if role != openapi.Customer {
		var createResp openapi.CreateUserResponse
		err = json.Unmarshal(createRec.Body.Bytes(), &createResp)
		s.Require().NoError(err)
		s.Require().NotNil(createResp.Id)

		updateRoleReq := openapi.UpdateUserRoleRequest{
			Role: role,
		}
		updateRoleBody, marshalErr := json.Marshal(updateRoleReq)
		s.Require().NoError(marshalErr)

		updateRoleRequest := httptest.NewRequest(
			http.MethodPatch,
			"/users/"+createResp.Id.String()+"/role",
			bytes.NewBuffer(updateRoleBody),
		)
		updateRoleRequest.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		updateRoleRec := httptest.NewRecorder()

		s.HTTPServer.ServeHTTP(updateRoleRec, updateRoleRequest)
		s.Require().Equal(http.StatusOK, updateRoleRec.Code)
	}

	loginRequest := openapi.LoginRequest{
		Email:    openapi_types.Email(email),
		Password: "password123",
	}
	loginBody, err := json.Marshal(loginRequest)
	s.Require().NoError(err)

	loginReq := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(loginBody))
	loginReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	loginRec := httptest.NewRecorder()

	s.HTTPServer.ServeHTTP(loginRec, loginReq)
	s.Require().Equal(http.StatusOK, loginRec.Code)

	var loginResp openapi.LoginResponse
	err = json.Unmarshal(loginRec.Body.Bytes(), &loginResp)
	s.Require().NoError(err)
	s.Require().NotEmpty(loginResp.Token)

	return loginResp.Token
}
