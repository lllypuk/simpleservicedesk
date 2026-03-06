//go:build integration
// +build integration

package shared

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"

	"simpleservicedesk/generated/openapi"
	userdomain "simpleservicedesk/internal/domain/users"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	openapi_types "github.com/oapi-codegen/runtime/types"
	"golang.org/x/crypto/bcrypt"
)

type TestAuthUser struct {
	UserID     uuid.UUID
	Email      string
	Passphrase string
	Role       userdomain.Role
	Token      string
}

func (s *IntegrationSuite) MustCreateAndLoginTestUser(role userdomain.Role) TestAuthUser {
	user := s.MustCreateTestUser(role)
	token, rec := s.LoginAndGetToken(user.Email, user.Passphrase)
	s.Require().Equal(http.StatusOK, rec.Code, "response: %s", rec.Body.String())
	s.Require().NotEmpty(token)
	user.Token = token
	return user
}

func (s *IntegrationSuite) MustCreateTestUser(role userdomain.Role) TestAuthUser {
	userID := uuid.New()
	email := fmt.Sprintf("%s-%s@example.com", role.String(), userID.String())
	passphrase := fmt.Sprintf("passphrase-%s", userID.String())
	name := fmt.Sprintf("Integration %s %s", strings.ToUpper(role.String()), userID.String()[:8])

	hash, err := bcrypt.GenerateFromPassword([]byte(passphrase), bcrypt.DefaultCost)
	s.Require().NoError(err)

	now := time.Now().UTC()
	_, err = s.UsersRepo.CreateUser(context.Background(), email, hash, func() (*userdomain.User, error) {
		return userdomain.NewUserWithDetails(userID, name, email, hash, role, nil, true, now, now)
	})
	s.Require().NoError(err)

	return TestAuthUser{
		UserID:     userID,
		Email:      email,
		Passphrase: passphrase,
		Role:       role,
	}
}

func (s *IntegrationSuite) Login(email, passphrase string) *httptest.ResponseRecorder {
	reqBody, err := json.Marshal(openapi.LoginRequest{
		Email:    openapi_types.Email(email),
		Password: passphrase,
	})
	s.Require().NoError(err)

	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	s.HTTPServer.ServeHTTP(rec, req)
	return rec
}

func (s *IntegrationSuite) GetTokenFromLoginResponse(rec *httptest.ResponseRecorder) string {
	if rec.Code != http.StatusOK {
		return ""
	}

	var resp openapi.LoginResponse
	err := json.Unmarshal(rec.Body.Bytes(), &resp)
	s.Require().NoError(err)

	return resp.Token
}

func (s *IntegrationSuite) LoginAndGetToken(email, passphrase string) (string, *httptest.ResponseRecorder) {
	rec := s.Login(email, passphrase)
	return s.GetTokenFromLoginResponse(rec), rec
}

func (s *IntegrationSuite) DefaultAdminToken() string {
	return s.defaultAdminToken
}

func (s *IntegrationSuite) ServeAuthenticatedHTTP(rec *httptest.ResponseRecorder, req *http.Request) {
	if strings.TrimSpace(req.Header.Get(echo.HeaderAuthorization)) == "" {
		s.Require().NotEmpty(s.defaultAdminToken)
		req.Header.Set(echo.HeaderAuthorization, fmt.Sprintf("Bearer %s", s.defaultAdminToken))
	}

	s.HTTPServer.ServeHTTP(rec, req)
}
