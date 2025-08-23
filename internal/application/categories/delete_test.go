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

func (s *CategoriesSuite) TestDeleteCategory() {
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
	description := "Category for delete testing"
	categoryReq := openapi.CreateCategoryRequest{
		Name:           "Delete Test Category",
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

	s.Run("Delete existing category", func() {
		deleteReq := httptest.NewRequest(http.MethodDelete, "/categories/"+createResp.Id.String(), nil)
		deleteRec := httptest.NewRecorder()
		s.HTTPServer.ServeHTTP(deleteRec, deleteReq)

		s.Require().Equal(http.StatusNoContent, deleteRec.Code)

		// Verify the category is deleted by trying to get it
		getReq := httptest.NewRequest(http.MethodGet, "/categories/"+createResp.Id.String(), nil)
		getRec := httptest.NewRecorder()
		s.HTTPServer.ServeHTTP(getRec, getReq)

		s.Require().Equal(http.StatusNotFound, getRec.Code)
	})

	s.Run("Delete non-existent category", func() {
		nonExistentID := uuid.New()
		deleteReq := httptest.NewRequest(http.MethodDelete, "/categories/"+nonExistentID.String(), nil)
		deleteRec := httptest.NewRecorder()
		s.HTTPServer.ServeHTTP(deleteRec, deleteReq)

		s.Require().Equal(http.StatusNotFound, deleteRec.Code)
	})

	s.Run("Delete category with invalid ID", func() {
		deleteReq := httptest.NewRequest(http.MethodDelete, "/categories/invalid-uuid", nil)
		deleteRec := httptest.NewRecorder()
		s.HTTPServer.ServeHTTP(deleteRec, deleteReq)

		s.Require().Equal(http.StatusBadRequest, deleteRec.Code)
	})
}
