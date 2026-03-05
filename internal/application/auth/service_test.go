package auth_test

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"testing"
	"time"

	appauth "simpleservicedesk/internal/application/auth"
	authdomain "simpleservicedesk/internal/domain/auth"
	"simpleservicedesk/internal/domain/users"
	"simpleservicedesk/internal/queries"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

type mockUserRepository struct {
	users      []*users.User
	err        error
	getUserErr error
}

func (m mockUserRepository) ListUsers(_ context.Context, filter queries.UserFilter) ([]*users.User, error) {
	if m.err != nil {
		return nil, m.err
	}
	if filter.Email == nil {
		return m.users, nil
	}

	matcher, err := regexp.Compile("(?i)" + *filter.Email)
	if err != nil {
		return nil, fmt.Errorf("invalid email matcher: %w", err)
	}

	matched := make([]*users.User, 0, len(m.users))
	for _, user := range m.users {
		if matcher.MatchString(user.Email()) {
			matched = append(matched, user)
		}
	}

	return matched, nil
}

func (m mockUserRepository) GetUser(_ context.Context, userID uuid.UUID) (*users.User, error) {
	if m.getUserErr != nil {
		return nil, m.getUserErr
	}

	for _, user := range m.users {
		if user.ID() == userID {
			return user, nil
		}
	}

	return nil, users.ErrUserNotFound
}

func TestNewService(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		repo           appauth.UserRepository
		signingKey     string
		tokenExpiresIn time.Duration
		expectError    string
	}{
		{
			name:           "missing repository",
			repo:           nil,
			signingKey:     "app-signing-key",
			tokenExpiresIn: time.Hour,
			expectError:    "user repository is required",
		},
		{
			name:           "missing signing key",
			repo:           mockUserRepository{},
			signingKey:     "  ",
			tokenExpiresIn: time.Hour,
			expectError:    "jwt signing key is required",
		},
		{
			name:           "invalid expiration",
			repo:           mockUserRepository{},
			signingKey:     "app-signing-key",
			tokenExpiresIn: 0,
			expectError:    "jwt expiration must be greater than zero",
		},
		{
			name:           "valid config",
			repo:           mockUserRepository{},
			signingKey:     "app-signing-key",
			tokenExpiresIn: time.Hour,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			service, err := appauth.NewService(tc.repo, tc.signingKey, tc.tokenExpiresIn)
			if tc.expectError == "" {
				require.NoError(t, err)
				require.NotNil(t, service)
				return
			}

			require.ErrorContains(t, err, tc.expectError)
			require.Nil(t, service)
		})
	}
}

func TestServiceLoginSuccess(t *testing.T) {
	t.Parallel()

	user := createTestUser(t, "alice@example.com", users.RoleAgent, true)
	repo := mockUserRepository{users: []*users.User{user}}
	service := createTestService(t, repo)

	token, err := service.Login(context.Background(), "alice@example.com", "correct-password")
	require.NoError(t, err)
	require.NotEmpty(t, token)

	claims, err := service.ValidateToken(context.Background(), token)
	require.NoError(t, err)
	require.Equal(t, user.ID().String(), claims.UserID)
	require.Equal(t, user.Role(), claims.Role)
	require.Equal(t, user.ID().String(), claims.Subject)
}

func TestServiceLoginRequiresExactEmail(t *testing.T) {
	t.Parallel()

	user := createTestUser(t, "newalice@example.com", users.RoleCustomer, true)
	repo := mockUserRepository{users: []*users.User{user}}
	service := createTestService(t, repo)

	token, err := service.Login(context.Background(), "alice@example.com", "correct-password")
	require.ErrorIs(t, err, appauth.ErrInvalidCredentials)
	require.Empty(t, token)
}

func TestServiceLoginWrongPassword(t *testing.T) {
	t.Parallel()

	user := createTestUser(t, "alice@example.com", users.RoleCustomer, true)
	repo := mockUserRepository{users: []*users.User{user}}
	service := createTestService(t, repo)

	token, err := service.Login(context.Background(), "alice@example.com", "wrong-password")
	require.ErrorIs(t, err, appauth.ErrInvalidCredentials)
	require.Empty(t, token)
}

func TestServiceLoginDuplicateEmailMatches(t *testing.T) {
	t.Parallel()

	userA := createTestUser(t, "alice@example.com", users.RoleCustomer, true)
	userB := createTestUser(t, "alice@example.com", users.RoleAgent, true)
	repo := mockUserRepository{users: []*users.User{userA, userB}}
	service := createTestService(t, repo)

	token, err := service.Login(context.Background(), "alice@example.com", "correct-password")
	require.ErrorIs(t, err, appauth.ErrInvalidCredentials)
	require.Empty(t, token)
}

func TestServiceLoginInactiveUser(t *testing.T) {
	t.Parallel()

	user := createTestUser(t, "alice@example.com", users.RoleCustomer, false)
	repo := mockUserRepository{users: []*users.User{user}}
	service := createTestService(t, repo)

	token, err := service.Login(context.Background(), "alice@example.com", "correct-password")
	require.ErrorIs(t, err, appauth.ErrInvalidCredentials)
	require.Empty(t, token)
}

func TestServiceLoginRepositoryError(t *testing.T) {
	t.Parallel()

	repo := mockUserRepository{err: errors.New("db unavailable")}
	service := createTestService(t, repo)

	token, err := service.Login(context.Background(), "alice@example.com", "correct-password")
	require.ErrorContains(t, err, "failed to find user by email")
	require.Empty(t, token)
}

func TestServiceGenerateTokenRequiresUser(t *testing.T) {
	t.Parallel()

	service := createTestService(t, mockUserRepository{})

	token, err := service.GenerateToken(nil)
	require.ErrorContains(t, err, "user is required")
	require.Empty(t, token)
}

func TestServiceValidateToken(t *testing.T) {
	t.Parallel()

	user := createTestUser(t, "alice@example.com", users.RoleAdmin, true)
	service := createTestService(t, mockUserRepository{users: []*users.User{user}})

	token, err := service.GenerateToken(user)
	require.NoError(t, err)

	claims, err := service.ValidateToken(context.Background(), token)
	require.NoError(t, err)
	require.Equal(t, user.ID().String(), claims.UserID)
	require.Equal(t, user.Role(), claims.Role)
}

func TestServiceValidateTokenInactiveUser(t *testing.T) {
	t.Parallel()

	user := createTestUser(t, "inactive@example.com", users.RoleCustomer, false)
	service := createTestService(t, mockUserRepository{users: []*users.User{user}})

	token, err := service.GenerateToken(user)
	require.NoError(t, err)

	claims, err := service.ValidateToken(context.Background(), token)
	require.ErrorIs(t, err, appauth.ErrInvalidToken)
	require.Nil(t, claims)
}

func TestServiceValidateTokenRepositoryError(t *testing.T) {
	t.Parallel()

	user := createTestUser(t, "alice@example.com", users.RoleCustomer, true)
	repo := mockUserRepository{
		users:      []*users.User{user},
		getUserErr: errors.New("db unavailable"),
	}
	service := createTestService(t, repo)

	token, err := service.GenerateToken(user)
	require.NoError(t, err)

	claims, err := service.ValidateToken(context.Background(), token)
	require.ErrorIs(t, err, appauth.ErrInvalidToken)
	require.Nil(t, claims)
	require.ErrorContains(t, err, "user not found")
}

func TestServiceValidateTokenRoleMismatch(t *testing.T) {
	t.Parallel()

	userID := uuid.New()
	passwordHash, err := bcrypt.GenerateFromPassword([]byte("correct-password"), bcrypt.DefaultCost)
	require.NoError(t, err)

	repoUser, err := users.NewUserWithDetails(
		userID,
		"Role Changed",
		"role.changed@example.com",
		passwordHash,
		users.RoleCustomer,
		nil,
		true,
		time.Now(),
		time.Now(),
	)
	require.NoError(t, err)

	service := createTestServiceWithKey(t, mockUserRepository{users: []*users.User{repoUser}}, "signing-key-a")
	tokenString, err := signCustomClaims(userID, users.RoleAdmin, time.Now().Add(time.Hour), []byte("signing-key-a"))
	require.NoError(t, err)

	claims, err := service.ValidateToken(context.Background(), tokenString)
	require.ErrorIs(t, err, appauth.ErrInvalidToken)
	require.Nil(t, claims)
}

func TestServiceValidateTokenInvalidToken(t *testing.T) {
	t.Parallel()

	service := createTestService(t, mockUserRepository{})

	claims, err := service.ValidateToken(context.Background(), "not-a-jwt")
	require.ErrorIs(t, err, appauth.ErrInvalidToken)
	require.Nil(t, claims)
}

func TestServiceValidateTokenEmptyToken(t *testing.T) {
	t.Parallel()

	service := createTestService(t, mockUserRepository{})

	claims, err := service.ValidateToken(context.Background(), "")
	require.ErrorIs(t, err, appauth.ErrInvalidToken)
	require.Nil(t, claims)
}

func TestServiceValidateTokenWrongSigningKey(t *testing.T) {
	t.Parallel()

	user := createTestUser(t, "alice@example.com", users.RoleAdmin, true)
	serviceA := createTestServiceWithKey(t, mockUserRepository{}, "signing-key-a")
	serviceB := createTestServiceWithKey(t, mockUserRepository{}, "signing-key-b")

	token, err := serviceA.GenerateToken(user)
	require.NoError(t, err)

	claims, err := serviceB.ValidateToken(context.Background(), token)
	require.ErrorIs(t, err, appauth.ErrInvalidToken)
	require.Nil(t, claims)
}

func TestServiceValidateTokenInvalidRole(t *testing.T) {
	t.Parallel()

	service := createTestServiceWithKey(t, mockUserRepository{}, "signing-key-a")
	userID := uuid.New()

	tokenString, err := signCustomClaims(
		userID,
		users.Role("invalid-role"),
		time.Now().Add(time.Hour),
		[]byte("signing-key-a"),
	)
	require.NoError(t, err)

	claims, err := service.ValidateToken(context.Background(), tokenString)
	require.ErrorIs(t, err, appauth.ErrInvalidToken)
	require.Nil(t, claims)
}

func TestServiceValidateTokenInvalidUserIDClaim(t *testing.T) {
	t.Parallel()

	service := createTestServiceWithKey(t, mockUserRepository{}, "signing-key-a")

	tokenString, err := signCustomClaimsWithUserID(
		"not-a-uuid",
		users.RoleAdmin,
		time.Now().Add(time.Hour),
		jwt.SigningMethodHS256,
		[]byte("signing-key-a"),
	)
	require.NoError(t, err)

	claims, err := service.ValidateToken(context.Background(), tokenString)
	require.ErrorIs(t, err, appauth.ErrInvalidToken)
	require.Nil(t, claims)
}

func TestServiceValidateTokenUnexpectedSigningAlgorithm(t *testing.T) {
	t.Parallel()

	service := createTestServiceWithKey(t, mockUserRepository{}, "signing-key-a")

	tokenString, err := signCustomClaimsWithUserID(
		uuid.NewString(),
		users.RoleAgent,
		time.Now().Add(time.Hour),
		jwt.SigningMethodHS384,
		[]byte("signing-key-a"),
	)
	require.NoError(t, err)

	claims, err := service.ValidateToken(context.Background(), tokenString)
	require.ErrorIs(t, err, appauth.ErrInvalidToken)
	require.Nil(t, claims)
}

func TestServiceValidateTokenExpired(t *testing.T) {
	t.Parallel()

	service := createTestServiceWithKey(t, mockUserRepository{}, "signing-key-a")
	userID := uuid.New()

	tokenString, err := signCustomClaims(
		userID,
		users.RoleAgent,
		time.Now().Add(-time.Minute),
		[]byte("signing-key-a"),
	)
	require.NoError(t, err)

	claims, err := service.ValidateToken(context.Background(), tokenString)
	require.ErrorIs(t, err, appauth.ErrInvalidToken)
	require.Nil(t, claims)
}

func createTestService(t *testing.T, repo appauth.UserRepository) *appauth.Service {
	t.Helper()

	return createTestServiceWithKey(t, repo, "test-signing-key")
}

func createTestServiceWithKey(t *testing.T, repo appauth.UserRepository, signingKey string) *appauth.Service {
	t.Helper()

	service, err := appauth.NewService(repo, signingKey, time.Hour)
	require.NoError(t, err)

	return service
}

func createTestUser(
	t *testing.T,
	email string,
	role users.Role,
	isActive bool,
) *users.User {
	t.Helper()

	passwordHash, err := bcrypt.GenerateFromPassword([]byte("correct-password"), bcrypt.DefaultCost)
	require.NoError(t, err)

	user, err := users.NewUserWithDetails(
		uuid.New(),
		"Test User",
		email,
		passwordHash,
		role,
		nil,
		isActive,
		time.Now(),
		time.Now(),
	)
	require.NoError(t, err)

	return user
}

func signCustomClaims(userID uuid.UUID, role users.Role, expiresAt time.Time, signingKey []byte) (string, error) {
	return signCustomClaimsWithUserID(userID.String(), role, expiresAt, jwt.SigningMethodHS256, signingKey)
}

func signCustomClaimsWithUserID(
	userID string,
	role users.Role,
	expiresAt time.Time,
	method jwt.SigningMethod,
	signingKey []byte,
) (string, error) {
	claims := authdomain.Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-time.Minute)),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
		UserID: userID,
		Role:   role,
	}

	token := jwt.NewWithClaims(method, claims)
	return token.SignedString(signingKey)
}
