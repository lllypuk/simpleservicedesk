package organizations_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"

	"simpleservicedesk/generated/openapi"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func (s *OrganizationsSuite) TestCreateOrganization() {
	s.Run("Create root organization", func() {
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
		var resp openapi.CreateOrganizationResponse
		err := json.Unmarshal(rec.Body.Bytes(), &resp)
		s.Require().NoError(err)
		s.Require().NotNil(resp.Id)
		s.Require().NotEqual(uuid.Nil, *resp.Id)
	})

	s.Run("Create sub-organization", func() {
		domain := "parent.com"
		parentOrgReq := openapi.CreateOrganizationRequest{
			Name:   "Parent Organization",
			Domain: &domain,
		}
		reqBody, _ := json.Marshal(parentOrgReq)

		req := httptest.NewRequest(http.MethodPost, "/organizations", bytes.NewBuffer(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		s.HTTPServer.ServeHTTP(rec, req)

		s.Require().Equal(http.StatusCreated, rec.Code)
		var parentResp openapi.CreateOrganizationResponse
		err := json.Unmarshal(rec.Body.Bytes(), &parentResp)
		s.Require().NoError(err)

		subDomain := "sub.parent.com"
		subOrgReq := openapi.CreateOrganizationRequest{
			Name:     "Sub Organization",
			Domain:   &subDomain,
			ParentId: parentResp.Id,
		}
		reqBody, _ = json.Marshal(subOrgReq)

		req = httptest.NewRequest(http.MethodPost, "/organizations", bytes.NewBuffer(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec = httptest.NewRecorder()
		s.HTTPServer.ServeHTTP(rec, req)

		s.Require().Equal(http.StatusCreated, rec.Code)
		var subResp openapi.CreateOrganizationResponse
		err = json.Unmarshal(rec.Body.Bytes(), &subResp)
		s.Require().NoError(err)
		s.Require().NotNil(subResp.Id)
		s.Require().NotEqual(uuid.Nil, *subResp.Id)
	})

	s.Run("Missing name returns 400", func() {
		domain := "example.com"
		orgReq := openapi.CreateOrganizationRequest{
			Name:   "",
			Domain: &domain,
		}
		reqBody, _ := json.Marshal(orgReq)

		req := httptest.NewRequest(http.MethodPost, "/organizations", bytes.NewBuffer(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		s.HTTPServer.ServeHTTP(rec, req)

		s.Require().Equal(http.StatusBadRequest, rec.Code)
	})

	s.Run("Missing domain returns 400", func() {
		orgReq := openapi.CreateOrganizationRequest{
			Name:   "Example Organization",
			Domain: nil,
		}
		reqBody, _ := json.Marshal(orgReq)

		req := httptest.NewRequest(http.MethodPost, "/organizations", bytes.NewBuffer(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		s.HTTPServer.ServeHTTP(rec, req)

		s.Require().Equal(http.StatusBadRequest, rec.Code)
	})

	s.Run("Empty domain returns 400", func() {
		domain := ""
		orgReq := openapi.CreateOrganizationRequest{
			Name:   "Example Organization",
			Domain: &domain,
		}
		reqBody, _ := json.Marshal(orgReq)

		req := httptest.NewRequest(http.MethodPost, "/organizations", bytes.NewBuffer(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		s.HTTPServer.ServeHTTP(rec, req)

		s.Require().Equal(http.StatusBadRequest, rec.Code)
	})
}
