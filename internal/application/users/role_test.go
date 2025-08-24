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

func (s *UsersSuite) TestUpdateUserRole() {
	// Create a user for role testing
	userReq := openapi.CreateUserRequest{
		Name:     "Role Test User",
		Email:    "role@test.com",
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

	s.Run("Update user role to agent", func() {
		newRole := openapi.Agent
		roleReq := openapi.UpdateUserRoleRequest{
			Role: newRole,
		}
		roleBody, _ := json.Marshal(roleReq)

		roleUpdateReq := httptest.NewRequest(
			http.MethodPatch,
			"/users/"+createResp.Id.String()+"/role",
			bytes.NewBuffer(roleBody),
		)
		roleUpdateReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		roleUpdateRec := httptest.NewRecorder()
		s.HTTPServer.ServeHTTP(roleUpdateRec, roleUpdateReq)

		s.Require().Equal(http.StatusOK, roleUpdateRec.Code)
		var roleResp openapi.GetUserResponse
		roleUpdateErr := json.Unmarshal(roleUpdateRec.Body.Bytes(), &roleResp)
		s.Require().NoError(roleUpdateErr)
		s.Require().NotNil(roleResp.Role)
		s.Require().Equal(newRole, *roleResp.Role)
	})

	s.Run("Update user role to admin", func() {
		newRole := openapi.Admin
		roleReq := openapi.UpdateUserRoleRequest{
			Role: newRole,
		}
		roleBody, _ := json.Marshal(roleReq)

		adminUpdateReq := httptest.NewRequest(
			http.MethodPatch,
			"/users/"+createResp.Id.String()+"/role",
			bytes.NewBuffer(roleBody),
		)
		adminUpdateReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		adminUpdateRec := httptest.NewRecorder()
		s.HTTPServer.ServeHTTP(adminUpdateRec, adminUpdateReq)

		s.Require().Equal(http.StatusOK, adminUpdateRec.Code)
		var adminResp openapi.GetUserResponse
		adminUpdateErr := json.Unmarshal(adminUpdateRec.Body.Bytes(), &adminResp)
		s.Require().NoError(adminUpdateErr)
		s.Require().NotNil(adminResp.Role)
		s.Require().Equal(newRole, *adminResp.Role)
	})

	s.Run("Update role with invalid role", func() {
		invalidRoleReq := openapi.UpdateUserRoleRequest{
			Role: "invalid_role",
		}
		invalidBody, _ := json.Marshal(invalidRoleReq)

		invalidUpdateReq := httptest.NewRequest(
			http.MethodPatch,
			"/users/"+createResp.Id.String()+"/role",
			bytes.NewBuffer(invalidBody),
		)
		invalidUpdateReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		invalidUpdateRec := httptest.NewRecorder()
		s.HTTPServer.ServeHTTP(invalidUpdateRec, invalidUpdateReq)

		s.Require().Equal(http.StatusBadRequest, invalidUpdateRec.Code)
	})

	s.Run("Update role for non-existent user", func() {
		nonExistentID := uuid.New()
		newRole := openapi.Agent
		roleReq := openapi.UpdateUserRoleRequest{
			Role: newRole,
		}
		roleBody, _ := json.Marshal(roleReq)

		nonExistentUpdateReq := httptest.NewRequest(
			http.MethodPatch,
			"/users/"+nonExistentID.String()+"/role",
			bytes.NewBuffer(roleBody),
		)
		nonExistentUpdateReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		nonExistentUpdateRec := httptest.NewRecorder()
		s.HTTPServer.ServeHTTP(nonExistentUpdateRec, nonExistentUpdateReq)

		s.Require().Equal(http.StatusNotFound, nonExistentUpdateRec.Code)
	})

	s.Run("Update role with same role", func() {
		// First get current user to see their role
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
		s.Require().NotNil(getResp.Role)

		// Now try to set the same role
		currentRole := *getResp.Role
		sameRoleReq := openapi.UpdateUserRoleRequest{
			Role: currentRole,
		}
		sameRoleBody, _ := json.Marshal(sameRoleReq)

		sameRoleUpdateReq := httptest.NewRequest(
			http.MethodPatch,
			"/users/"+createResp.Id.String()+"/role",
			bytes.NewBuffer(sameRoleBody),
		)
		sameRoleUpdateReq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		sameRoleUpdateRec := httptest.NewRecorder()
		s.HTTPServer.ServeHTTP(sameRoleUpdateRec, sameRoleUpdateReq)

		s.Require().Equal(http.StatusOK, sameRoleUpdateRec.Code)
		var sameRoleResp openapi.GetUserResponse
		sameRoleUpdateErr := json.Unmarshal(sameRoleUpdateRec.Body.Bytes(), &sameRoleResp)
		s.Require().NoError(sameRoleUpdateErr)
		s.Require().NotNil(sameRoleResp.Role)
		s.Require().Equal(currentRole, *sameRoleResp.Role)
	})
}
