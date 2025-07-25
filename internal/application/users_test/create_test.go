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

func (s *UsersSuite) TestCreateUser() {
	s.Run("HTTP", func() {
		userReq := openapi.CreateUserRequest{
			Name:     "John Doe",
			Email:    "john.doe@example.com",
			Password: "password123",
		}
		reqBody, _ := json.Marshal(userReq)

		req := httptest.NewRequest(http.MethodPost, "/users", bytes.NewBuffer(reqBody))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		s.HTTPServer.ServeHTTP(rec, req)

		s.Require().Equal(http.StatusCreated, rec.Code)
		var resp openapi.CreateUserResponse
		err := json.Unmarshal(rec.Body.Bytes(), &resp)
		s.Require().NoError(err)
		s.Require().NotEqual("", resp.Id)
		s.Require().NotEqual(uuid.Nil, resp.Id)
	})
}
