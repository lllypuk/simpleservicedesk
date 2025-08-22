package application

import (
	"context"

	"simpleservicedesk/internal/domain/users"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/suite"
)

type ServerSuite struct {
	suite.Suite

	HTTPServer *echo.Echo
	UsersRepo  UserRepository // Interface for repository
}

// mockUserRepository is a simple mock for testing
type mockUserRepository struct {
	createdEmails map[string]bool
	users         map[uuid.UUID]*users.User
}

func newMockUserRepository() *mockUserRepository {
	return &mockUserRepository{
		createdEmails: make(map[string]bool),
		users:         make(map[uuid.UUID]*users.User),
	}
}

func (m *mockUserRepository) CreateUser(
	_ context.Context,
	email string,
	_ []byte,
	createFn func() (*users.User, error),
) (*users.User, error) {
	// Check for duplicate email
	if m.createdEmails[email] {
		return nil, users.ErrUserAlreadyExist
	}

	user, err := createFn()
	if err != nil {
		return nil, err
	}

	// Track the email and store the user
	m.createdEmails[email] = true
	m.users[user.ID()] = user

	return user, nil
}

func (m *mockUserRepository) UpdateUser(
	_ context.Context,
	id uuid.UUID,
	updateFn func(*users.User) (bool, error),
) (*users.User, error) {
	// Simple mock - just create a dummy user and call updateFn
	user, _ := users.NewUser(id, "Test User", "test@example.com", []byte("hash"))
	updated, err := updateFn(user)
	if err != nil {
		return nil, err
	}
	if updated {
		return user, nil
	}
	return user, nil
}

func (m *mockUserRepository) GetUser(_ context.Context, id uuid.UUID) (*users.User, error) {
	// Return actual stored user or error if not found
	user, exists := m.users[id]
	if !exists {
		return nil, users.ErrUserNotFound
	}
	return user, nil
}

// SetupTest for integration tests
func (s *ServerSuite) SetupTest() {
	// Initialize mock repository with fresh state
	s.UsersRepo = newMockUserRepository()

	// Initialize HTTP server with mock repository
	s.HTTPServer = SetupHTTPServer(s.UsersRepo)
}
