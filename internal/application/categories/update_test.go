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

func (s *CategoriesSuite) TestUpdateCategory() {
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
	description := "Category for update testing"
	categoryReq := openapi.CreateCategoryRequest{
		Name:           "Update Test Category",
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

	s.Run("Update category name", func() {
		newName := "Updated Category Name"
		updateReq := openapi.UpdateCategoryRequest{
			Name: &newName,
		}
		updateBody, _ := json.Marshal(updateReq)

		nameUpdateReq := httptest.NewRequest(
			http.MethodPut,
			"/categories/"+createResp.Id.String(),
			bytes.NewBuffer(updateBody),
		)
		nameUpdateReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		nameUpdateRec := httptest.NewRecorder()
		s.HTTPServer.ServeHTTP(nameUpdateRec, nameUpdateReq)

		s.Require().Equal(http.StatusOK, nameUpdateRec.Code)
		var updateResp openapi.GetCategoryResponse
		nameUpdateErr := json.Unmarshal(nameUpdateRec.Body.Bytes(), &updateResp)
		s.Require().NoError(nameUpdateErr)
		s.Require().NotNil(updateResp.Name)
		s.Require().Equal(newName, *updateResp.Name)
	})

	s.Run("Update category description", func() {
		newDescription := "Updated category description"
		updateReq := openapi.UpdateCategoryRequest{
			Description: &newDescription,
		}
		updateBody, _ := json.Marshal(updateReq)

		descUpdateReq := httptest.NewRequest(
			http.MethodPut,
			"/categories/"+createResp.Id.String(),
			bytes.NewBuffer(updateBody),
		)
		descUpdateReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		descUpdateRec := httptest.NewRecorder()
		s.HTTPServer.ServeHTTP(descUpdateRec, descUpdateReq)

		s.Require().Equal(http.StatusOK, descUpdateRec.Code)
		var updateResp openapi.GetCategoryResponse
		descUpdateErr := json.Unmarshal(descUpdateRec.Body.Bytes(), &updateResp)
		s.Require().NoError(descUpdateErr)
		s.Require().NotNil(updateResp.Description)
		s.Require().Equal(newDescription, *updateResp.Description)
	})

	s.Run("Update category active status", func() {
		isActive := false
		updateReq := openapi.UpdateCategoryRequest{
			IsActive: &isActive,
		}
		updateBody, _ := json.Marshal(updateReq)

		statusUpdateReq := httptest.NewRequest(
			http.MethodPut,
			"/categories/"+createResp.Id.String(),
			bytes.NewBuffer(updateBody),
		)
		statusUpdateReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		statusUpdateRec := httptest.NewRecorder()
		s.HTTPServer.ServeHTTP(statusUpdateRec, statusUpdateReq)

		s.Require().Equal(http.StatusOK, statusUpdateRec.Code)
		var updateResp openapi.GetCategoryResponse
		statusUpdateErr := json.Unmarshal(statusUpdateRec.Body.Bytes(), &updateResp)
		s.Require().NoError(statusUpdateErr)
		s.Require().NotNil(updateResp.IsActive)
		s.Require().False(*updateResp.IsActive)
	})

	s.Run("Update non-existent category", func() {
		nonExistentID := uuid.New()
		newName := "Updated Name"
		updateReq := openapi.UpdateCategoryRequest{
			Name: &newName,
		}
		updateBody, _ := json.Marshal(updateReq)

		nonExistentReq := httptest.NewRequest(
			http.MethodPut,
			"/categories/"+nonExistentID.String(),
			bytes.NewBuffer(updateBody),
		)
		nonExistentReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		nonExistentRec := httptest.NewRecorder()
		s.HTTPServer.ServeHTTP(nonExistentRec, nonExistentReq)

		s.Require().Equal(http.StatusNotFound, nonExistentRec.Code)
	})

	s.Run("Update category with invalid name", func() {
		invalidName := "" // Empty name is invalid
		updateReq := openapi.UpdateCategoryRequest{
			Name: &invalidName,
		}
		updateBody, _ := json.Marshal(updateReq)

		invalidUpdateReq := httptest.NewRequest(
			http.MethodPut,
			"/categories/"+createResp.Id.String(),
			bytes.NewBuffer(updateBody),
		)
		invalidUpdateReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		invalidUpdateRec := httptest.NewRecorder()
		s.HTTPServer.ServeHTTP(invalidUpdateRec, invalidUpdateReq)

		s.Require().Equal(http.StatusBadRequest, invalidUpdateRec.Code)
	})
}
