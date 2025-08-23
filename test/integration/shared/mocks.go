//go:build integration
// +build integration

package shared

import (
	"context"

	"simpleservicedesk/internal/application"
	"simpleservicedesk/internal/domain/categories"
	"simpleservicedesk/internal/domain/organizations"
	"simpleservicedesk/internal/domain/tickets"
	"simpleservicedesk/internal/domain/users"

	"github.com/google/uuid"
)

// mockUserRepository is a simple mock for integration testing
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

// mockCategoryRepository is a simple mock for integration testing
type mockCategoryRepository struct {
	categories  map[uuid.UUID]*categories.Category
	nameIndex   map[string]uuid.UUID      // organization_id:name -> category_id
	parentIndex map[uuid.UUID][]uuid.UUID // parent_id -> []child_ids
}

func newMockCategoryRepository() *mockCategoryRepository {
	return &mockCategoryRepository{
		categories:  make(map[uuid.UUID]*categories.Category),
		nameIndex:   make(map[string]uuid.UUID),
		parentIndex: make(map[uuid.UUID][]uuid.UUID),
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

	// Check for duplicate name in same organization
	nameKey := category.OrganizationID().String() + ":" + category.Name()
	if _, exists := m.nameIndex[nameKey]; exists {
		return nil, categories.ErrCategoryAlreadyExist
	}

	// Check parent exists if provided
	if category.HasParent() {
		if _, exists := m.categories[*category.ParentID()]; !exists {
			return nil, categories.ErrCategoryNotFound
		}
	}

	// Store category
	m.categories[category.ID()] = category
	m.nameIndex[nameKey] = category.ID()

	// Update parent index
	if category.HasParent() {
		parentID := *category.ParentID()
		m.parentIndex[parentID] = append(m.parentIndex[parentID], category.ID())
	}

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

	// Create a copy to avoid modifying original during validation
	categoryClone, err := categories.NewCategory(
		category.ID(),
		category.Name(),
		category.Description(),
		category.OrganizationID(),
		category.ParentID(),
	)
	if err != nil {
		return nil, err
	}

	// Restore activation state
	if !category.IsActive() {
		categoryClone.Deactivate()
	}

	updated, err := updateFn(categoryClone)
	if err != nil {
		return nil, err
	}
	if !updated {
		return categoryClone, nil
	}

	// Check for circular reference if parent changed
	if categoryClone.HasParent() {
		if m.wouldCreateCircle(id, *categoryClone.ParentID()) {
			return nil, categories.ErrCircularReference
		}
	}

	// Update storage
	m.categories[id] = categoryClone

	return categoryClone, nil
}

func (m *mockCategoryRepository) GetCategory(_ context.Context, id uuid.UUID) (*categories.Category, error) {
	category, exists := m.categories[id]
	if !exists {
		return nil, categories.ErrCategoryNotFound
	}
	return category, nil
}

func (m *mockCategoryRepository) ListCategories(
	_ context.Context,
	filter application.CategoryFilter,
) ([]*categories.Category, error) {
	var result []*categories.Category

	for _, category := range m.categories {
		// Apply filters
		if filter.OrganizationID != nil && category.OrganizationID() != *filter.OrganizationID {
			continue
		}
		if filter.ParentID != nil {
			if category.ParentID() == nil || *category.ParentID() != *filter.ParentID {
				continue
			}
		}
		if filter.IsActive != nil && category.IsActive() != *filter.IsActive {
			continue
		}
		if filter.IsRootOnly && category.HasParent() {
			continue
		}
		if filter.Name != nil && category.Name() != *filter.Name {
			continue
		}

		result = append(result, category)
	}

	// Apply pagination
	if filter.Offset > 0 && filter.Offset < len(result) {
		result = result[filter.Offset:]
	} else if filter.Offset >= len(result) {
		result = []*categories.Category{}
	}

	if filter.Limit > 0 && filter.Limit < len(result) {
		result = result[:filter.Limit]
	}

	return result, nil
}

func (m *mockCategoryRepository) GetCategoryHierarchy(
	ctx context.Context,
	rootID uuid.UUID,
) (*application.CategoryTree, error) {
	root, err := m.GetCategory(ctx, rootID)
	if err != nil {
		return nil, err
	}

	tree := &application.CategoryTree{
		Category: root,
		Children: []*application.CategoryTree{},
	}

	// Get children and build their hierarchies
	childIDs := m.parentIndex[rootID]
	for _, childID := range childIDs {
		childTree, err := m.GetCategoryHierarchy(ctx, childID)
		if err != nil {
			return nil, err
		}
		tree.Children = append(tree.Children, childTree)
	}

	return tree, nil
}

func (m *mockCategoryRepository) DeleteCategory(_ context.Context, id uuid.UUID) error {
	category, exists := m.categories[id]
	if !exists {
		return categories.ErrCategoryNotFound
	}

	// Check for children
	if children := m.parentIndex[id]; len(children) > 0 {
		return categories.ErrInvalidCategory // Cannot delete category with children
	}

	// Remove from storage
	delete(m.categories, id)

	// Remove from name index
	nameKey := category.OrganizationID().String() + ":" + category.Name()
	delete(m.nameIndex, nameKey)

	// Remove from parent index
	if category.HasParent() {
		parentID := *category.ParentID()
		children := m.parentIndex[parentID]
		for i, childID := range children {
			if childID == id {
				m.parentIndex[parentID] = append(children[:i], children[i+1:]...)
				break
			}
		}
	}

	return nil
}

// wouldCreateCircle checks if setting newParentID as parent would create circular reference
func (m *mockCategoryRepository) wouldCreateCircle(categoryID, newParentID uuid.UUID) bool {
	if categoryID == newParentID {
		return true
	}

	visited := make(map[uuid.UUID]bool)
	currentID := newParentID

	for {
		if visited[currentID] {
			break // Found cycle but not involving categoryID
		}
		visited[currentID] = true

		if currentID == categoryID {
			return true // Found circle
		}

		category, exists := m.categories[currentID]
		if !exists || !category.HasParent() {
			break
		}
		currentID = *category.ParentID()
	}

	return false
}

// mockOrganizationRepository is a simple mock for integration testing
type mockOrganizationRepository struct {
	organizations map[uuid.UUID]*organizations.Organization
	nameIndex     map[string]uuid.UUID      // name -> organization_id
	parentIndex   map[uuid.UUID][]uuid.UUID // parent_id -> []child_ids
}

func newMockOrganizationRepository() *mockOrganizationRepository {
	return &mockOrganizationRepository{
		organizations: make(map[uuid.UUID]*organizations.Organization),
		nameIndex:     make(map[string]uuid.UUID),
		parentIndex:   make(map[uuid.UUID][]uuid.UUID),
	}
}

func (m *mockOrganizationRepository) CreateOrganization(
	_ context.Context,
	createFn func() (*organizations.Organization, error),
) (*organizations.Organization, error) {
	organization, err := createFn()
	if err != nil {
		return nil, err
	}

	// Check for duplicate name
	if _, exists := m.nameIndex[organization.Name()]; exists {
		return nil, organizations.ErrOrganizationAlreadyExist
	}

	// Check parent exists if provided
	if organization.HasParent() {
		if _, exists := m.organizations[*organization.ParentID()]; !exists {
			return nil, organizations.ErrOrganizationNotFound
		}
	}

	// Store organization
	m.organizations[organization.ID()] = organization
	m.nameIndex[organization.Name()] = organization.ID()

	// Update parent index
	if organization.HasParent() {
		parentID := *organization.ParentID()
		m.parentIndex[parentID] = append(m.parentIndex[parentID], organization.ID())
	}

	return organization, nil
}

func (m *mockOrganizationRepository) UpdateOrganization(
	_ context.Context,
	id uuid.UUID,
	updateFn func(*organizations.Organization) (bool, error),
) (*organizations.Organization, error) {
	organization, exists := m.organizations[id]
	if !exists {
		return nil, organizations.ErrOrganizationNotFound
	}

	// Create a copy to avoid modifying original during validation
	orgClone, err := organizations.NewOrganization(
		organization.ID(),
		organization.Name(),
		organization.Domain(),
		organization.ParentID(),
	)
	if err != nil {
		return nil, err
	}

	// Restore settings and activation state
	orgClone.UpdateSettings(organization.Settings())
	if !organization.IsActive() {
		orgClone.Deactivate()
	}

	updated, err := updateFn(orgClone)
	if err != nil {
		return nil, err
	}
	if !updated {
		return orgClone, nil
	}

	// Check for circular reference if parent changed
	if orgClone.HasParent() {
		if m.wouldCreateCircleOrg(id, *orgClone.ParentID()) {
			return nil, organizations.ErrCircularReference
		}
	}

	// Update name index if name changed
	if orgClone.Name() != organization.Name() {
		delete(m.nameIndex, organization.Name())
		m.nameIndex[orgClone.Name()] = id
	}

	// Update parent index if parent changed
	if organization.HasParent() && (!orgClone.HasParent() || *organization.ParentID() != *orgClone.ParentID()) {
		// Remove from old parent
		oldParentID := *organization.ParentID()
		children := m.parentIndex[oldParentID]
		for i, childID := range children {
			if childID == id {
				m.parentIndex[oldParentID] = append(children[:i], children[i+1:]...)
				break
			}
		}
	}

	if orgClone.HasParent() && (!organization.HasParent() || *organization.ParentID() != *orgClone.ParentID()) {
		// Add to new parent
		newParentID := *orgClone.ParentID()
		m.parentIndex[newParentID] = append(m.parentIndex[newParentID], id)
	}

	// Update storage
	m.organizations[id] = orgClone

	return orgClone, nil
}

func (m *mockOrganizationRepository) GetOrganization(_ context.Context, id uuid.UUID) (*organizations.Organization, error) {
	organization, exists := m.organizations[id]
	if !exists {
		return nil, organizations.ErrOrganizationNotFound
	}
	return organization, nil
}

func (m *mockOrganizationRepository) ListOrganizations(
	_ context.Context,
	filter application.OrganizationFilter,
) ([]*organizations.Organization, error) {
	var result []*organizations.Organization

	for _, organization := range m.organizations {
		// Apply filters
		if filter.ParentID != nil {
			if organization.ParentID() == nil || *organization.ParentID() != *filter.ParentID {
				continue
			}
		}
		if filter.IsActive != nil && organization.IsActive() != *filter.IsActive {
			continue
		}
		if filter.IsRootOnly && organization.HasParent() {
			continue
		}
		if filter.Name != nil && organization.Name() != *filter.Name {
			continue
		}
		if filter.Domain != nil && organization.Domain() != *filter.Domain {
			continue
		}

		result = append(result, organization)
	}

	// Apply pagination
	if filter.Offset > 0 && filter.Offset < len(result) {
		result = result[filter.Offset:]
	} else if filter.Offset >= len(result) {
		result = []*organizations.Organization{}
	}

	if filter.Limit > 0 && filter.Limit < len(result) {
		result = result[:filter.Limit]
	}

	return result, nil
}

func (m *mockOrganizationRepository) GetOrganizationHierarchy(
	ctx context.Context,
	rootID uuid.UUID,
) (*application.OrganizationTree, error) {
	root, err := m.GetOrganization(ctx, rootID)
	if err != nil {
		return nil, err
	}

	tree := &application.OrganizationTree{
		Organization: root,
		Children:     []*application.OrganizationTree{},
	}

	// Get children and build their hierarchies
	childIDs := m.parentIndex[rootID]
	for _, childID := range childIDs {
		childTree, err := m.GetOrganizationHierarchy(ctx, childID)
		if err != nil {
			return nil, err
		}
		tree.Children = append(tree.Children, childTree)
	}

	return tree, nil
}

func (m *mockOrganizationRepository) DeleteOrganization(_ context.Context, id uuid.UUID) error {
	organization, exists := m.organizations[id]
	if !exists {
		return organizations.ErrOrganizationNotFound
	}

	// Check for children
	if children := m.parentIndex[id]; len(children) > 0 {
		return organizations.ErrInvalidOrganization // Cannot delete organization with children
	}

	// Remove from storage
	delete(m.organizations, id)

	// Remove from name index
	delete(m.nameIndex, organization.Name())

	// Remove from parent index
	if organization.HasParent() {
		parentID := *organization.ParentID()
		children := m.parentIndex[parentID]
		for i, childID := range children {
			if childID == id {
				m.parentIndex[parentID] = append(children[:i], children[i+1:]...)
				break
			}
		}
	}

	return nil
}

// wouldCreateCircleOrg checks if setting newParentID as parent would create circular reference
func (m *mockOrganizationRepository) wouldCreateCircleOrg(orgID, newParentID uuid.UUID) bool {
	if orgID == newParentID {
		return true
	}

	visited := make(map[uuid.UUID]bool)
	currentID := newParentID

	for !visited[currentID] {
		visited[currentID] = true

		if currentID == orgID {
			return true // Found circle
		}

		organization, exists := m.organizations[currentID]
		if !exists || !organization.HasParent() {
			break
		}
		currentID = *organization.ParentID()
	}

	return false
}

// mockTicketRepository is a simple mock for integration testing
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
	filter application.TicketFilter,
) ([]*tickets.Ticket, error) {
	var result []*tickets.Ticket

	for _, ticket := range m.tickets {
		// Apply filters
		if filter.Status != nil && ticket.Status() != *filter.Status {
			continue
		}
		if filter.Priority != nil && ticket.Priority() != *filter.Priority {
			continue
		}
		if filter.AssigneeID != nil {
			if ticket.AssigneeID() == nil || *ticket.AssigneeID() != *filter.AssigneeID {
				continue
			}
		}
		if filter.AuthorID != nil && ticket.AuthorID() != *filter.AuthorID {
			continue
		}
		if filter.OrganizationID != nil && ticket.OrganizationID() != *filter.OrganizationID {
			continue
		}
		if filter.CategoryID != nil {
			if ticket.CategoryID() == nil || *ticket.CategoryID() != *filter.CategoryID {
				continue
			}
		}

		result = append(result, ticket)
	}

	// Apply pagination
	if filter.Offset > 0 && filter.Offset < len(result) {
		result = result[filter.Offset:]
	} else if filter.Offset >= len(result) {
		result = []*tickets.Ticket{}
	}

	if filter.Limit > 0 && filter.Limit < len(result) {
		result = result[:filter.Limit]
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
