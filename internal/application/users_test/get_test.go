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

func (s *UsersSuite) TestGetUser() {
	// Создаем пользователя через HTTP API
	userReq := openapi.CreateUserRequest{
		Name:     "test",
		Email:    "test@test.com",
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

	s.Run("HTTP", func() {
		req := httptest.NewRequest(http.MethodGet, "/users/"+createResp.Id.String(), nil)
		rec := httptest.NewRecorder()
		s.HTTPServer.ServeHTTP(rec, req)

		s.Require().Equal(http.StatusOK, rec.Code, rec.Body.String())
		var resp openapi.GetUserResponse
		err = json.Unmarshal(rec.Body.Bytes(), &resp)
		s.Require().NoError(err)
		s.Require().Equal(createResp.Id.String(), resp.Id.String())
		s.Require().Equal("test", *resp.Name)
		s.Require().Equal("test@test.com", string(*resp.Email))
	})
}

func (s *UsersSuite) TestGetUserNotFound() {
	s.Run("HTTP", func() {
		req := httptest.NewRequest(http.MethodGet, "/users/"+uuid.New().String(), nil)
		rec := httptest.NewRecorder()
		s.HTTPServer.ServeHTTP(rec, req)

		s.Require().Equal(http.StatusNotFound, rec.Code, rec.Body.String())
		s.Require().Equal(`{"message":"user not found"}`+"\n", rec.Body.String())
	})
}
