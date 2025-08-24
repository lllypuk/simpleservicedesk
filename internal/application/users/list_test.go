package users_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"

	"simpleservicedesk/generated/openapi"

	"github.com/labstack/echo/v4"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

func (s *UsersSuite) TestListUsers() {
	// First create an organization
	domain := "test.com"
	orgReq := openapi.CreateOrganizationRequest{
		Name:   "Test Organization",
		Domain: &domain,
	}
	orgReqBody, _ := json.Marshal(orgReq)

	orgRequest := httptest.NewRequest(http.MethodPost, "/organizations", bytes.NewBuffer(orgReqBody))
	orgRequest.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	orgRec := httptest.NewRecorder()
	s.HTTPServer.ServeHTTP(orgRec, orgRequest)

	s.Require().Equal(http.StatusCreated, orgRec.Code)
	var orgResp openapi.CreateOrganizationResponse
	err := json.Unmarshal(orgRec.Body.Bytes(), &orgResp)
	s.Require().NoError(err)

	// Create multiple users
	for i := 1; i <= 3; i++ {
		email := openapi_types.Email("user" + string(rune('0'+i)) + "@test.com")
		userReq := openapi.CreateUserRequest{
			Name:     "Test User " + string(rune('0'+i)),
			Email:    email,
			Password: "password123",
		}
		reqBody, _ := json.Marshal(userReq)

		req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBuffer(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		s.HTTPServer.ServeHTTP(rec, req)

		s.Require().Equal(http.StatusCreated, rec.Code)
	}

	s.Run("List all users", func() {
		getReq := httptest.NewRequest(http.MethodGet, "/users", nil)
		getRec := httptest.NewRecorder()
		s.HTTPServer.ServeHTTP(getRec, getReq)

		s.Require().Equal(http.StatusOK, getRec.Code)
		var listResp openapi.ListUsersResponse
		listUnmarshalErr := json.Unmarshal(getRec.Body.Bytes(), &listResp)
		s.Require().NoError(listUnmarshalErr)
		s.Require().NotNil(listResp.Users)
		s.Require().GreaterOrEqual(len(*listResp.Users), 3)
	})

	s.Run("List users with email filter", func() {
		getReq := httptest.NewRequest(http.MethodGet, "/users?email=user1", nil)
		getRec := httptest.NewRecorder()
		s.HTTPServer.ServeHTTP(getRec, getReq)

		s.Require().Equal(http.StatusOK, getRec.Code)
		var listResp openapi.ListUsersResponse
		filterUnmarshalErr := json.Unmarshal(getRec.Body.Bytes(), &listResp)
		s.Require().NoError(filterUnmarshalErr)
		s.Require().NotNil(listResp.Users)
	})

	s.Run("List users with role filter", func() {
		getReq := httptest.NewRequest(http.MethodGet, "/users?role=customer", nil)
		getRec := httptest.NewRecorder()
		s.HTTPServer.ServeHTTP(getRec, getReq)

		s.Require().Equal(http.StatusOK, getRec.Code)
		var listResp openapi.ListUsersResponse
		roleFilterUnmarshalErr := json.Unmarshal(getRec.Body.Bytes(), &listResp)
		s.Require().NoError(roleFilterUnmarshalErr)
		s.Require().NotNil(listResp.Users)
	})

	s.Run("List users with pagination", func() {
		getReq := httptest.NewRequest(http.MethodGet, "/users?page=1&limit=2", nil)
		getRec := httptest.NewRecorder()
		s.HTTPServer.ServeHTTP(getRec, getReq)

		s.Require().Equal(http.StatusOK, getRec.Code)
		var listResp openapi.ListUsersResponse
		pageUnmarshalErr := json.Unmarshal(getRec.Body.Bytes(), &listResp)
		s.Require().NoError(pageUnmarshalErr)
		s.Require().NotNil(listResp.Users)
		s.Require().NotNil(listResp.Pagination)
		s.Require().NotNil(listResp.Pagination.Page)
		s.Require().Equal(1, *listResp.Pagination.Page)
		s.Require().NotNil(listResp.Pagination.Limit)
		s.Require().Equal(2, *listResp.Pagination.Limit)
	})
}
