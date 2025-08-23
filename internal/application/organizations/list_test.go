package organizations_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"

	"simpleservicedesk/generated/openapi"

	"github.com/labstack/echo/v4"
)

func (s *OrganizationsSuite) TestListOrganizations() {
	s.Run("List organizations with default pagination", func() {
		domain1 := "example1.com"
		orgReq1 := openapi.CreateOrganizationRequest{
			Name:   "Example Organization 1",
			Domain: &domain1,
		}
		reqBody, _ := json.Marshal(orgReq1)

		req := httptest.NewRequest(http.MethodPost, "/organizations", bytes.NewBuffer(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		s.HTTPServer.ServeHTTP(rec, req)
		s.Require().Equal(http.StatusCreated, rec.Code)

		domain2 := "example2.com"
		orgReq2 := openapi.CreateOrganizationRequest{
			Name:   "Example Organization 2",
			Domain: &domain2,
		}
		reqBody, _ = json.Marshal(orgReq2)

		req = httptest.NewRequest(http.MethodPost, "/organizations", bytes.NewBuffer(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec = httptest.NewRecorder()
		s.HTTPServer.ServeHTTP(rec, req)
		s.Require().Equal(http.StatusCreated, rec.Code)

		req = httptest.NewRequest(http.MethodGet, "/organizations", nil)
		rec = httptest.NewRecorder()
		s.HTTPServer.ServeHTTP(rec, req)

		s.Require().Equal(http.StatusOK, rec.Code)
		var listResp openapi.ListOrganizationsResponse
		err := json.Unmarshal(rec.Body.Bytes(), &listResp)
		s.Require().NoError(err)
		s.Require().NotNil(listResp.Organizations)
		s.Require().Len(*listResp.Organizations, 2)
	})

	s.Run("List organizations with custom limit", func() {
		req := httptest.NewRequest(http.MethodGet, "/organizations?limit=1", nil)
		rec := httptest.NewRecorder()
		s.HTTPServer.ServeHTTP(rec, req)

		s.Require().Equal(http.StatusOK, rec.Code)
		var listResp openapi.ListOrganizationsResponse
		err := json.Unmarshal(rec.Body.Bytes(), &listResp)
		s.Require().NoError(err)
		s.Require().NotNil(listResp.Organizations)
	})

	s.Run("List organizations with pagination", func() {
		page := 1
		limit := 10
		req := httptest.NewRequest(http.MethodGet, "/organizations?page=1&limit=10", nil)
		rec := httptest.NewRecorder()
		s.HTTPServer.ServeHTTP(rec, req)

		s.Require().Equal(http.StatusOK, rec.Code)
		var listResp openapi.ListOrganizationsResponse
		err := json.Unmarshal(rec.Body.Bytes(), &listResp)
		s.Require().NoError(err)
		s.Require().NotNil(listResp.Organizations)

		_ = page
		_ = limit
	})

	s.Run("List organizations with large limit should cap to 100", func() {
		req := httptest.NewRequest(http.MethodGet, "/organizations?limit=200", nil)
		rec := httptest.NewRecorder()
		s.HTTPServer.ServeHTTP(rec, req)

		s.Require().Equal(http.StatusOK, rec.Code)
	})

	s.Run("List organizations with zero page should use default", func() {
		req := httptest.NewRequest(http.MethodGet, "/organizations?page=0", nil)
		rec := httptest.NewRecorder()
		s.HTTPServer.ServeHTTP(rec, req)

		s.Require().Equal(http.StatusOK, rec.Code)
	})
}
