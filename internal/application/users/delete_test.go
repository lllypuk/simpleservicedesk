package users_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"

	"simpleservicedesk/generated/openapi"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func (s *UsersSuite) TestDeleteUser() {
	// Create a user for deletion
	userReq := openapi.CreateUserRequest{
		Name:     "Delete Test User",
		Email:    "delete@test.com",
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

	s.Run("Delete existing user", func() {
		deleteReq := httptest.NewRequest(
			http.MethodDelete,
			"/users/"+createResp.Id.String(),
			nil,
		)
		deleteRec := httptest.NewRecorder()
		s.HTTPServer.ServeHTTP(deleteRec, deleteReq)

		s.Require().Equal(http.StatusNoContent, deleteRec.Code)

		// Verify user is deactivated by trying to get it
		getReq := httptest.NewRequest(
			http.MethodGet,
			"/users/"+createResp.Id.String(),
			nil,
		)
		getRec := httptest.NewRecorder()
		s.HTTPServer.ServeHTTP(getRec, getReq)

		s.Require().Equal(http.StatusOK, getRec.Code)
		var getResp openapi.GetUserResponse
		getUnmarshalErr := json.Unmarshal(getRec.Body.Bytes(), &getResp)
		s.Require().NoError(getUnmarshalErr)
		if getResp.IsActive != nil {
			s.Require().False(*getResp.IsActive) // Should be deactivated
		}
	})

	s.Run("Delete non-existent user", func() {
		nonExistentID := uuid.New()
		deleteReq := httptest.NewRequest(
			http.MethodDelete,
			"/users/"+nonExistentID.String(),
			nil,
		)
		deleteRec := httptest.NewRecorder()
		s.HTTPServer.ServeHTTP(deleteRec, deleteReq)

		s.Require().Equal(http.StatusNotFound, deleteRec.Code)
	})

	s.Run("Delete user with invalid ID", func() {
		deleteReq := httptest.NewRequest(http.MethodDelete, "/users/invalid-uuid", nil)
		deleteRec := httptest.NewRecorder()
		s.HTTPServer.ServeHTTP(deleteRec, deleteReq)

		s.Require().Equal(http.StatusBadRequest, deleteRec.Code)
	})
}
