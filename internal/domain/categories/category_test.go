package categories_test

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	domain "simpleservicedesk/internal/domain/categories"
)

func TestNewCategory_Valid(t *testing.T) {
	id := uuid.New()
	orgID := uuid.New()
	parentID := uuid.New()
	name := "IT Support"
	description := "Technical support category"

	cat, err := domain.NewCategory(id, name, description, orgID, &parentID)

	require.NoError(t, err)
	require.Equal(t, id, cat.ID())
	require.Equal(t, name, cat.Name())
	require.Equal(t, description, cat.Description())
	require.Equal(t, orgID, cat.OrganizationID())
	require.Equal(t, &parentID, cat.ParentID())
	require.True(t, cat.IsActive())
	require.False(t, cat.CreatedAt().IsZero())
	require.False(t, cat.UpdatedAt().IsZero())
	require.False(t, cat.IsRootCategory())
	require.True(t, cat.HasParent())
}

func TestNewCategory_RootCategory(t *testing.T) {
	id := uuid.New()
	orgID := uuid.New()
	name := "General"
	description := "General category"

	cat, err := domain.NewCategory(id, name, description, orgID, nil)

	require.NoError(t, err)
	require.Equal(t, id, cat.ID())
	require.Nil(t, cat.ParentID())
	require.True(t, cat.IsRootCategory())
	require.False(t, cat.HasParent())
}

func TestNewCategory_InvalidName(t *testing.T) {
	tests := []struct {
		name     string
		catName  string
		hasError bool
	}{
		{"empty name", "", true},
		{"too short", "A", true},
		{"valid short", "IT", false},
		{"valid long", "Information Technology Support", false},
		{"whitespace only", "   ", true},
	}

	orgID := uuid.New()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := domain.NewCategory(uuid.New(), tt.catName, "Description", orgID, nil)
			if tt.hasError {
				require.Error(t, err)
				require.ErrorIs(t, err, domain.ErrCategoryValidation)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestNewCategory_InvalidDescription(t *testing.T) {
	longDescription := strings.Repeat("a", domain.MaxDescriptionLength+1)

	tests := []struct {
		name        string
		description string
		hasError    bool
	}{
		{"empty description", "", false}, // описание может быть пустым
		{"valid description", "Valid description", false},
		{"too long description", longDescription, true},
	}

	orgID := uuid.New()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := domain.NewCategory(uuid.New(), "Test Category", tt.description, orgID, nil)
			if tt.hasError {
				require.Error(t, err)
				require.ErrorIs(t, err, domain.ErrCategoryValidation)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestNewCategory_InvalidOrganizationID(t *testing.T) {
	_, err := domain.NewCategory(uuid.New(), "Test Category", "Description", uuid.Nil, nil)

	require.Error(t, err)
	require.ErrorIs(t, err, domain.ErrCategoryValidation)
}

func TestCreateCategory(t *testing.T) {
	orgID := uuid.New()
	parentID := uuid.New()
	name := "Hardware"
	description := "Hardware issues"

	cat, err := domain.CreateCategory(name, description, orgID, &parentID)

	require.NoError(t, err)
	require.NotEqual(t, uuid.Nil, cat.ID())
	require.Equal(t, name, cat.Name())
	require.Equal(t, description, cat.Description())
	require.Equal(t, orgID, cat.OrganizationID())
	require.Equal(t, &parentID, cat.ParentID())
}

func TestCreateRootCategory(t *testing.T) {
	orgID := uuid.New()
	name := "IT"
	description := "Information Technology"

	cat, err := domain.CreateRootCategory(name, description, orgID)

	require.NoError(t, err)
	require.True(t, cat.IsRootCategory())
	require.Nil(t, cat.ParentID())
}

func TestCreateSubCategory(t *testing.T) {
	orgID := uuid.New()
	parentID := uuid.New()
	name := "Laptops"
	description := "Laptop issues"

	cat, err := domain.CreateSubCategory(name, description, orgID, parentID)

	require.NoError(t, err)
	require.False(t, cat.IsRootCategory())
	require.Equal(t, &parentID, cat.ParentID())
}

func TestCategory_ChangeName(t *testing.T) {
	cat, err := domain.CreateRootCategory("Original Name", "Description", uuid.New())
	require.NoError(t, err)

	originalUpdatedAt := cat.UpdatedAt()
	time.Sleep(time.Millisecond)

	newName := "New Category Name"
	err = cat.ChangeName(newName)

	require.NoError(t, err)
	require.Equal(t, newName, cat.Name())
	require.True(t, cat.UpdatedAt().After(originalUpdatedAt))
}

func TestCategory_ChangeName_Invalid(t *testing.T) {
	cat, err := domain.CreateRootCategory("Original Name", "Description", uuid.New())
	require.NoError(t, err)

	originalName := cat.Name()
	originalUpdatedAt := cat.UpdatedAt()

	err = cat.ChangeName("")

	require.Error(t, err)
	require.ErrorIs(t, err, domain.ErrCategoryValidation)
	require.Equal(t, originalName, cat.Name())
	require.Equal(t, originalUpdatedAt, cat.UpdatedAt())
}

func TestCategory_ChangeDescription(t *testing.T) {
	cat, err := domain.CreateRootCategory("Category", "Old description", uuid.New())
	require.NoError(t, err)

	originalUpdatedAt := cat.UpdatedAt()
	time.Sleep(time.Millisecond)

	newDescription := "New description"
	err = cat.ChangeDescription(newDescription)

	require.NoError(t, err)
	require.Equal(t, newDescription, cat.Description())
	require.True(t, cat.UpdatedAt().After(originalUpdatedAt))
}

func TestCategory_ChangeParent(t *testing.T) {
	cat, err := domain.CreateRootCategory("Category", "Description", uuid.New())
	require.NoError(t, err)
	require.True(t, cat.IsRootCategory())

	originalUpdatedAt := cat.UpdatedAt()
	time.Sleep(time.Millisecond)

	newParentID := uuid.New()
	err = cat.ChangeParent(&newParentID)

	require.NoError(t, err)
	require.Equal(t, &newParentID, cat.ParentID())
	require.False(t, cat.IsRootCategory())
	require.True(t, cat.UpdatedAt().After(originalUpdatedAt))
}

func TestCategory_ChangeParent_SelfReference(t *testing.T) {
	cat, err := domain.CreateRootCategory("Category", "Description", uuid.New())
	require.NoError(t, err)

	catID := cat.ID()
	err = cat.ChangeParent(&catID)

	require.Error(t, err)
	require.ErrorIs(t, err, domain.ErrCircularReference)
}

func TestCategory_MoveToRoot(t *testing.T) {
	orgID := uuid.New()
	parentID := uuid.New()
	cat, err := domain.CreateSubCategory("Category", "Description", orgID, parentID)
	require.NoError(t, err)
	require.False(t, cat.IsRootCategory())

	originalUpdatedAt := cat.UpdatedAt()
	time.Sleep(time.Millisecond)

	cat.MoveToRoot()

	require.True(t, cat.IsRootCategory())
	require.Nil(t, cat.ParentID())
	require.True(t, cat.UpdatedAt().After(originalUpdatedAt))
}

func TestCategory_ActivateDeactivate(t *testing.T) {
	cat, err := domain.CreateRootCategory("Category", "Description", uuid.New())
	require.NoError(t, err)
	require.True(t, cat.IsActive())

	originalUpdatedAt := cat.UpdatedAt()
	time.Sleep(time.Millisecond)

	cat.Deactivate()
	require.False(t, cat.IsActive())
	require.True(t, cat.UpdatedAt().After(originalUpdatedAt))

	deactivatedAt := cat.UpdatedAt()
	time.Sleep(time.Millisecond)

	cat.Activate()
	require.True(t, cat.IsActive())
	require.True(t, cat.UpdatedAt().After(deactivatedAt))
}

func TestCategory_BelongsToOrganization(t *testing.T) {
	orgID := uuid.New()
	otherOrgID := uuid.New()
	cat, err := domain.CreateRootCategory("Category", "Description", orgID)
	require.NoError(t, err)

	require.True(t, cat.BelongsToOrganization(orgID))
	require.False(t, cat.BelongsToOrganization(otherOrgID))
}

func TestCategory_FullPath(t *testing.T) {
	cat, err := domain.CreateRootCategory("IT", "IT Department", uuid.New())
	require.NoError(t, err)

	// Тест для корневой категории
	path, err := cat.FullPath(nil)
	require.NoError(t, err)
	require.Equal(t, "IT", path)

	// Тест для категории с родителем
	parentID := uuid.New()
	subCat, err := domain.CreateSubCategory("Hardware", "Hardware issues", uuid.New(), parentID)
	require.NoError(t, err)

	getParentName := func(id uuid.UUID) (string, error) {
		if id == parentID {
			return "IT", nil
		}
		return "", errors.New("parent not found")
	}

	path, err = subCat.FullPath(getParentName)
	require.NoError(t, err)
	require.Equal(t, "IT / Hardware", path)
}
