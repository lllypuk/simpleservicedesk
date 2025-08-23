package application

import (
	"context"

	"simpleservicedesk/internal/domain/tickets"
	"simpleservicedesk/internal/domain/users"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/suite"
)

type ServerSuite struct {
	suite.Suite

	HTTPServer  *echo.Echo
	UsersRepo   UserRepository   // Interface for repository
	TicketsRepo TicketRepository // Interface for ticket repository
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

// mockTicketRepository is a simple mock for testing
type mockTicketRepository struct {
	tickets map[uuid.UUID]*tickets.Ticket
}

func newMockTicketRepository() *mockTicketRepository {
	return &mockTicketRepository{
		tickets: make(map[uuid.UUID]*tickets.Ticket),
	}
}

func (m *mockTicketRepository) CreateTicket(
	_ context.Context,
	createFn func() (*tickets.Ticket, error),
) (*tickets.Ticket, error) {
	ticket, err := createFn()
	if err != nil {
		return nil, err
	}
	m.tickets[ticket.ID()] = ticket
	return ticket, nil
}

func (m *mockTicketRepository) UpdateTicket(
	_ context.Context,
	id uuid.UUID,
	updateFn func(*tickets.Ticket) (bool, error),
) (*tickets.Ticket, error) {
	ticket, exists := m.tickets[id]
	if !exists {
		return nil, tickets.ErrTicketNotFound
	}
	updated, err := updateFn(ticket)
	if err != nil {
		return nil, err
	}
	if updated {
		m.tickets[id] = ticket
	}
	return ticket, nil
}

func (m *mockTicketRepository) GetTicket(_ context.Context, id uuid.UUID) (*tickets.Ticket, error) {
	ticket, exists := m.tickets[id]
	if !exists {
		return nil, tickets.ErrTicketNotFound
	}
	return ticket, nil
}

func (m *mockTicketRepository) ListTickets(
	_ context.Context,
	_ TicketFilter,
) ([]*tickets.Ticket, error) {
	result := make([]*tickets.Ticket, 0, len(m.tickets))
	for _, ticket := range m.tickets {
		result = append(result, ticket)
	}
	return result, nil
}

func (m *mockTicketRepository) DeleteTicket(_ context.Context, id uuid.UUID) error {
	_, exists := m.tickets[id]
	if !exists {
		return tickets.ErrTicketNotFound
	}
	delete(m.tickets, id)
	return nil
}

// SetupTest for integration tests
func (s *ServerSuite) SetupTest() {
	// Initialize mock repositories with fresh state
	s.UsersRepo = newMockUserRepository()
	s.TicketsRepo = newMockTicketRepository()

	// Initialize HTTP server with mock repositories
	s.HTTPServer = SetupHTTPServer(s.UsersRepo, s.TicketsRepo)
}
