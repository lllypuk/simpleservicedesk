package auth_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"

	"simpleservicedesk/generated/openapi"

	"github.com/labstack/echo/v4"
)

func (s *AuthSuite) TestLogin() {
	s.Run("Success", func() {
		createUserRequest := openapi.CreateUserRequest{
			Name:     "Alice",
			Email:    "alice@example.com",
			Password: "correct-password",
		}
		createUserBody, err := json.Marshal(createUserRequest)
		s.Require().NoError(err)

		createUserReq := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBuffer(createUserBody))
		createUserReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		createUserRecorder := httptest.NewRecorder()
		s.HTTPServer.ServeHTTP(createUserRecorder, createUserReq)
		s.Require().Equal(http.StatusCreated, createUserRecorder.Code)

		loginRequest := openapi.LoginRequest{
			Email:    "alice@example.com",
			Password: "correct-password",
		}
		loginBody, err := json.Marshal(loginRequest)
		s.Require().NoError(err)

		req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(loginBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		s.HTTPServer.ServeHTTP(rec, req)

		s.Require().Equal(http.StatusOK, rec.Code)

		var response openapi.LoginResponse
		err = json.Unmarshal(rec.Body.Bytes(), &response)
		s.Require().NoError(err)
		s.Require().NotEmpty(response.Token)
	})

	s.Run("Wrong password returns 401", func() {
		createUserRequest := openapi.CreateUserRequest{
			Name:     "Bob",
			Email:    "bob@example.com",
			Password: "correct-password",
		}
		createUserBody, err := json.Marshal(createUserRequest)
		s.Require().NoError(err)

		createUserReq := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBuffer(createUserBody))
		createUserReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		createUserRecorder := httptest.NewRecorder()
		s.HTTPServer.ServeHTTP(createUserRecorder, createUserReq)
		s.Require().Equal(http.StatusCreated, createUserRecorder.Code)

		loginRequest := openapi.LoginRequest{
			Email:    "bob@example.com",
			Password: "wrong-password",
		}
		loginBody, err := json.Marshal(loginRequest)
		s.Require().NoError(err)

		req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(loginBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		s.HTTPServer.ServeHTTP(rec, req)

		s.Require().Equal(http.StatusUnauthorized, rec.Code)
	})

	s.Run("Nonexistent user returns 401", func() {
		loginRequest := openapi.LoginRequest{
			Email:    "ghost@example.com",
			Password: "any-password",
		}
		loginBody, err := json.Marshal(loginRequest)
		s.Require().NoError(err)

		req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(loginBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		s.HTTPServer.ServeHTTP(rec, req)

		s.Require().Equal(http.StatusUnauthorized, rec.Code)
	})

	s.Run("Invalid JSON returns 400", func() {
		req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBufferString(`{"email":`))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		s.HTTPServer.ServeHTTP(rec, req)

		s.Require().Equal(http.StatusBadRequest, rec.Code)
	})
}
