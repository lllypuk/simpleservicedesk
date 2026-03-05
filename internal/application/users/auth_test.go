package users_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"simpleservicedesk/generated/openapi"
	appusers "simpleservicedesk/internal/application/users"
	authdomain "simpleservicedesk/internal/domain/auth"
	userdomain "simpleservicedesk/internal/domain/users"
	"simpleservicedesk/internal/queries"
	"simpleservicedesk/pkg/contextkeys"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
)

type userRepoSpy struct {
	getUserCalls    int
	updateUserCalls int
}

func (r *userRepoSpy) CreateUser(
	_ context.Context,
	_ string,
	_ []byte,
	_ func() (*userdomain.User, error),
) (*userdomain.User, error) {
	panic("unexpected CreateUser call")
}

func (r *userRepoSpy) UpdateUser(
	_ context.Context,
	_ uuid.UUID,
	_ func(*userdomain.User) (bool, error),
) (*userdomain.User, error) {
	r.updateUserCalls++
	return nil, userdomain.ErrUserNotFound
}

func (r *userRepoSpy) GetUser(_ context.Context, _ uuid.UUID) (*userdomain.User, error) {
	r.getUserCalls++
	return nil, userdomain.ErrUserNotFound
}

func (r *userRepoSpy) ListUsers(_ context.Context, _ queries.UserFilter) ([]*userdomain.User, error) {
	panic("unexpected ListUsers call")
}

func (r *userRepoSpy) DeleteUser(_ context.Context, _ uuid.UUID) error {
	panic("unexpected DeleteUser call")
}

func (r *userRepoSpy) CountUsers(_ context.Context, _ queries.UserFilter) (int64, error) {
	panic("unexpected CountUsers call")
}

func TestDeleteUsersIDRequiresAdminInHandler(t *testing.T) {
	tests := []struct {
		name         string
		claims       *authdomain.Claims
		expectedCode int
	}{
		{
			name:         "missing claims returns 401",
			claims:       nil,
			expectedCode: http.StatusUnauthorized,
		},
		{
			name: "non-admin role returns 403",
			claims: &authdomain.Claims{
				UserID: uuid.NewString(),
				Role:   userdomain.RoleCustomer,
			},
			expectedCode: http.StatusForbidden,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := &userRepoSpy{}
			handlers := appusers.SetupHandlers(repo)
			e := echo.New()

			req := httptest.NewRequest(http.MethodDelete, "/users/"+uuid.NewString(), nil)
			req = withUserClaims(req, tc.claims)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			err := handlers.DeleteUsersID(c, uuid.New())
			require.NoError(t, err)
			require.Equal(t, tc.expectedCode, rec.Code)
			require.Zero(t, repo.getUserCalls)
			require.Zero(t, repo.updateUserCalls)
		})
	}
}

func TestPatchUsersIDRoleRequiresAdminInHandler(t *testing.T) {
	tests := []struct {
		name         string
		claims       *authdomain.Claims
		expectedCode int
	}{
		{
			name:         "missing claims returns 401",
			claims:       nil,
			expectedCode: http.StatusUnauthorized,
		},
		{
			name: "non-admin role returns 403",
			claims: &authdomain.Claims{
				UserID: uuid.NewString(),
				Role:   userdomain.RoleAgent,
			},
			expectedCode: http.StatusForbidden,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := &userRepoSpy{}
			handlers := appusers.SetupHandlers(repo)
			e := echo.New()

			payload, err := json.Marshal(openapi.UpdateUserRoleRequest{Role: openapi.Agent})
			require.NoError(t, err)

			req := httptest.NewRequest(
				http.MethodPatch,
				"/users/"+uuid.NewString()+"/role",
				bytes.NewReader(payload),
			)
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			req = withUserClaims(req, tc.claims)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			err = handlers.PatchUsersIDRole(c, uuid.New())
			require.NoError(t, err)
			require.Equal(t, tc.expectedCode, rec.Code)
			require.Zero(t, repo.getUserCalls)
			require.Zero(t, repo.updateUserCalls)
		})
	}
}

func withUserClaims(req *http.Request, claims *authdomain.Claims) *http.Request {
	if claims == nil {
		return req
	}

	ctx := context.WithValue(req.Context(), contextkeys.AuthClaimsCtxKey, claims)
	return req.WithContext(ctx)
}
