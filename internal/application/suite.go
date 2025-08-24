package application

import (
	"context"
	"strings"

	"simpleservicedesk/internal/domain/categories"
	"simpleservicedesk/internal/domain/organizations"
	"simpleservicedesk/internal/domain/tickets"
	"simpleservicedesk/internal/domain/users"
	infraUsers "simpleservicedesk/internal/infrastructure/users"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/suite"
)

type ServerSuite struct {
	suite.Suite

	HTTPServer        *echo.Echo
	UsersRepo         UserRepository         // Interface for repository
	TicketsRepo       TicketRepository       // Interface for ticket repository
	OrganizationsRepo OrganizationRepository // Interface for organization repository
	CategoriesRepo    CategoryRepository     // Interface for category repository
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
	// Check if user exists in our mock store
	user, exists := m.users[id]
	if !exists {
		return nil, users.ErrUserNotFound
	}

	updated, err := updateFn(user)
	if err != nil {
		return nil, err
	}
	if updated {
		// Store updated user back
		m.users[id] = user
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

func (m *mockUserRepository) ListUsers(_ context.Context, filter infraUsers.UserFilter) ([]*users.User, error) {
	var result []*users.User
	count := 0

	for _, user := range m.users {
		if !m.userMatchesFilter(user, filter) {
			continue
		}

		if m.shouldLimitResults(filter, count) {
			break
		}
		if m.shouldIncludeInResults(filter, count) {
			result = append(result, user)
		}
		count++
	}

	return result, nil
}

func (m *mockUserRepository) userMatchesFilter(user *users.User, filter infraUsers.UserFilter) bool {
	if filter.Name != "" && !strings.Contains(strings.ToLower(user.Name()), strings.ToLower(filter.Name)) {
		return false
	}
	if filter.Email != "" && !strings.Contains(strings.ToLower(user.Email()), strings.ToLower(filter.Email)) {
		return false
	}
	if filter.Role != nil && user.Role() != *filter.Role {
		return false
	}
	if filter.OrganizationID != nil {
		if user.OrganizationID() == nil || *user.OrganizationID() != *filter.OrganizationID {
			return false
		}
	}
	if filter.IsActive != nil && user.IsActive() != *filter.IsActive {
		return false
	}
	return true
}

func (m *mockUserRepository) shouldLimitResults(filter infraUsers.UserFilter, count int) bool {
	return filter.Limit > 0 && count >= filter.Offset+filter.Limit
}

func (m *mockUserRepository) shouldIncludeInResults(filter infraUsers.UserFilter, count int) bool {
	return count >= filter.Offset
}

func (m *mockUserRepository) DeleteUser(_ context.Context, id uuid.UUID) error {
	_, exists := m.users[id]
	if !exists {
		return users.ErrUserNotFound
	}
	delete(m.users, id)
	return nil
}

func (m *mockUserRepository) CountUsers(_ context.Context, filter infraUsers.UserFilter) (int64, error) {
	count := int64(0)

	for _, user := range m.users {
		// Apply the same filtering as ListUsers
		if filter.Name != "" && !strings.Contains(strings.ToLower(user.Name()), strings.ToLower(filter.Name)) {
			continue
		}
		if filter.Email != "" && !strings.Contains(strings.ToLower(user.Email()), strings.ToLower(filter.Email)) {
			continue
		}
		if filter.Role != nil && user.Role() != *filter.Role {
			continue
		}
		if filter.OrganizationID != nil {
			if user.OrganizationID() == nil || *user.OrganizationID() != *filter.OrganizationID {
				continue
			}
		}
		if filter.IsActive != nil && user.IsActive() != *filter.IsActive {
			continue
		}
		count++
	}

	return count, nil
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

// mockOrganizationRepository is a simple mock for testing
type mockOrganizationRepository struct {
	orgs map[uuid.UUID]*organizations.Organization
}

func newMockOrganizationRepository() *mockOrganizationRepository {
	return &mockOrganizationRepository{
		orgs: make(map[uuid.UUID]*organizations.Organization),
	}
}

func (m *mockOrganizationRepository) CreateOrganization(
	_ context.Context,
	createFn func() (*organizations.Organization, error),
) (*organizations.Organization, error) {
	org, err := createFn()
	if err != nil {
		return nil, err
	}

	m.orgs[org.ID()] = org
	return org, nil
}

func (m *mockOrganizationRepository) UpdateOrganization(
	_ context.Context,
	id uuid.UUID,
	updateFn func(*organizations.Organization) (bool, error),
) (*organizations.Organization, error) {
	org, exists := m.orgs[id]
	if !exists {
		return nil, organizations.ErrOrganizationNotFound
	}

	_, err := updateFn(org)
	if err != nil {
		return nil, err
	}

	return org, nil
}

func (m *mockOrganizationRepository) GetOrganization(
	_ context.Context,
	id uuid.UUID,
) (*organizations.Organization, error) {
	org, exists := m.orgs[id]
	if !exists {
		return nil, organizations.ErrOrganizationNotFound
	}
	return org, nil
}

func (m *mockOrganizationRepository) ListOrganizations(
	_ context.Context,
	filter OrganizationFilter,
) ([]*organizations.Organization, error) {
	var result []*organizations.Organization
	count := 0

	for _, org := range m.orgs {
		if filter.Limit > 0 && count >= filter.Offset+filter.Limit {
			break
		}
		if count >= filter.Offset {
			result = append(result, org)
		}
		count++
	}

	return result, nil
}

func (m *mockOrganizationRepository) DeleteOrganization(_ context.Context, id uuid.UUID) error {
	_, exists := m.orgs[id]
	if !exists {
		return organizations.ErrOrganizationNotFound
	}
	delete(m.orgs, id)
	return nil
}

func (m *mockOrganizationRepository) GetOrganizationHierarchy(
	_ context.Context,
	rootID uuid.UUID,
) (*OrganizationTree, error) {
	root, exists := m.orgs[rootID]
	if !exists {
		return nil, organizations.ErrOrganizationNotFound
	}

	return &OrganizationTree{
		Organization: root,
		Children:     []*OrganizationTree{}, // Simplified implementation for tests
	}, nil
}

// mockCategoryRepository is a simple mock for testing
type mockCategoryRepository struct {
	categories map[uuid.UUID]*categories.Category
}

func newMockCategoryRepository() *mockCategoryRepository {
	return &mockCategoryRepository{
		categories: make(map[uuid.UUID]*categories.Category),
	}
}

func (m *mockCategoryRepository) CreateCategory(
	_ context.Context,
	createFn func() (*categories.Category, error),
) (*categories.Category, error) {
	category, err := createFn()
	if err != nil {
		return nil, err
	}

	m.categories[category.ID()] = category
	return category, nil
}

func (m *mockCategoryRepository) UpdateCategory(
	_ context.Context,
	id uuid.UUID,
	updateFn func(*categories.Category) (bool, error),
) (*categories.Category, error) {
	category, exists := m.categories[id]
	if !exists {
		return nil, categories.ErrCategoryNotFound
	}

	_, err := updateFn(category)
	if err != nil {
		return nil, err
	}

	return category, nil
}

func (m *mockCategoryRepository) GetCategory(
	_ context.Context,
	id uuid.UUID,
) (*categories.Category, error) {
	category, exists := m.categories[id]
	if !exists {
		return nil, categories.ErrCategoryNotFound
	}
	return category, nil
}

func (m *mockCategoryRepository) ListCategories(
	_ context.Context,
	filter CategoryFilter,
) ([]*categories.Category, error) {
	var result []*categories.Category
	count := 0

	for _, category := range m.categories {
		if !m.matchesFilter(category, filter) {
			continue
		}

		if m.shouldBreakOnLimit(filter, count) {
			break
		}
		if m.shouldIncludeInResult(filter, count) {
			result = append(result, category)
		}
		count++
	}

	return result, nil
}

func (m *mockCategoryRepository) matchesFilter(category *categories.Category, filter CategoryFilter) bool {
	if filter.OrganizationID != nil && category.OrganizationID() != *filter.OrganizationID {
		return false
	}
	if filter.IsActive != nil && category.IsActive() != *filter.IsActive {
		return false
	}
	return m.matchesParentFilter(category, filter)
}

func (m *mockCategoryRepository) matchesParentFilter(category *categories.Category, filter CategoryFilter) bool {
	if filter.ParentID == nil {
		return true
	}
	if category.ParentID() == nil && *filter.ParentID != uuid.Nil {
		return false
	}
	if category.ParentID() != nil && *category.ParentID() != *filter.ParentID {
		return false
	}
	return true
}

func (m *mockCategoryRepository) shouldBreakOnLimit(filter CategoryFilter, count int) bool {
	return filter.Limit > 0 && count >= filter.Offset+filter.Limit
}

func (m *mockCategoryRepository) shouldIncludeInResult(filter CategoryFilter, count int) bool {
	return count >= filter.Offset
}

func (m *mockCategoryRepository) DeleteCategory(_ context.Context, id uuid.UUID) error {
	_, exists := m.categories[id]
	if !exists {
		return categories.ErrCategoryNotFound
	}
	delete(m.categories, id)
	return nil
}

func (m *mockCategoryRepository) GetCategoryHierarchy(
	_ context.Context,
	rootID uuid.UUID,
) (*CategoryTree, error) {
	root, exists := m.categories[rootID]
	if !exists {
		return nil, categories.ErrCategoryNotFound
	}

	return &CategoryTree{
		Category: root,
		Children: []*CategoryTree{}, // Simplified implementation for tests
	}, nil
}

// SetupTest for integration tests
func (s *ServerSuite) SetupTest() {
	// Initialize mock repositories with fresh state
	s.UsersRepo = newMockUserRepository()
	s.TicketsRepo = newMockTicketRepository()
	s.OrganizationsRepo = newMockOrganizationRepository()
	s.CategoriesRepo = newMockCategoryRepository()

	// Initialize HTTP server with mock repositories
	s.HTTPServer = SetupHTTPServer(s.UsersRepo, s.TicketsRepo, s.OrganizationsRepo, s.CategoriesRepo)
}
