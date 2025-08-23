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

func (s *OrganizationsSuite) TestUpdateOrganization() {
	s.Run("Update organization name", func() {
		domain := "example.com"
		orgReq := openapi.CreateOrganizationRequest{
			Name:   "Original Organization",
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

		newName := "Updated Organization"
		updateReq := openapi.UpdateOrganizationRequest{
			Name: &newName,
		}
		reqBody, _ = json.Marshal(updateReq)

		req = httptest.NewRequest(
			http.MethodPut,
			fmt.Sprintf("/organizations/%s", createResp.Id.String()),
			bytes.NewBuffer(reqBody),
		)
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec = httptest.NewRecorder()
		s.HTTPServer.ServeHTTP(rec, req)

		s.Require().Equal(http.StatusOK, rec.Code)
		var updateResp openapi.GetOrganizationResponse
		err = json.Unmarshal(rec.Body.Bytes(), &updateResp)
		s.Require().NoError(err)
		s.Require().NotNil(updateResp.Name)
		s.Require().Equal(newName, *updateResp.Name)
	})

	s.Run("Update organization domain", func() {
		domain := "original.com"
		orgReq := openapi.CreateOrganizationRequest{
			Name:   "Organization",
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

		newDomain := "updated.com"
		updateReq := openapi.UpdateOrganizationRequest{
			Domain: &newDomain,
		}
		reqBody, _ = json.Marshal(updateReq)

		req = httptest.NewRequest(
			http.MethodPut,
			fmt.Sprintf("/organizations/%s", createResp.Id.String()),
			bytes.NewBuffer(reqBody),
		)
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec = httptest.NewRecorder()
		s.HTTPServer.ServeHTTP(rec, req)

		s.Require().Equal(http.StatusOK, rec.Code)
		var updateResp openapi.GetOrganizationResponse
		err = json.Unmarshal(rec.Body.Bytes(), &updateResp)
		s.Require().NoError(err)
		s.Require().NotNil(updateResp.Domain)
		s.Require().Equal(newDomain, *updateResp.Domain)
	})

	s.Run("Update organization status", func() {
		domain := "example.com"
		orgReq := openapi.CreateOrganizationRequest{
			Name:   "Organization",
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

		isActive := false
		updateReq := openapi.UpdateOrganizationRequest{
			IsActive: &isActive,
		}
		reqBody, _ = json.Marshal(updateReq)

		req = httptest.NewRequest(
			http.MethodPut,
			fmt.Sprintf("/organizations/%s", createResp.Id.String()),
			bytes.NewBuffer(reqBody),
		)
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec = httptest.NewRecorder()
		s.HTTPServer.ServeHTTP(rec, req)

		s.Require().Equal(http.StatusOK, rec.Code)
		var updateResp openapi.GetOrganizationResponse
		err = json.Unmarshal(rec.Body.Bytes(), &updateResp)
		s.Require().NoError(err)
		s.Require().NotNil(updateResp.IsActive)
		s.Require().False(*updateResp.IsActive)
	})

	s.Run("Update non-existent organization returns 404", func() {
		nonExistentID := uuid.New()
		newName := "Updated Organization"
		updateReq := openapi.UpdateOrganizationRequest{
			Name: &newName,
		}
		reqBody, _ := json.Marshal(updateReq)

		req := httptest.NewRequest(
			http.MethodPut,
			fmt.Sprintf("/organizations/%s", nonExistentID.String()),
			bytes.NewBuffer(reqBody),
		)
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		s.HTTPServer.ServeHTTP(rec, req)

		s.Require().Equal(http.StatusNotFound, rec.Code)
	})

	s.Run("Update organization with invalid ID returns 400", func() {
		newName := "Updated Organization"
		updateReq := openapi.UpdateOrganizationRequest{
			Name: &newName,
		}
		reqBody, _ := json.Marshal(updateReq)

		req := httptest.NewRequest(http.MethodPut, "/organizations/invalid-uuid", bytes.NewBuffer(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		s.HTTPServer.ServeHTTP(rec, req)

		s.Require().Equal(http.StatusBadRequest, rec.Code)
	})

	s.Run("Update organization with empty request should not modify", func() {
		domain := "example.com"
		orgReq := openapi.CreateOrganizationRequest{
			Name:   "Organization",
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

		updateReq := openapi.UpdateOrganizationRequest{}
		reqBody, _ = json.Marshal(updateReq)

		req = httptest.NewRequest(
			http.MethodPut,
			fmt.Sprintf("/organizations/%s", createResp.Id.String()),
			bytes.NewBuffer(reqBody),
		)
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec = httptest.NewRecorder()
		s.HTTPServer.ServeHTTP(rec, req)

		s.Require().Equal(http.StatusOK, rec.Code)
	})
}
