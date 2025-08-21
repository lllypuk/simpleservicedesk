package mocks

import (
	"context"
	"sync"

	"simpleservicedesk/internal/domain/users"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

// UserRepositoryMock is a mock implementation of UserRepository interface
type UserRepositoryMock struct {
	mock.Mock

	mu sync.RWMutex
	// In-memory storage for testing
	users  map[uuid.UUID]*users.User
	emails map[string]uuid.UUID // email -> userID mapping for duplicate check

	// Flag to determine if we should use expectations or default behavior
	useDefaultBehavior bool
}

// NewUserRepositoryMock creates a new instance of UserRepositoryMock
func NewUserRepositoryMock() *UserRepositoryMock {
	return &UserRepositoryMock{
		users:              make(map[uuid.UUID]*users.User),
		emails:             make(map[string]uuid.UUID),
		useDefaultBehavior: true, // By default, use realistic behavior
	}
}

// EnableMockExpectations disables default behavior and requires explicit expectations
func (m *UserRepositoryMock) EnableMockExpectations() {
	m.useDefaultBehavior = false
}

// CreateUser mocks the CreateUser method
func (m *UserRepositoryMock) CreateUser(
	ctx context.Context,
	email string,
	passwordHash []byte,
	createFn func() (*users.User, error),
) (*users.User, error) {
	if !m.useDefaultBehavior {
		// Use testify mock expectations
		args := m.Called(ctx, email, passwordHash, createFn)
		user, ok := args.Get(0).(*users.User)
		if !ok || user == nil {
			return nil, args.Error(1)
		}
		return user, args.Error(1)
	}

	// Default behavior: simulate real repository logic
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if email already exists
	if _, exists := m.emails[email]; exists {
		return nil, users.ErrUserAlreadyExist
	}

	// Call the create function
	user, err := createFn()
	if err != nil {
		return nil, err
	}

	// Store the user
	m.users[user.ID()] = user
	m.emails[email] = user.ID()

	return user, nil
}

// UpdateUser mocks the UpdateUser method
func (m *UserRepositoryMock) UpdateUser(
	ctx context.Context,
	id uuid.UUID,
	updateFn func(*users.User) (bool, error),
) (*users.User, error) {
	if !m.useDefaultBehavior {
		// Use testify mock expectations
		args := m.Called(ctx, id, updateFn)
		user, ok := args.Get(0).(*users.User)
		if !ok || user == nil {
			return nil, args.Error(1)
		}
		return user, args.Error(1)
	}

	// Default behavior: simulate real repository logic
	m.mu.Lock()
	defer m.mu.Unlock()

	user, exists := m.users[id]
	if !exists {
		return nil, users.ErrUserNotFound
	}

	// Create a copy to avoid modifying the original
	userCopy, err := users.NewUser(user.ID(), user.Name(), user.Email(), []byte("dummy"))
	if err != nil {
		return nil, err
	}

	updated, err := updateFn(userCopy)
	if err != nil {
		return nil, err
	}

	if updated {
		// Update email mapping if email changed
		if user.Email() != userCopy.Email() {
			delete(m.emails, user.Email())
			m.emails[userCopy.Email()] = id
		}
		m.users[id] = userCopy
	}

	return userCopy, nil
}

// GetUser mocks the GetUser method
func (m *UserRepositoryMock) GetUser(ctx context.Context, id uuid.UUID) (*users.User, error) {
	if !m.useDefaultBehavior {
		// Use testify mock expectations
		args := m.Called(ctx, id)
		user, ok := args.Get(0).(*users.User)
		if !ok || user == nil {
			return nil, args.Error(1)
		}
		return user, args.Error(1)
	}

	// Default behavior: simulate real repository logic
	m.mu.RLock()
	defer m.mu.RUnlock()

	user, exists := m.users[id]
	if !exists {
		return nil, users.ErrUserNotFound
	}

	return user, nil
}

// Reset clears all stored data and mock expectations
func (m *UserRepositoryMock) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.users = make(map[uuid.UUID]*users.User)
	m.emails = make(map[string]uuid.UUID)
	m.Mock = mock.Mock{}
	m.useDefaultBehavior = true
}

// GetAllUsers returns all stored users (helper for testing)
func (m *UserRepositoryMock) GetAllUsers() map[uuid.UUID]*users.User {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make(map[uuid.UUID]*users.User)
	for k, v := range m.users {
		result[k] = v
	}
	return result
}

// SetUser manually sets a user in the mock storage (helper for testing)
func (m *UserRepositoryMock) SetUser(user *users.User) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.users[user.ID()] = user
	m.emails[user.Email()] = user.ID()
}
