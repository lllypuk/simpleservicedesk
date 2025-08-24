package users_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"

	"simpleservicedesk/generated/openapi"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

func (s *UsersSuite) TestUpdateUser() {
	// First create a user
	email := openapi_types.Email("update@test.com")
	userReq := openapi.CreateUserRequest{
		Name:     "Update Test User",
		Email:    email,
		Password: "password123",
	}
	reqBody, _ := json.Marshal(userReq)

	req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBuffer(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	s.HTTPServer.ServeHTTP(rec, req)

	s.Require().Equal(http.StatusCreated, rec.Code)
	var createResp openapi.CreateUserResponse
	err := json.Unmarshal(rec.Body.Bytes(), &createResp)
	s.Require().NoError(err)

	s.Run("Update user name", func() {
		newName := "Updated User Name"
		updateReq := openapi.UpdateUserRequest{
			Name: &newName,
		}
		updateBody, _ := json.Marshal(updateReq)

		nameUpdateReq := httptest.NewRequest(
			http.MethodPut,
			"/users/"+createResp.Id.String(),
			bytes.NewBuffer(updateBody),
		)
		nameUpdateReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		nameUpdateRec := httptest.NewRecorder()
		s.HTTPServer.ServeHTTP(nameUpdateRec, nameUpdateReq)

		s.Require().Equal(http.StatusOK, nameUpdateRec.Code)
		var updateResp openapi.GetUserResponse
		nameUpdateErr := json.Unmarshal(nameUpdateRec.Body.Bytes(), &updateResp)
		s.Require().NoError(nameUpdateErr)
		s.Require().NotNil(updateResp.Name)
		s.Require().Equal(newName, *updateResp.Name)
	})

	s.Run("Update user email", func() {
		newEmail := openapi_types.Email("newemail@test.com")
		updateReq := openapi.UpdateUserRequest{
			Email: &newEmail,
		}
		updateBody, _ := json.Marshal(updateReq)

		emailUpdateReq := httptest.NewRequest(
			http.MethodPut,
			"/users/"+createResp.Id.String(),
			bytes.NewBuffer(updateBody),
		)
		emailUpdateReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		emailUpdateRec := httptest.NewRecorder()
		s.HTTPServer.ServeHTTP(emailUpdateRec, emailUpdateReq)

		s.Require().Equal(http.StatusOK, emailUpdateRec.Code)
		var updateResp openapi.GetUserResponse
		emailUpdateErr := json.Unmarshal(emailUpdateRec.Body.Bytes(), &updateResp)
		s.Require().NoError(emailUpdateErr)
		s.Require().NotNil(updateResp.Email)
		s.Require().Equal(string(newEmail), string(*updateResp.Email))
	})

	s.Run("Update user active status", func() {
		isActive := false
		updateReq := openapi.UpdateUserRequest{
			IsActive: &isActive,
		}
		updateBody, _ := json.Marshal(updateReq)

		statusUpdateReq := httptest.NewRequest(
			http.MethodPut,
			"/users/"+createResp.Id.String(),
			bytes.NewBuffer(updateBody),
		)
		statusUpdateReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		statusUpdateRec := httptest.NewRecorder()
		s.HTTPServer.ServeHTTP(statusUpdateRec, statusUpdateReq)

		s.Require().Equal(http.StatusOK, statusUpdateRec.Code)
		var updateResp openapi.GetUserResponse
		statusUpdateErr := json.Unmarshal(statusUpdateRec.Body.Bytes(), &updateResp)
		s.Require().NoError(statusUpdateErr)
		s.Require().NotNil(updateResp.IsActive)
		s.Require().False(*updateResp.IsActive)
	})

	s.Run("Update non-existent user", func() {
		nonExistentID := uuid.New()
		newName := "Updated Name"
		updateReq := openapi.UpdateUserRequest{
			Name: &newName,
		}
		updateBody, _ := json.Marshal(updateReq)

		nonExistentReq := httptest.NewRequest(
			http.MethodPut,
			"/users/"+nonExistentID.String(),
			bytes.NewBuffer(updateBody),
		)
		nonExistentReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		nonExistentRec := httptest.NewRecorder()
		s.HTTPServer.ServeHTTP(nonExistentRec, nonExistentReq)

		s.Require().Equal(http.StatusNotFound, nonExistentRec.Code)
	})

	s.Run("Update user with invalid data", func() {
		invalidName := "" // Empty name is invalid
		updateReq := openapi.UpdateUserRequest{
			Name: &invalidName,
		}
		updateBody, _ := json.Marshal(updateReq)

		invalidUpdateReq := httptest.NewRequest(
			http.MethodPut,
			"/users/"+createResp.Id.String(),
			bytes.NewBuffer(updateBody),
		)
		invalidUpdateReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		invalidUpdateRec := httptest.NewRecorder()
		s.HTTPServer.ServeHTTP(invalidUpdateRec, invalidUpdateReq)

		s.Require().Equal(http.StatusBadRequest, invalidUpdateRec.Code)
	})
}
