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

func (s *OrganizationsSuite) TestGetOrganizationTickets() {
	s.Run("Get tickets from organization - not implemented", func() {
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

		req = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/organizations/%s/tickets", createResp.Id.String()), nil)
		rec = httptest.NewRecorder()
		s.HTTPServer.ServeHTTP(rec, req)

		s.Require().Equal(http.StatusNotImplemented, rec.Code)
		var errorResp openapi.ErrorResponse
		err = json.Unmarshal(rec.Body.Bytes(), &errorResp)
		s.Require().NoError(err)
		s.Require().Contains(*errorResp.Message, "not implemented")
	})

	s.Run("Get tickets from non-existent organization - not implemented", func() {
		nonExistentID := uuid.New()
		req := httptest.NewRequest(
			http.MethodGet,
			fmt.Sprintf("/organizations/%s/tickets", nonExistentID.String()),
			nil,
		)
		rec := httptest.NewRecorder()
		s.HTTPServer.ServeHTTP(rec, req)

		s.Require().Equal(http.StatusNotImplemented, rec.Code)
	})

	s.Run("Get tickets with invalid organization ID - not implemented", func() {
		req := httptest.NewRequest(http.MethodGet, "/organizations/invalid-uuid/tickets", nil)
		rec := httptest.NewRecorder()
		s.HTTPServer.ServeHTTP(rec, req)

		s.Require().Equal(http.StatusBadRequest, rec.Code)
	})
}

func (s *OrganizationsSuite) TestGetOrganizationHierarchy() {
	s.Run("Get organization hierarchy - endpoint not found", func() {
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

		req = httptest.NewRequest(
			http.MethodGet,
			fmt.Sprintf("/organizations/%s/hierarchy", createResp.Id.String()),
			nil,
		)
		rec = httptest.NewRecorder()
		s.HTTPServer.ServeHTTP(rec, req)

		s.Require().Equal(http.StatusNotFound, rec.Code)
	})

	s.Run("Get hierarchy from non-existent organization - endpoint not found", func() {
		nonExistentID := uuid.New()
		req := httptest.NewRequest(
			http.MethodGet,
			fmt.Sprintf("/organizations/%s/hierarchy", nonExistentID.String()),
			nil,
		)
		rec := httptest.NewRecorder()
		s.HTTPServer.ServeHTTP(rec, req)

		s.Require().Equal(http.StatusNotFound, rec.Code)
	})

	s.Run("Get hierarchy with invalid organization ID - endpoint not found", func() {
		req := httptest.NewRequest(http.MethodGet, "/organizations/invalid-uuid/hierarchy", nil)
		rec := httptest.NewRecorder()
		s.HTTPServer.ServeHTTP(rec, req)

		s.Require().Equal(http.StatusNotFound, rec.Code)
	})
}
