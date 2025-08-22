//go:build integration
// +build integration

package shared

import (
	"context"

	"simpleservicedesk/internal/application"
	"simpleservicedesk/internal/domain/categories"
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
