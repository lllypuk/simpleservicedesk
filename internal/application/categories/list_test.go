package categories_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"

	"simpleservicedesk/generated/openapi"

	"github.com/labstack/echo/v4"
)

func (s *CategoriesSuite) TestListCategories() {
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

	// Create multiple categories
	for i := 1; i <= 3; i++ {
		description := "Category " + strconv.Itoa(i) + " for list testing"
		categoryReq := openapi.CreateCategoryRequest{
			Name:           "List Test Category " + strconv.Itoa(i),
			Description:    &description,
			OrganizationId: *orgResp.Id,
		}
		reqBody, _ := json.Marshal(categoryReq)

		req := httptest.NewRequest(http.MethodPost, "/categories", bytes.NewBuffer(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		s.HTTPServer.ServeHTTP(rec, req)

		s.Require().Equal(http.StatusCreated, rec.Code)
	}

	s.Run("List all categories", func() {
		getReq := httptest.NewRequest(http.MethodGet, "/categories", nil)
		getRec := httptest.NewRecorder()
		s.HTTPServer.ServeHTTP(getRec, getReq)

		s.Require().Equal(http.StatusOK, getRec.Code)
		var listResp openapi.ListCategoriesResponse
		listUnmarshalErr := json.Unmarshal(getRec.Body.Bytes(), &listResp)
		s.Require().NoError(listUnmarshalErr)
		s.Require().NotNil(listResp.Categories)
		s.Require().Len(*listResp.Categories, 3)

		// Verify all categories belong to our organization
		for _, category := range *listResp.Categories {
			s.Require().NotNil(category.OrganizationId)
			s.Require().Equal(*orgResp.Id, *category.OrganizationId)
			s.Require().NotNil(category.IsActive)
			s.Require().True(*category.IsActive)
		}
	})

	s.Run("List categories with organization filter", func() {
		orgIDParam := orgResp.Id.String()
		getReq := httptest.NewRequest(http.MethodGet, "/categories?organization_id="+orgIDParam, nil)
		getRec := httptest.NewRecorder()
		s.HTTPServer.ServeHTTP(getRec, getReq)

		s.Require().Equal(http.StatusOK, getRec.Code)
		var listResp openapi.ListCategoriesResponse
		orgFilterUnmarshalErr := json.Unmarshal(getRec.Body.Bytes(), &listResp)
		s.Require().NoError(orgFilterUnmarshalErr)
		s.Require().NotNil(listResp.Categories)
		s.Require().Len(*listResp.Categories, 3)
	})

	s.Run("List categories with active filter", func() {
		getReq := httptest.NewRequest(http.MethodGet, "/categories?is_active=true", nil)
		getRec := httptest.NewRecorder()
		s.HTTPServer.ServeHTTP(getRec, getReq)

		s.Require().Equal(http.StatusOK, getRec.Code)
		var listResp openapi.ListCategoriesResponse
		activeFilterUnmarshalErr := json.Unmarshal(getRec.Body.Bytes(), &listResp)
		s.Require().NoError(activeFilterUnmarshalErr)
		s.Require().NotNil(listResp.Categories)
		s.Require().Len(*listResp.Categories, 3)
	})
}
