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

func (s *OrganizationsSuite) TestGetOrganization() {
	s.Run("Get existing organization", func() {
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

		req = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/organizations/%s", createResp.Id.String()), nil)
		rec = httptest.NewRecorder()
		s.HTTPServer.ServeHTTP(rec, req)

		s.Require().Equal(http.StatusOK, rec.Code)
		var getResp openapi.GetOrganizationResponse
		err = json.Unmarshal(rec.Body.Bytes(), &getResp)
		s.Require().NoError(err)
		s.Require().NotNil(getResp.Id)
		s.Require().Equal(*createResp.Id, *getResp.Id)
		s.Require().Equal("Example Organization", *getResp.Name)
		s.Require().Equal(domain, *getResp.Domain)
	})

	s.Run("Get non-existent organization returns 404", func() {
		nonExistentID := uuid.New()
		req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/organizations/%s", nonExistentID.String()), nil)
		rec := httptest.NewRecorder()
		s.HTTPServer.ServeHTTP(rec, req)

		s.Require().Equal(http.StatusNotFound, rec.Code)
	})

	s.Run("Get organization with invalid ID returns 400", func() {
		req := httptest.NewRequest(http.MethodGet, "/organizations/invalid-uuid", nil)
		rec := httptest.NewRecorder()
		s.HTTPServer.ServeHTTP(rec, req)

		s.Require().Equal(http.StatusBadRequest, rec.Code)
	})
}
