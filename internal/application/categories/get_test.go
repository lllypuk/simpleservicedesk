package categories_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"

	"simpleservicedesk/generated/openapi"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func (s *CategoriesSuite) TestGetCategory() {
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

	// Create a category
	description := "Category for get testing"
	categoryReq := openapi.CreateCategoryRequest{
		Name:           "Get Test Category",
		Description:    &description,
		OrganizationId: *orgResp.Id,
	}
	reqBody, _ := json.Marshal(categoryReq)

	req := httptest.NewRequest(http.MethodPost, "/categories", bytes.NewBuffer(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	s.HTTPServer.ServeHTTP(rec, req)

	s.Require().Equal(http.StatusCreated, rec.Code)
	var createResp openapi.CreateCategoryResponse
	err = json.Unmarshal(rec.Body.Bytes(), &createResp)
	s.Require().NoError(err)

	s.Run("Get existing category", func() {
		getReq := httptest.NewRequest(http.MethodGet, "/categories/"+createResp.Id.String(), nil)
		getRec := httptest.NewRecorder()
		s.HTTPServer.ServeHTTP(getRec, getReq)

		s.Require().Equal(http.StatusOK, getRec.Code)
		var getResp openapi.GetCategoryResponse
		getUnmarshalErr := json.Unmarshal(getRec.Body.Bytes(), &getResp)
		s.Require().NoError(getUnmarshalErr)
		s.Require().NotNil(getResp.Id)
		s.Require().Equal(*createResp.Id, *getResp.Id)
		s.Require().NotNil(getResp.Name)
		s.Require().Equal("Get Test Category", *getResp.Name)
		s.Require().NotNil(getResp.Description)
		s.Require().Equal("Category for get testing", *getResp.Description)
		s.Require().NotNil(getResp.OrganizationId)
		s.Require().Equal(*orgResp.Id, *getResp.OrganizationId)
		s.Require().NotNil(getResp.IsActive)
		s.Require().True(*getResp.IsActive)
	})

	s.Run("Get non-existent category", func() {
		nonExistentID := uuid.New()
		getReq := httptest.NewRequest(http.MethodGet, "/categories/"+nonExistentID.String(), nil)
		getRec := httptest.NewRecorder()
		s.HTTPServer.ServeHTTP(getRec, getReq)

		s.Require().Equal(http.StatusNotFound, getRec.Code)
	})

	s.Run("Get category with invalid ID", func() {
		getReq := httptest.NewRequest(http.MethodGet, "/categories/invalid-uuid", nil)
		getRec := httptest.NewRecorder()
		s.HTTPServer.ServeHTTP(getRec, getReq)

		s.Require().Equal(http.StatusBadRequest, getRec.Code)
	})
}
