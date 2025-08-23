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

func (s *CategoriesSuite) TestCreateCategory() {
	// First create an organization to associate the category with
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
	s.Require().NotNil(orgResp.Id)

	s.Run("Create root category", func() {
		description := "Root category for testing"
		categoryReq := openapi.CreateCategoryRequest{
			Name:           "Test Category",
			Description:    &description,
			OrganizationId: *orgResp.Id,
		}
		reqBody, _ := json.Marshal(categoryReq)

		req := httptest.NewRequest(http.MethodPost, "/categories", bytes.NewBuffer(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		s.HTTPServer.ServeHTTP(rec, req)

		s.Require().Equal(http.StatusCreated, rec.Code)
		var resp openapi.CreateCategoryResponse
		unmarshalErr := json.Unmarshal(rec.Body.Bytes(), &resp)
		s.Require().NoError(unmarshalErr)
		s.Require().NotNil(resp.Id)
		s.Require().NotEqual(uuid.Nil, *resp.Id)
	})

	s.Run("Create sub-category", func() {
		// First create a parent category
		description := "Parent category for testing"
		parentCategoryReq := openapi.CreateCategoryRequest{
			Name:           "Parent Category",
			Description:    &description,
			OrganizationId: *orgResp.Id,
		}
		reqBody, _ := json.Marshal(parentCategoryReq)

		req := httptest.NewRequest(http.MethodPost, "/categories", bytes.NewBuffer(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		s.HTTPServer.ServeHTTP(rec, req)

		s.Require().Equal(http.StatusCreated, rec.Code)
		var parentResp openapi.CreateCategoryResponse
		parentUnmarshalErr := json.Unmarshal(rec.Body.Bytes(), &parentResp)
		s.Require().NoError(parentUnmarshalErr)

		// Now create a sub-category
		subDescription := "Sub category for testing"
		subCategoryReq := openapi.CreateCategoryRequest{
			Name:           "Sub Category",
			Description:    &subDescription,
			OrganizationId: *orgResp.Id,
			ParentId:       parentResp.Id,
		}
		subReqBody, _ := json.Marshal(subCategoryReq)

		subReq := httptest.NewRequest(http.MethodPost, "/categories", bytes.NewBuffer(subReqBody))
		subReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		subRec := httptest.NewRecorder()
		s.HTTPServer.ServeHTTP(subRec, subReq)

		s.Require().Equal(http.StatusCreated, subRec.Code)
		var subResp openapi.CreateCategoryResponse
		subUnmarshalErr := json.Unmarshal(subRec.Body.Bytes(), &subResp)
		s.Require().NoError(subUnmarshalErr)
		s.Require().NotNil(subResp.Id)
		s.Require().NotEqual(uuid.Nil, *subResp.Id)
	})

	s.Run("Create category with invalid data", func() {
		categoryReq := openapi.CreateCategoryRequest{
			Name:           "", // Invalid: empty name
			OrganizationId: *orgResp.Id,
		}
		reqBody, _ := json.Marshal(categoryReq)

		req := httptest.NewRequest(http.MethodPost, "/categories", bytes.NewBuffer(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		s.HTTPServer.ServeHTTP(rec, req)

		s.Require().Equal(http.StatusBadRequest, rec.Code)
	})
}
