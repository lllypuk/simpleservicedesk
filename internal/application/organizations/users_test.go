package organizations_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"

	"simpleservicedesk/generated/openapi"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func (s *OrganizationsSuite) TestGetOrganizationUsers() {
	s.Run("Get users from organization - not implemented", func() {
		domain := "example.com"
		orgReq := openapi.CreateOrganizationRequest{
			Name:   "Example Organization",
			Domain: &domain,
		}
		reqBody, _ := json.Marshal(orgReq)

		req := httptest.NewRequest(http.MethodPost, "/organizations", bytes.NewBuffer(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		s.HTTPServer.ServeHTTP(rec, req)

		s.Require().Equal(http.StatusCreated, rec.Code)
		var createResp openapi.CreateOrganizationResponse
		err := json.Unmarshal(rec.Body.Bytes(), &createResp)
		s.Require().NoError(err)

		req = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/organizations/%s/users", createResp.Id.String()), nil)
		rec = httptest.NewRecorder()
		s.HTTPServer.ServeHTTP(rec, req)

		s.Require().Equal(http.StatusNotImplemented, rec.Code)
		var errorResp openapi.ErrorResponse
		err = json.Unmarshal(rec.Body.Bytes(), &errorResp)
		s.Require().NoError(err)
		s.Require().Contains(*errorResp.Message, "not implemented")
	})

	s.Run("Get users from non-existent organization - not implemented", func() {
		nonExistentID := uuid.New()
		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/organizations/%s/users", nonExistentID.String()), nil)
		rec := httptest.NewRecorder()
		s.HTTPServer.ServeHTTP(rec, req)

		s.Require().Equal(http.StatusNotImplemented, rec.Code)
	})

	s.Run("Get users with invalid organization ID - not implemented", func() {
		req := httptest.NewRequest(http.MethodGet, "/organizations/invalid-uuid/users", nil)
		rec := httptest.NewRecorder()
		s.HTTPServer.ServeHTTP(rec, req)

		s.Require().Equal(http.StatusBadRequest, rec.Code)
	})
}
