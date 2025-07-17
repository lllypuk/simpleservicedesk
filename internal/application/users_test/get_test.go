package users_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"

	"simpleservicedesk/generated/openapi"
	"simpleservicedesk/internal/domain/users"

	"github.com/google/uuid"
)

func (s *UsersSuite) TestGetUser() {
	user, _ := users.CreateUser("test", "test@test.com")
	err := s.UsersRepo.SaveUser(context.Background(), *user)
	s.Require().NoError(err)

	s.Run("HTTP", func() {
		req := httptest.NewRequest(http.MethodGet, "/users/"+user.ID().String(), nil)
		rec := httptest.NewRecorder()
		s.HTTPServer.ServeHTTP(rec, req)

		s.Require().Equal(http.StatusOK, rec.Code, rec.Body.String())
		var resp openapi.GetUserResponse
		err = json.Unmarshal(rec.Body.Bytes(), &resp)
		s.Require().NoError(err)
		s.Require().Equal(user.ID(), *resp.Id)
		s.Require().Equal(user.Name(), *resp.Name)
		s.Require().Equal(user.Email(), string(*resp.Email))
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
