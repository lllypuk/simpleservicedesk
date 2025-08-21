package categories_test

import (
	"errors"
	"fmt"
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

// Nested Categories Tests - Testing hierarchical category structures

func TestCategory_NestedCategories_BasicHierarchy(t *testing.T) {
	orgID := uuid.New()

	// Level 1: Root category
	rootCat, err := domain.CreateRootCategory("IT Support", "Information Technology Support", orgID)
	require.NoError(t, err)
	require.True(t, rootCat.IsRootCategory())

	// Level 2: Main categories
	hardwareCat, err := domain.CreateSubCategory("Hardware", "Hardware related issues", orgID, rootCat.ID())
	require.NoError(t, err)
	require.False(t, hardwareCat.IsRootCategory())
	require.Equal(t, rootCat.ID(), *hardwareCat.ParentID())

	softwareCat, err := domain.CreateSubCategory("Software", "Software related issues", orgID, rootCat.ID())
	require.NoError(t, err)
	require.Equal(t, rootCat.ID(), *softwareCat.ParentID())

	// Level 3: Specific categories
	laptopsCat, err := domain.CreateSubCategory("Laptops", "Laptop hardware issues", orgID, hardwareCat.ID())
	require.NoError(t, err)
	require.Equal(t, hardwareCat.ID(), *laptopsCat.ParentID())

	driversCat, err := domain.CreateSubCategory("Drivers", "Driver related software issues", orgID, softwareCat.ID())
	require.NoError(t, err)
	require.Equal(t, softwareCat.ID(), *driversCat.ParentID())

	// Level 4: Very specific categories
	macbooksCat, err := domain.CreateSubCategory("MacBooks", "MacBook specific issues", orgID, laptopsCat.ID())
	require.NoError(t, err)
	require.Equal(t, laptopsCat.ID(), *macbooksCat.ParentID())

	// Verify organizational structure
	require.True(t, rootCat.BelongsToOrganization(orgID))
	require.True(t, hardwareCat.BelongsToOrganization(orgID))
	require.True(t, softwareCat.BelongsToOrganization(orgID))
	require.True(t, laptopsCat.BelongsToOrganization(orgID))
	require.True(t, driversCat.BelongsToOrganization(orgID))
	require.True(t, macbooksCat.BelongsToOrganization(orgID))
}

func TestCategory_NestedCategories_FullPathGeneration(t *testing.T) {
	orgID := uuid.New()

	// Create multi-level hierarchy
	rootCat, err := domain.CreateRootCategory("Support", "Customer Support", orgID)
	require.NoError(t, err)

	level2Cat, err := domain.CreateSubCategory("Technical", "Technical Issues", orgID, rootCat.ID())
	require.NoError(t, err)

	level3Cat, err := domain.CreateSubCategory("Hardware", "Hardware Problems", orgID, level2Cat.ID())
	require.NoError(t, err)

	level4Cat, err := domain.CreateSubCategory("Computers", "Computer Hardware", orgID, level3Cat.ID())
	require.NoError(t, err)

	level5Cat, err := domain.CreateSubCategory("Gaming PCs", "Gaming Computer Issues", orgID, level4Cat.ID())
	require.NoError(t, err)

	// Create lookup function that simulates repository
	categories := map[uuid.UUID]*domain.Category{
		rootCat.ID():   rootCat,
		level2Cat.ID(): level2Cat,
		level3Cat.ID(): level3Cat,
		level4Cat.ID(): level4Cat,
		level5Cat.ID(): level5Cat,
	}

	// Build full paths recursively
	var buildFullPath func(*domain.Category) (string, error)
	buildFullPath = func(cat *domain.Category) (string, error) {
		if cat.IsRootCategory() {
			return cat.Name(), nil
		}

		parent := categories[*cat.ParentID()]
		parentPath, parentErr := buildFullPath(parent)
		if parentErr != nil {
			return "", parentErr
		}
		return parentPath + " / " + cat.Name(), nil
	}

	// Test full path generation
	tests := []struct {
		category     *domain.Category
		expectedPath string
	}{
		{rootCat, "Support"},
		{level2Cat, "Support / Technical"},
		{level3Cat, "Support / Technical / Hardware"},
		{level4Cat, "Support / Technical / Hardware / Computers"},
		{level5Cat, "Support / Technical / Hardware / Computers / Gaming PCs"},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("path_for_%s", tt.category.Name()), func(t *testing.T) {
			path, pathErr := buildFullPath(tt.category)
			require.NoError(t, pathErr)
			require.Equal(t, tt.expectedPath, path)
		})
	}
}

func TestCategory_NestedCategories_ParentChanges(t *testing.T) {
	orgID := uuid.New()

	// Create initial hierarchy
	rootCat1, err := domain.CreateRootCategory("IT", "IT Department", orgID)
	require.NoError(t, err)

	rootCat2, err := domain.CreateRootCategory("HR", "Human Resources", orgID)
	require.NoError(t, err)

	childCat, err := domain.CreateSubCategory("Hardware", "Hardware Issues", orgID, rootCat1.ID())
	require.NoError(t, err)

	subChildCat, err := domain.CreateSubCategory("Laptops", "Laptop Issues", orgID, childCat.ID())
	require.NoError(t, err)

	// Verify initial structure
	require.Equal(t, rootCat1.ID(), *childCat.ParentID())
	require.Equal(t, childCat.ID(), *subChildCat.ParentID())

	// Test moving child category to different parent
	originalUpdatedAt := childCat.UpdatedAt()
	time.Sleep(time.Millisecond)

	rootCat2ID := rootCat2.ID()
	err = childCat.ChangeParent(&rootCat2ID)
	require.NoError(t, err)
	require.Equal(t, rootCat2.ID(), *childCat.ParentID())
	require.True(t, childCat.UpdatedAt().After(originalUpdatedAt))

	// Sub-child should still point to the moved child
	require.Equal(t, childCat.ID(), *subChildCat.ParentID())

	// Test moving child to root
	err = childCat.ChangeParent(nil)
	require.NoError(t, err)
	require.True(t, childCat.IsRootCategory())
	require.Nil(t, childCat.ParentID())

	// Test moving child to be under its own sub-child (should be prevented at domain level)
	subChildCatID := subChildCat.ID()
	err = childCat.ChangeParent(&subChildCatID)
	require.NoError(t, err) // Domain doesn't prevent this, service layer should

	// But self-reference should be prevented
	childCatID := childCat.ID()
	err = childCat.ChangeParent(&childCatID)
	require.Error(t, err)
	require.ErrorIs(t, err, domain.ErrCircularReference)
}

func TestCategory_NestedCategories_OrganizationalIsolation(t *testing.T) {
	org1ID := uuid.New()
	org2ID := uuid.New()

	// Create identical hierarchy in two organizations
	// Organization 1
	org1Root, err := domain.CreateRootCategory("Support", "Customer Support", org1ID)
	require.NoError(t, err)

	org1Tech, err := domain.CreateSubCategory("Technical", "Technical Issues", org1ID, org1Root.ID())
	require.NoError(t, err)

	org1Hardware, err := domain.CreateSubCategory("Hardware", "Hardware Issues", org1ID, org1Tech.ID())
	require.NoError(t, err)

	// Organization 2
	org2Root, err := domain.CreateRootCategory("Support", "Customer Support", org2ID)
	require.NoError(t, err)

	org2Tech, err := domain.CreateSubCategory("Technical", "Technical Issues", org2ID, org2Root.ID())
	require.NoError(t, err)

	org2Hardware, err := domain.CreateSubCategory("Hardware", "Hardware Issues", org2ID, org2Tech.ID())
	require.NoError(t, err)

	// Verify organizational isolation
	require.True(t, org1Root.BelongsToOrganization(org1ID))
	require.False(t, org1Root.BelongsToOrganization(org2ID))
	require.True(t, org1Tech.BelongsToOrganization(org1ID))
	require.False(t, org1Tech.BelongsToOrganization(org2ID))
	require.True(t, org1Hardware.BelongsToOrganization(org1ID))
	require.False(t, org1Hardware.BelongsToOrganization(org2ID))

	require.True(t, org2Root.BelongsToOrganization(org2ID))
	require.False(t, org2Root.BelongsToOrganization(org1ID))
	require.True(t, org2Tech.BelongsToOrganization(org2ID))
	require.False(t, org2Tech.BelongsToOrganization(org1ID))
	require.True(t, org2Hardware.BelongsToOrganization(org2ID))
	require.False(t, org2Hardware.BelongsToOrganization(org1ID))

	// Verify IDs are different even with same names
	require.NotEqual(t, org1Root.ID(), org2Root.ID())
	require.NotEqual(t, org1Tech.ID(), org2Tech.ID())
	require.NotEqual(t, org1Hardware.ID(), org2Hardware.ID())

	// Test that parent relationships are correct within organizations
	require.Equal(t, org1Root.ID(), *org1Tech.ParentID())
	require.Equal(t, org1Tech.ID(), *org1Hardware.ParentID())
	require.Equal(t, org2Root.ID(), *org2Tech.ParentID())
	require.Equal(t, org2Tech.ID(), *org2Hardware.ParentID())

	// Cross-organization parent references would be invalid (but domain layer doesn't prevent this)
	// This should be validated at the service layer
}

func TestCategory_NestedCategories_DeepNesting(t *testing.T) {
	orgID := uuid.New()

	// Create very deep nesting (10 levels)
	categories := make([]*domain.Category, 10)
	names := []string{
		"Level1_Root", "Level2_Region", "Level3_Division", "Level4_Department", "Level5_Team",
		"Level6_SubTeam", "Level7_Project", "Level8_Module", "Level9_Component", "Level10_Detail",
	}

	// Create root category
	var err error
	categories[0], err = domain.CreateRootCategory(names[0], fmt.Sprintf("Description for %s", names[0]), orgID)
	require.NoError(t, err)

	// Create nested categories
	for i := 1; i < 10; i++ {
		categories[i], err = domain.CreateSubCategory(
			names[i],
			fmt.Sprintf("Description for %s", names[i]),
			orgID,
			categories[i-1].ID(),
		)
		require.NoError(t, err)
	}

	// Verify the deep nesting structure
	require.True(t, categories[0].IsRootCategory())
	for i := 1; i < 10; i++ {
		require.False(t, categories[i].IsRootCategory())
		require.Equal(t, categories[i-1].ID(), *categories[i].ParentID())
		require.True(t, categories[i].BelongsToOrganization(orgID))
	}

	// Test moving deepest category to root
	deepest := categories[9]
	originalParent := *deepest.ParentID()

	deepest.MoveToRoot()
	require.True(t, deepest.IsRootCategory())
	require.Nil(t, deepest.ParentID())

	// Move back to original position
	err = deepest.ChangeParent(&originalParent)
	require.NoError(t, err)
	require.Equal(t, originalParent, *deepest.ParentID())
}

func TestCategory_NestedCategories_BulkOperations(t *testing.T) {
	orgID := uuid.New()

	// Create root category
	rootCat, err := domain.CreateRootCategory("Bulk Test Root", "Root for bulk operations", orgID)
	require.NoError(t, err)

	// Create multiple categories under root
	const numCategories = 50
	categories := make([]*domain.Category, numCategories)

	for i := range numCategories {
		categories[i], err = domain.CreateSubCategory(
			fmt.Sprintf("Category_%d", i),
			fmt.Sprintf("Description for category %d", i),
			orgID,
			rootCat.ID(),
		)
		require.NoError(t, err)
		require.Equal(t, rootCat.ID(), *categories[i].ParentID())
	}

	// Verify all categories belong to the same organization and parent
	for i := range numCategories {
		require.True(t, categories[i].BelongsToOrganization(orgID))
		require.Equal(t, rootCat.ID(), *categories[i].ParentID())
		require.False(t, categories[i].IsRootCategory())
	}

	// Test bulk activation/deactivation
	for i := range numCategories {
		if i%2 == 0 {
			categories[i].Deactivate()
		}
	}

	// Verify activation states
	for i := range numCategories {
		if i%2 == 0 {
			require.False(t, categories[i].IsActive(), "Category %d should be inactive", i)
		} else {
			require.True(t, categories[i].IsActive(), "Category %d should be active", i)
		}
	}

	// Reactivate all
	for i := range numCategories {
		categories[i].Activate()
		require.True(t, categories[i].IsActive())
	}
}

// Edge Cases and Validation Tests

func TestCategory_EdgeCases_NameValidation(t *testing.T) {
	orgID := uuid.New()

	tests := []struct {
		name       string
		catName    string
		shouldPass bool
	}{
		{"empty name", "", false},
		{"whitespace only", "   ", false},
		{"single char", "A", false},
		{"minimum valid", "AB", true},
		{"with leading spaces", "  Valid Name  ", true}, // Should be trimmed
		{"unicode characters", "Категория", true},
		{"special characters", "IT & Hardware", true},
		{"numbers and symbols", "Level-1_Category_2023", true},
		{"very long name", strings.Repeat("A", domain.MaxNameLength), true},
		{"too long name", strings.Repeat("A", domain.MaxNameLength+1), false},
		{"newlines and tabs", "Name\nWith\tSpecial\rChars", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := domain.CreateRootCategory(tt.catName, "Test description", orgID)
			if tt.shouldPass {
				require.NoError(t, err, "Category creation should succeed for: %s", tt.catName)
			} else {
				require.Error(t, err, "Category creation should fail for: %s", tt.catName)
				require.ErrorIs(t, err, domain.ErrCategoryValidation)
			}
		})
	}
}

func TestCategory_EdgeCases_DescriptionValidation(t *testing.T) {
	orgID := uuid.New()

	tests := []struct {
		name        string
		description string
		shouldPass  bool
	}{
		{"empty description", "", true},
		{"whitespace only", "   ", true}, // Should be trimmed to empty
		{"normal description", "This is a valid description", true},
		{"unicode description", "Описание на русском языке", true},
		{"description with newlines", "Line 1\nLine 2\nLine 3", true},
		{"maximum length", strings.Repeat("A", domain.MaxDescriptionLength), true},
		{"too long description", strings.Repeat("A", domain.MaxDescriptionLength+1), false},
		{"description with special chars", "Description with @#$%^&*()_+ symbols", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := domain.CreateRootCategory("Valid Name", tt.description, orgID)
			if tt.shouldPass {
				require.NoError(t, err, "Category creation should succeed with description: %s", tt.description)
			} else {
				require.Error(t, err, "Category creation should fail with description: %s", tt.description)
				require.ErrorIs(t, err, domain.ErrCategoryValidation)
			}
		})
	}
}

func TestCategory_EdgeCases_CircularReferenceDetection(t *testing.T) {
	orgID := uuid.New()

	// Create categories
	cat1, err := domain.CreateRootCategory("Category 1", "First category", orgID)
	require.NoError(t, err)

	cat2, err := domain.CreateSubCategory("Category 2", "Second category", orgID, cat1.ID())
	require.NoError(t, err)

	cat3, err := domain.CreateSubCategory("Category 3", "Third category", orgID, cat2.ID())
	require.NoError(t, err)

	// Test direct self-reference (should be prevented)
	cat1ID := cat1.ID()
	err = cat1.ChangeParent(&cat1ID)
	require.Error(t, err)
	require.ErrorIs(t, err, domain.ErrCircularReference)

	cat2ID := cat2.ID()
	err = cat2.ChangeParent(&cat2ID)
	require.Error(t, err)
	require.ErrorIs(t, err, domain.ErrCircularReference)

	// Test indirect circular reference (domain level only prevents direct self-reference)
	// These should be prevented at service level, but domain allows them
	cat3ID := cat3.ID()
	err = cat1.ChangeParent(&cat3ID) // This would create 3 -> 2 -> 1 -> 3 cycle
	require.NoError(t, err)          // Domain layer doesn't prevent this

	// Verify the change was applied (even though it creates a logical issue)
	require.Equal(t, cat3.ID(), *cat1.ParentID())
}

func TestCategory_EdgeCases_MultipleRootCategories(t *testing.T) {
	orgID := uuid.New()

	// Create multiple root categories for the same organization
	roots := make([]*domain.Category, 5)
	for i := range 5 {
		var err error
		roots[i], err = domain.CreateRootCategory(
			fmt.Sprintf("Root Category %d", i),
			fmt.Sprintf("Root description %d", i),
			orgID,
		)
		require.NoError(t, err)
		require.True(t, roots[i].IsRootCategory())
		require.Nil(t, roots[i].ParentID())
		require.True(t, roots[i].BelongsToOrganization(orgID))
	}

	// Verify all have unique IDs
	for i := range 5 {
		for j := i + 1; j < 5; j++ {
			require.NotEqual(t, roots[i].ID(), roots[j].ID())
		}
	}

	// Create sub-categories under different roots
	for i := range 5 {
		subCat, err := domain.CreateSubCategory(
			fmt.Sprintf("Sub of Root %d", i),
			"Sub category",
			orgID,
			roots[i].ID(),
		)
		require.NoError(t, err)
		require.Equal(t, roots[i].ID(), *subCat.ParentID())
	}
}

func TestCategory_EdgeCases_StateConsistency(t *testing.T) {
	orgID := uuid.New()

	// Create category and verify initial state
	cat, err := domain.CreateRootCategory("Consistency Test", "Testing state consistency", orgID)
	require.NoError(t, err)

	// Store initial values
	initialID := cat.ID()
	initialOrgID := cat.OrganizationID()
	initialCreatedAt := cat.CreatedAt()
	initialIsActive := cat.IsActive()
	initialIsRoot := cat.IsRootCategory()

	// Perform various operations that should not affect immutable fields
	err = cat.ChangeName("New Name")
	require.NoError(t, err)

	err = cat.ChangeDescription("New description")
	require.NoError(t, err)

	cat.Deactivate()
	cat.Activate()

	parentID := uuid.New()
	err = cat.ChangeParent(&parentID)
	require.NoError(t, err)

	cat.MoveToRoot()

	// Verify immutable fields haven't changed
	require.Equal(t, initialID, cat.ID(), "ID should never change")
	require.Equal(t, initialOrgID, cat.OrganizationID(), "OrganizationID should never change")
	require.Equal(t, initialCreatedAt, cat.CreatedAt(), "CreatedAt should never change")

	// Verify mutable fields reflect current state
	require.Equal(t, "New Name", cat.Name())
	require.Equal(t, "New description", cat.Description())
	require.Equal(t, initialIsActive, cat.IsActive())        // Should be back to initial state
	require.Equal(t, initialIsRoot, cat.IsRootCategory())    // Should be back to root
	require.True(t, cat.UpdatedAt().After(initialCreatedAt)) // Should be updated
}
