package categories_test

import (
	"context"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"simpleservicedesk/internal/application"
	domain "simpleservicedesk/internal/domain/categories"
	categoriesInfra "simpleservicedesk/internal/infrastructure/categories"
	"simpleservicedesk/internal/queries"
)

type MongoRepoSuite struct {
	suite.Suite

	container testcontainers.Container
	db        *mongo.Database
	repo      *categoriesInfra.MongoRepo
	orgID     uuid.UUID
}

func (s *MongoRepoSuite) SetupSuite() {
	ctx := context.Background()
	req := testcontainers.ContainerRequest{
		Image:        "mongo:latest",
		ExposedPorts: []string{"27017/tcp"},
		WaitingFor:   wait.ForLog("Waiting for connections").WithStartupTimeout(10 * time.Second),
	}
	mongoContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	s.Require().NoError(err)
	s.container = mongoContainer

	host, err := mongoContainer.Host(ctx)
	s.Require().NoError(err)
	port, err := mongoContainer.MappedPort(ctx, "27017")
	s.Require().NoError(err)

	uri := fmt.Sprintf("mongodb://%s", net.JoinHostPort(host, port.Port()))
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	s.Require().NoError(err)

	s.db = client.Database("testdb")
	s.repo = categoriesInfra.NewMongoRepo(s.db)
	s.orgID = uuid.New()
}

func (s *MongoRepoSuite) TearDownSuite() {
	ctx := context.Background()
	err := s.db.Client().Disconnect(ctx)
	s.Require().NoError(err)
	err = s.container.Terminate(ctx)
	s.Require().NoError(err)
}

func (s *MongoRepoSuite) SetupTest() {
	err := s.db.Collection("categories").Drop(context.Background())
	s.Require().NoError(err)
}

func (s *MongoRepoSuite) TestCreateAndGetCategory() {
	ctx := context.Background()
	name := "IT Support"
	description := "Information Technology Support Category"

	createdCategory, err := s.repo.CreateCategory(ctx, func() (*domain.Category, error) {
		return domain.CreateRootCategory(name, description, s.orgID)
	})
	s.Require().NoError(err)
	s.Require().NotNil(createdCategory)
	s.Equal(name, createdCategory.Name())
	s.Equal(description, createdCategory.Description())
	s.Equal(s.orgID, createdCategory.OrganizationID())
	s.True(createdCategory.IsRootCategory())

	fetchedCategory, err := s.repo.GetCategory(ctx, createdCategory.ID())
	s.Require().NoError(err)
	s.Require().NotNil(fetchedCategory)
	s.Equal(createdCategory.ID(), fetchedCategory.ID())
	s.Equal(createdCategory.Name(), fetchedCategory.Name())
	s.Equal(createdCategory.Description(), fetchedCategory.Description())
	s.Equal(createdCategory.OrganizationID(), fetchedCategory.OrganizationID())
	s.True(fetchedCategory.IsRootCategory())
}

func (s *MongoRepoSuite) TestCreateSubCategory() {
	ctx := context.Background()

	// Create parent category
	parentCategory, err := s.repo.CreateCategory(ctx, func() (*domain.Category, error) {
		return domain.CreateRootCategory("IT Support", "IT Support Category", s.orgID)
	})
	s.Require().NoError(err)

	// Create sub-category
	subCategory, err := s.repo.CreateCategory(ctx, func() (*domain.Category, error) {
		return domain.CreateSubCategory("Hardware", "Hardware Issues", s.orgID, parentCategory.ID())
	})
	s.Require().NoError(err)
	s.Require().NotNil(subCategory)
	s.Equal("Hardware", subCategory.Name())
	s.False(subCategory.IsRootCategory())
	s.True(subCategory.HasParent())
	s.Equal(parentCategory.ID(), *subCategory.ParentID())
}

func (s *MongoRepoSuite) TestCreateCategoryWithNonExistentParent() {
	ctx := context.Background()
	nonExistentParentID := uuid.New()

	_, err := s.repo.CreateCategory(ctx, func() (*domain.Category, error) {
		return domain.CreateSubCategory("Hardware", "Hardware Issues", s.orgID, nonExistentParentID)
	})
	s.Require().Error(err)
	s.Contains(err.Error(), "parent category not found")
}

func (s *MongoRepoSuite) TestCreateDuplicateCategory() {
	ctx := context.Background()
	name := "Duplicate Category"
	description := "Test duplicate category"

	// Create first category
	_, err := s.repo.CreateCategory(ctx, func() (*domain.Category, error) {
		return domain.CreateRootCategory(name, description, s.orgID)
	})
	s.Require().NoError(err)

	// Try to create duplicate category in same organization
	_, err = s.repo.CreateCategory(ctx, func() (*domain.Category, error) {
		return domain.CreateRootCategory(name, description, s.orgID)
	})
	s.Require().Error(err)
	s.ErrorIs(err, domain.ErrCategoryAlreadyExist)
}

func (s *MongoRepoSuite) TestUpdateCategory() {
	ctx := context.Background()

	category, err := s.repo.CreateCategory(ctx, func() (*domain.Category, error) {
		return domain.CreateRootCategory("Original Name", "Original Description", s.orgID)
	})
	s.Require().NoError(err)

	newName := "Updated Name"
	newDescription := "Updated Description"

	updatedCategory, err := s.repo.UpdateCategory(ctx, category.ID(), func(c *domain.Category) (bool, error) {
		changeNameErr := c.ChangeName(newName)
		if changeNameErr != nil {
			return false, changeNameErr
		}
		changeDescErr := c.ChangeDescription(newDescription)
		if changeDescErr != nil {
			return false, changeDescErr
		}
		return true, nil
	})

	s.Require().NoError(err)
	s.Equal(newName, updatedCategory.Name())
	s.Equal(newDescription, updatedCategory.Description())

	// Verify persistence
	fetchedCategory, err := s.repo.GetCategory(ctx, category.ID())
	s.Require().NoError(err)
	s.Equal(newName, fetchedCategory.Name())
	s.Equal(newDescription, fetchedCategory.Description())
}

func (s *MongoRepoSuite) TestUpdateCategoryParent() {
	ctx := context.Background()

	// Create categories
	rootCategory, err := s.repo.CreateCategory(ctx, func() (*domain.Category, error) {
		return domain.CreateRootCategory("Root", "Root Category", s.orgID)
	})
	s.Require().NoError(err)

	newParent, err := s.repo.CreateCategory(ctx, func() (*domain.Category, error) {
		return domain.CreateRootCategory("New Parent", "New Parent Category", s.orgID)
	})
	s.Require().NoError(err)

	// Update parent
	newParentID := newParent.ID()
	updatedCategory, err := s.repo.UpdateCategory(ctx, rootCategory.ID(), func(c *domain.Category) (bool, error) {
		return true, c.ChangeParent(&newParentID)
	})

	s.Require().NoError(err)
	s.True(updatedCategory.HasParent())
	s.Equal(newParent.ID(), *updatedCategory.ParentID())
}

func (s *MongoRepoSuite) TestUpdateCategoryCircularReference() {
	ctx := context.Background()

	// Create hierarchy: parent -> child
	parent, err := s.repo.CreateCategory(ctx, func() (*domain.Category, error) {
		return domain.CreateRootCategory("Parent", "Parent Category", s.orgID)
	})
	s.Require().NoError(err)

	child, err := s.repo.CreateCategory(ctx, func() (*domain.Category, error) {
		return domain.CreateSubCategory("Child", "Child Category", s.orgID, parent.ID())
	})
	s.Require().NoError(err)

	// Try to make parent a child of child (circular reference)
	childID := child.ID()
	_, err = s.repo.UpdateCategory(ctx, parent.ID(), func(c *domain.Category) (bool, error) {
		return true, c.ChangeParent(&childID)
	})

	s.Require().Error(err)
	s.ErrorIs(err, domain.ErrCircularReference)
}

func (s *MongoRepoSuite) TestListCategories() {
	ctx := context.Background()

	// Create test categories
	categories := []struct {
		name        string
		description string
		isActive    bool
	}{
		{"IT Support", "IT Support Category", true},
		{"HR", "Human Resources Category", true},
		{"Finance", "Finance Category", false},
	}

	createdCategories := make([]*domain.Category, len(categories))
	for i, cat := range categories {
		category, err := s.repo.CreateCategory(ctx, func() (*domain.Category, error) {
			return domain.CreateRootCategory(cat.name, cat.description, s.orgID)
		})
		s.Require().NoError(err)

		if !cat.isActive {
			_, err = s.repo.UpdateCategory(ctx, category.ID(), func(c *domain.Category) (bool, error) {
				c.Deactivate()
				return true, nil
			})
			s.Require().NoError(err)
		}
		createdCategories[i] = category
	}

	// Test listing all categories
	filter := queries.CategoryFilter{
		OrganizationID: &s.orgID,
	}
	result, err := s.repo.ListCategories(ctx, filter)
	s.Require().NoError(err)
	s.Len(result, 3)

	// Test filtering by active status
	active := true
	filter.IsActive = &active
	result, err = s.repo.ListCategories(ctx, filter)
	s.Require().NoError(err)
	s.Len(result, 2)

	// Test filtering by name
	name := "IT"
	filter.Name = &name
	filter.IsActive = nil
	result, err = s.repo.ListCategories(ctx, filter)
	s.Require().NoError(err)
	s.Len(result, 1)
	s.Equal("IT Support", result[0].Name())
}

func (s *MongoRepoSuite) TestListCategoriesWithPagination() {
	ctx := context.Background()

	// Create multiple categories
	for i := range 10 {
		_, err := s.repo.CreateCategory(ctx, func() (*domain.Category, error) {
			return domain.CreateRootCategory(fmt.Sprintf("Category %02d", i), "Description", s.orgID)
		})
		s.Require().NoError(err)
	}

	// Test pagination
	filter := queries.CategoryFilter{
		OrganizationID: &s.orgID,
	}
	filter.Limit = 5
	filter.Offset = 0
	filter.SortBy = "name"
	filter.SortOrder = "asc"

	firstPage, err := s.repo.ListCategories(ctx, filter)
	s.Require().NoError(err)
	s.Len(firstPage, 5)

	filter.Offset = 5
	secondPage, err := s.repo.ListCategories(ctx, filter)
	s.Require().NoError(err)
	s.Len(secondPage, 5)

	// Verify no overlap
	for _, cat1 := range firstPage {
		for _, cat2 := range secondPage {
			s.NotEqual(cat1.ID(), cat2.ID())
		}
	}
}

func (s *MongoRepoSuite) TestGetCategoryHierarchy() {
	ctx := context.Background()

	// Create hierarchy: IT -> Hardware -> Laptops
	root, err := s.repo.CreateCategory(ctx, func() (*domain.Category, error) {
		return domain.CreateRootCategory("IT", "IT Category", s.orgID)
	})
	s.Require().NoError(err)

	hardware, err := s.repo.CreateCategory(ctx, func() (*domain.Category, error) {
		return domain.CreateSubCategory("Hardware", "Hardware Issues", s.orgID, root.ID())
	})
	s.Require().NoError(err)

	laptops, err := s.repo.CreateCategory(ctx, func() (*domain.Category, error) {
		return domain.CreateSubCategory("Laptops", "Laptop Issues", s.orgID, hardware.ID())
	})
	s.Require().NoError(err)

	// Create another branch: IT -> Software
	_, err = s.repo.CreateCategory(ctx, func() (*domain.Category, error) {
		return domain.CreateSubCategory("Software", "Software Issues", s.orgID, root.ID())
	})
	s.Require().NoError(err)

	// Get hierarchy
	hierarchy, err := s.repo.GetCategoryHierarchy(ctx, root.ID())
	s.Require().NoError(err)
	s.Require().NotNil(hierarchy)

	// Verify root
	s.Equal(root.ID(), hierarchy.Category.ID())
	s.Len(hierarchy.Children, 2) // Hardware and Software

	// Find hardware branch
	var hardwareBranch *application.CategoryTree
	for _, child := range hierarchy.Children {
		if child.Category.Name() == "Hardware" {
			hardwareBranch = child
			break
		}
	}
	s.Require().NotNil(hardwareBranch)
	s.Equal(hardware.ID(), hardwareBranch.Category.ID())
	s.Len(hardwareBranch.Children, 1) // Laptops

	// Verify laptops
	s.Equal(laptops.ID(), hardwareBranch.Children[0].Category.ID())
	s.Empty(hardwareBranch.Children[0].Children) // No further children
}

func (s *MongoRepoSuite) TestDeleteCategory() {
	ctx := context.Background()

	category, err := s.repo.CreateCategory(ctx, func() (*domain.Category, error) {
		return domain.CreateRootCategory("Delete Test", "Category to delete", s.orgID)
	})
	s.Require().NoError(err)

	// Delete category
	err = s.repo.DeleteCategory(ctx, category.ID())
	s.Require().NoError(err)

	// Verify deletion
	_, err = s.repo.GetCategory(ctx, category.ID())
	s.Require().Error(err)
	s.ErrorIs(err, domain.ErrCategoryNotFound)
}

func (s *MongoRepoSuite) TestDeleteCategoryWithChildren() {
	ctx := context.Background()

	parent, err := s.repo.CreateCategory(ctx, func() (*domain.Category, error) {
		return domain.CreateRootCategory("Parent", "Parent Category", s.orgID)
	})
	s.Require().NoError(err)

	_, err = s.repo.CreateCategory(ctx, func() (*domain.Category, error) {
		return domain.CreateSubCategory("Child", "Child Category", s.orgID, parent.ID())
	})
	s.Require().NoError(err)

	// Try to delete parent with children
	err = s.repo.DeleteCategory(ctx, parent.ID())
	s.Require().Error(err)
	s.Contains(err.Error(), "cannot delete category with children")
}

func (s *MongoRepoSuite) TestGetCategoryNotFound() {
	ctx := context.Background()
	nonExistentID := uuid.New()

	_, err := s.repo.GetCategory(ctx, nonExistentID)
	s.Require().Error(err)
	s.ErrorIs(err, domain.ErrCategoryNotFound)
}

func (s *MongoRepoSuite) TestUpdateCategoryNotFound() {
	ctx := context.Background()
	nonExistentID := uuid.New()

	_, err := s.repo.UpdateCategory(ctx, nonExistentID, func(*domain.Category) (bool, error) {
		return true, nil
	})
	s.Require().Error(err)
	s.ErrorIs(err, domain.ErrCategoryNotFound)
}

func (s *MongoRepoSuite) TestDeleteCategoryNotFound() {
	ctx := context.Background()
	nonExistentID := uuid.New()

	err := s.repo.DeleteCategory(ctx, nonExistentID)
	s.Require().Error(err)
	s.ErrorIs(err, domain.ErrCategoryNotFound)
}

func (s *MongoRepoSuite) TestHierarchicalOperations() {
	ctx := context.Background()

	// Create complex hierarchy
	//   Root
	//   ├── IT
	//   │   ├── Hardware
	//   │   │   └── Laptops
	//   │   └── Software
	//   │       └── Applications
	//   └── HR
	//       └── Recruitment

	root, err := s.repo.CreateCategory(ctx, func() (*domain.Category, error) {
		return domain.CreateRootCategory("Root", "Root Category", s.orgID)
	})
	s.Require().NoError(err)

	it, err := s.repo.CreateCategory(ctx, func() (*domain.Category, error) {
		return domain.CreateSubCategory("IT", "IT Department", s.orgID, root.ID())
	})
	s.Require().NoError(err)

	hr, err := s.repo.CreateCategory(ctx, func() (*domain.Category, error) {
		return domain.CreateSubCategory("HR", "Human Resources", s.orgID, root.ID())
	})
	s.Require().NoError(err)

	hardware, err := s.repo.CreateCategory(ctx, func() (*domain.Category, error) {
		return domain.CreateSubCategory("Hardware", "Hardware Issues", s.orgID, it.ID())
	})
	s.Require().NoError(err)

	software, err := s.repo.CreateCategory(ctx, func() (*domain.Category, error) {
		return domain.CreateSubCategory("Software", "Software Issues", s.orgID, it.ID())
	})
	s.Require().NoError(err)

	_, err = s.repo.CreateCategory(ctx, func() (*domain.Category, error) {
		return domain.CreateSubCategory("Laptops", "Laptop Issues", s.orgID, hardware.ID())
	})
	s.Require().NoError(err)

	_, err = s.repo.CreateCategory(ctx, func() (*domain.Category, error) {
		return domain.CreateSubCategory("Applications", "Application Issues", s.orgID, software.ID())
	})
	s.Require().NoError(err)

	_, err = s.repo.CreateCategory(ctx, func() (*domain.Category, error) {
		return domain.CreateSubCategory("Recruitment", "Recruitment Process", s.orgID, hr.ID())
	})
	s.Require().NoError(err)

	// Test filtering by parent
	itID := it.ID()
	filter := queries.CategoryFilter{
		ParentID: &itID,
	}
	itChildren, err := s.repo.ListCategories(ctx, filter)
	s.Require().NoError(err)
	s.Len(itChildren, 2) // Hardware and Software

	// Test root-only filter
	filter = queries.CategoryFilter{
		IsRootOnly: true,
	}
	rootCategories, err := s.repo.ListCategories(ctx, filter)
	s.Require().NoError(err)
	s.Len(rootCategories, 1) // Only Root category
	s.Equal("Root", rootCategories[0].Name())

	// Get full hierarchy
	hierarchy, err := s.repo.GetCategoryHierarchy(ctx, root.ID())
	s.Require().NoError(err)
	s.Len(hierarchy.Children, 2) // IT and HR

	// Verify IT branch has 2 children
	var itBranch *application.CategoryTree
	for _, child := range hierarchy.Children {
		if child.Category.Name() == "IT" {
			itBranch = child
			break
		}
	}
	s.Require().NotNil(itBranch)
	s.Len(itBranch.Children, 2) // Hardware and Software
}

func (s *MongoRepoSuite) TestConcurrentCategoryCreation() {
	ctx := context.Background()

	const numGoroutines = 10
	results := make(chan error, numGoroutines)

	for i := range numGoroutines {
		go func(index int) {
			name := fmt.Sprintf("Concurrent Category %d", index)
			_, createErr := s.repo.CreateCategory(ctx, func() (*domain.Category, error) {
				return domain.CreateRootCategory(name, "Concurrent test", s.orgID)
			})
			results <- createErr
		}(i)
	}

	successCount := 0
	for range numGoroutines {
		resultErr := <-results
		if resultErr == nil {
			successCount++
		}
	}

	s.Equal(numGoroutines, successCount, "All concurrent category creations should succeed with unique names")
}

func (s *MongoRepoSuite) TestConcurrentHierarchyOperations() {
	ctx := context.Background()

	// Create root category
	root, err := s.repo.CreateCategory(ctx, func() (*domain.Category, error) {
		return domain.CreateRootCategory("Concurrent Root", "Root for concurrent test", s.orgID)
	})
	s.Require().NoError(err)

	const numGoroutines = 5
	results := make(chan error, numGoroutines)

	// Concurrent creation of sub-categories
	for i := range numGoroutines {
		go func(index int) {
			name := fmt.Sprintf("Child %d", index)
			_, createErr := s.repo.CreateCategory(ctx, func() (*domain.Category, error) {
				return domain.CreateSubCategory(name, "Child category", s.orgID, root.ID())
			})
			results <- createErr
		}(i)
	}

	successCount := 0
	for range numGoroutines {
		resultErr := <-results
		if resultErr == nil {
			successCount++
		}
	}

	s.Equal(numGoroutines, successCount, "All concurrent child category creations should succeed")

	// Verify hierarchy
	hierarchy, err := s.repo.GetCategoryHierarchy(ctx, root.ID())
	s.Require().NoError(err)
	s.Len(hierarchy.Children, numGoroutines)
}

func (s *MongoRepoSuite) TestPerformanceHierarchicalQueries() {
	if testing.Short() {
		s.T().Skip("Skipping performance test in short mode")
	}

	ctx := context.Background()

	// Create hierarchy with depth 5 and breadth 5 at each level
	const depth = 5
	const breadth = 5

	root, err := s.repo.CreateCategory(ctx, func() (*domain.Category, error) {
		return domain.CreateRootCategory("Performance Root", "Root for performance test", s.orgID)
	})
	s.Require().NoError(err)

	// Create categories level by level
	currentLevel := []*domain.Category{root}
	totalCategories := 1

	for d := 1; d < depth; d++ {
		var nextLevel []*domain.Category
		for _, parent := range currentLevel {
			for b := range breadth {
				name := fmt.Sprintf("L%d-P%s-C%d", d, parent.Name()[len(parent.Name())-10:], b)
				child, childErr := s.repo.CreateCategory(ctx, func() (*domain.Category, error) {
					return domain.CreateSubCategory(name, "Performance test category", s.orgID, parent.ID())
				})
				s.Require().NoError(childErr)
				nextLevel = append(nextLevel, child)
				totalCategories++
			}
		}
		currentLevel = nextLevel
	}

	s.T().Logf("Created hierarchy with %d categories", totalCategories)

	// Test hierarchy retrieval performance
	start := time.Now()
	hierarchy, err := s.repo.GetCategoryHierarchy(ctx, root.ID())
	s.Require().NoError(err)
	duration := time.Since(start)

	s.T().Logf("Retrieved full hierarchy (%d categories) in %v", totalCategories, duration)
	s.Less(duration.Seconds(), 5.0, "Hierarchy retrieval should take less than 5 seconds")
	s.NotNil(hierarchy)
	s.Equal(root.ID(), hierarchy.Category.ID())
}

func (s *MongoRepoSuite) TestLargeHierarchyOperations() {
	if testing.Short() {
		s.T().Skip("Skipping large hierarchy test in short mode")
	}

	ctx := context.Background()

	// Create a wide hierarchy (many categories at the same level)
	root, err := s.repo.CreateCategory(ctx, func() (*domain.Category, error) {
		return domain.CreateRootCategory("Large Root", "Root for large test", s.orgID)
	})
	s.Require().NoError(err)

	const numChildren = 100
	children := make([]*domain.Category, 0, numChildren)

	start := time.Now()
	for i := range numChildren {
		name := fmt.Sprintf("Child %03d", i)
		child, childCreateErr := s.repo.CreateCategory(ctx, func() (*domain.Category, error) {
			return domain.CreateSubCategory(name, "Large test child", s.orgID, root.ID())
		})
		s.Require().NoError(childCreateErr)
		children = append(children, child) //nolint:staticcheck // children is used in later test assertions

		if i%20 == 0 {
			s.T().Logf("Created %d/%d children", i+1, numChildren)
		}
	}
	createDuration := time.Since(start)
	s.T().Logf("Created %d children in %v", numChildren, createDuration)

	// Test listing children
	start = time.Now()
	rootID := root.ID()
	filter := queries.CategoryFilter{
		ParentID: &rootID,
	}
	result, err := s.repo.ListCategories(ctx, filter)
	s.Require().NoError(err)
	listDuration := time.Since(start)
	s.Len(result, numChildren)
	s.T().Logf("Listed %d children in %v", numChildren, listDuration)

	// Test hierarchy retrieval
	start = time.Now()
	hierarchy, err := s.repo.GetCategoryHierarchy(ctx, root.ID())
	s.Require().NoError(err)
	hierarchyDuration := time.Since(start)
	s.Len(hierarchy.Children, numChildren)
	s.T().Logf("Retrieved hierarchy with %d children in %v", numChildren, hierarchyDuration)

	s.Less(listDuration.Seconds(), 2.0, "Listing children should take less than 2 seconds")
	s.Less(hierarchyDuration.Seconds(), 5.0, "Hierarchy retrieval should take less than 5 seconds")
}

func TestMongoRepoSuite(t *testing.T) {
	suite.Run(t, new(MongoRepoSuite))
}
