package organizations_test

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
	domain "simpleservicedesk/internal/domain/organizations"
	organizationsInfra "simpleservicedesk/internal/infrastructure/organizations"
	"simpleservicedesk/internal/queries"
)

type MongoRepoSuite struct {
	suite.Suite

	container testcontainers.Container
	db        *mongo.Database
	repo      *organizationsInfra.MongoRepo
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
	s.repo = organizationsInfra.NewMongoRepo(s.db)
}

func (s *MongoRepoSuite) TearDownSuite() {
	ctx := context.Background()
	err := s.db.Client().Disconnect(ctx)
	s.Require().NoError(err)
	err = s.container.Terminate(ctx)
	s.Require().NoError(err)
}

func (s *MongoRepoSuite) SetupTest() {
	err := s.db.Collection("organizations").Drop(context.Background())
	s.Require().NoError(err)
}

func (s *MongoRepoSuite) TestCreateAndGetOrganization() {
	ctx := context.Background()
	name := "Acme Corporation"
	domainName := "acme.com"

	createdOrg, err := s.repo.CreateOrganization(ctx, func() (*domain.Organization, error) {
		return domain.CreateRootOrganization(name, domainName)
	})
	s.Require().NoError(err)
	s.Require().NotNil(createdOrg)
	s.Equal(name, createdOrg.Name())
	s.Equal(domainName, createdOrg.Domain())
	s.True(createdOrg.IsRootOrganization())

	fetchedOrg, err := s.repo.GetOrganization(ctx, createdOrg.ID())
	s.Require().NoError(err)
	s.Require().NotNil(fetchedOrg)
	s.Equal(createdOrg.ID(), fetchedOrg.ID())
	s.Equal(createdOrg.Name(), fetchedOrg.Name())
	s.Equal(createdOrg.Domain(), fetchedOrg.Domain())
	s.True(fetchedOrg.IsRootOrganization())
}

func (s *MongoRepoSuite) TestCreateSubOrganization() {
	ctx := context.Background()

	// Create parent organization
	parentOrg, err := s.repo.CreateOrganization(ctx, func() (*domain.Organization, error) {
		return domain.CreateRootOrganization("Parent Corp", "parent.com")
	})
	s.Require().NoError(err)

	// Create sub-organization
	subOrg, err := s.repo.CreateOrganization(ctx, func() (*domain.Organization, error) {
		return domain.CreateSubOrganization("IT Department", "it.parent.com", parentOrg.ID())
	})
	s.Require().NoError(err)
	s.Require().NotNil(subOrg)
	s.Equal("IT Department", subOrg.Name())
	s.False(subOrg.IsRootOrganization())
	s.True(subOrg.HasParent())
	s.Equal(parentOrg.ID(), *subOrg.ParentID())
}

func (s *MongoRepoSuite) TestCreateOrganizationWithNonExistentParent() {
	ctx := context.Background()
	nonExistentParentID := uuid.New()

	_, err := s.repo.CreateOrganization(ctx, func() (*domain.Organization, error) {
		return domain.CreateSubOrganization("Child", "child.com", nonExistentParentID)
	})
	s.Require().Error(err)
	s.Contains(err.Error(), "parent organization not found")
}

func (s *MongoRepoSuite) TestCreateDuplicateOrganization() {
	ctx := context.Background()
	name := "Duplicate Corp"
	domainName := "duplicate.com"

	// Create first organization
	_, err := s.repo.CreateOrganization(ctx, func() (*domain.Organization, error) {
		return domain.CreateRootOrganization(name, domainName)
	})
	s.Require().NoError(err)

	// Try to create duplicate organization with same name
	_, err = s.repo.CreateOrganization(ctx, func() (*domain.Organization, error) {
		return domain.CreateRootOrganization(name, "different.com")
	})
	s.Require().Error(err)
	s.ErrorIs(err, domain.ErrOrganizationAlreadyExist)
}

func (s *MongoRepoSuite) TestUpdateOrganization() {
	ctx := context.Background()

	org, err := s.repo.CreateOrganization(ctx, func() (*domain.Organization, error) {
		return domain.CreateRootOrganization("Original Name", "original.com")
	})
	s.Require().NoError(err)

	newName := "Updated Name"
	newDomain := "updated.com"

	updatedOrg, err := s.repo.UpdateOrganization(ctx, org.ID(), func(o *domain.Organization) (bool, error) {
		changeNameErr := o.ChangeName(newName)
		if changeNameErr != nil {
			return false, changeNameErr
		}
		changeDomainErr := o.ChangeDomain(newDomain)
		if changeDomainErr != nil {
			return false, changeDomainErr
		}
		return true, nil
	})

	s.Require().NoError(err)
	s.Equal(newName, updatedOrg.Name())
	s.Equal(newDomain, updatedOrg.Domain())

	// Verify persistence
	fetchedOrg, err := s.repo.GetOrganization(ctx, org.ID())
	s.Require().NoError(err)
	s.Equal(newName, fetchedOrg.Name())
	s.Equal(newDomain, fetchedOrg.Domain())
}

func (s *MongoRepoSuite) TestUpdateOrganizationParent() {
	ctx := context.Background()

	// Create organizations
	rootOrg, err := s.repo.CreateOrganization(ctx, func() (*domain.Organization, error) {
		return domain.CreateRootOrganization("Root", "root.com")
	})
	s.Require().NoError(err)

	newParent, err := s.repo.CreateOrganization(ctx, func() (*domain.Organization, error) {
		return domain.CreateRootOrganization("New Parent", "newparent.com")
	})
	s.Require().NoError(err)

	// Update parent
	newParentID := newParent.ID()
	updatedOrg, err := s.repo.UpdateOrganization(ctx, rootOrg.ID(), func(o *domain.Organization) (bool, error) {
		return true, o.ChangeParent(&newParentID)
	})

	s.Require().NoError(err)
	s.True(updatedOrg.HasParent())
	s.Equal(newParent.ID(), *updatedOrg.ParentID())
}

func (s *MongoRepoSuite) TestUpdateOrganizationCircularReference() {
	ctx := context.Background()

	// Create hierarchy: parent -> child
	parent, err := s.repo.CreateOrganization(ctx, func() (*domain.Organization, error) {
		return domain.CreateRootOrganization("Parent", "parent.com")
	})
	s.Require().NoError(err)

	child, err := s.repo.CreateOrganization(ctx, func() (*domain.Organization, error) {
		return domain.CreateSubOrganization("Child", "child.com", parent.ID())
	})
	s.Require().NoError(err)

	// Try to make parent a child of child (circular reference)
	childID := child.ID()
	_, err = s.repo.UpdateOrganization(ctx, parent.ID(), func(o *domain.Organization) (bool, error) {
		return true, o.ChangeParent(&childID)
	})

	s.Require().Error(err)
	s.ErrorIs(err, domain.ErrCircularReference)
}

func (s *MongoRepoSuite) TestListOrganizations() {
	ctx := context.Background()

	// Create test organizations
	organizations := []struct {
		name     string
		domain   string
		isActive bool
	}{
		{"Tech Corp", "tech.com", true},
		{"Finance LLC", "finance.com", true},
		{"Marketing Inc", "marketing.com", false},
	}

	createdOrgs := make([]*domain.Organization, len(organizations))
	for i, org := range organizations {
		organization, err := s.repo.CreateOrganization(ctx, func() (*domain.Organization, error) {
			return domain.CreateRootOrganization(org.name, org.domain)
		})
		s.Require().NoError(err)

		if !org.isActive {
			_, err = s.repo.UpdateOrganization(ctx, organization.ID(), func(o *domain.Organization) (bool, error) {
				o.Deactivate()
				return true, nil
			})
			s.Require().NoError(err)
		}
		createdOrgs[i] = organization
	}

	// Test listing all organizations
	filter := queries.OrganizationFilter{}
	result, err := s.repo.ListOrganizations(ctx, filter)
	s.Require().NoError(err)
	s.Len(result, 3)

	// Test filtering by active status
	active := true
	filter.IsActive = &active
	result, err = s.repo.ListOrganizations(ctx, filter)
	s.Require().NoError(err)
	s.Len(result, 2)

	// Test filtering by name
	name := "Tech"
	filter.Name = &name
	filter.IsActive = nil
	result, err = s.repo.ListOrganizations(ctx, filter)
	s.Require().NoError(err)
	s.Len(result, 1)
	s.Equal("Tech Corp", result[0].Name())

	// Test filtering by domain
	domainName := "finance"
	filter.Domain = &domainName
	filter.Name = nil
	result, err = s.repo.ListOrganizations(ctx, filter)
	s.Require().NoError(err)
	s.Len(result, 1)
	s.Equal("Finance LLC", result[0].Name())
}

func (s *MongoRepoSuite) TestListOrganizationsWithPagination() {
	ctx := context.Background()

	// Create multiple organizations
	for i := range 10 {
		_, err := s.repo.CreateOrganization(ctx, func() (*domain.Organization, error) {
			return domain.CreateRootOrganization(fmt.Sprintf("Organization %02d", i), fmt.Sprintf("org%d.com", i))
		})
		s.Require().NoError(err)
	}

	// Test pagination
	filter := queries.OrganizationFilter{}
	filter.Limit = 5
	filter.Offset = 0
	filter.SortBy = "name"
	filter.SortOrder = "asc"

	firstPage, err := s.repo.ListOrganizations(ctx, filter)
	s.Require().NoError(err)
	s.Len(firstPage, 5)

	filter.Offset = 5
	secondPage, err := s.repo.ListOrganizations(ctx, filter)
	s.Require().NoError(err)
	s.Len(secondPage, 5)

	// Verify no overlap
	for _, org1 := range firstPage {
		for _, org2 := range secondPage {
			s.NotEqual(org1.ID(), org2.ID())
		}
	}
}

func (s *MongoRepoSuite) TestGetOrganizationHierarchy() {
	ctx := context.Background()

	// Create hierarchy: Root -> IT -> Development
	root, err := s.repo.CreateOrganization(ctx, func() (*domain.Organization, error) {
		return domain.CreateRootOrganization("Root Corp", "root.com")
	})
	s.Require().NoError(err)

	it, err := s.repo.CreateOrganization(ctx, func() (*domain.Organization, error) {
		return domain.CreateSubOrganization("IT Department", "it.root.com", root.ID())
	})
	s.Require().NoError(err)

	development, err := s.repo.CreateOrganization(ctx, func() (*domain.Organization, error) {
		return domain.CreateSubOrganization("Development", "dev.root.com", it.ID())
	})
	s.Require().NoError(err)

	// Create another branch: Root -> HR
	_, err = s.repo.CreateOrganization(ctx, func() (*domain.Organization, error) {
		return domain.CreateSubOrganization("HR Department", "hr.root.com", root.ID())
	})
	s.Require().NoError(err)

	// Get hierarchy
	hierarchy, err := s.repo.GetOrganizationHierarchy(ctx, root.ID())
	s.Require().NoError(err)
	s.Require().NotNil(hierarchy)

	// Verify root
	s.Equal(root.ID(), hierarchy.Organization.ID())
	s.Len(hierarchy.Children, 2) // IT and HR

	// Find IT branch
	var itBranch *application.OrganizationTree
	for _, child := range hierarchy.Children {
		if child.Organization.Name() == "IT Department" {
			itBranch = child
			break
		}
	}
	s.Require().NotNil(itBranch)
	s.Equal(it.ID(), itBranch.Organization.ID())
	s.Len(itBranch.Children, 1) // Development

	// Verify development
	s.Equal(development.ID(), itBranch.Children[0].Organization.ID())
	s.Empty(itBranch.Children[0].Children) // No further children
}

func (s *MongoRepoSuite) TestDeleteOrganization() {
	ctx := context.Background()

	org, err := s.repo.CreateOrganization(ctx, func() (*domain.Organization, error) {
		return domain.CreateRootOrganization("Delete Test", "delete.com")
	})
	s.Require().NoError(err)

	// Delete organization
	err = s.repo.DeleteOrganization(ctx, org.ID())
	s.Require().NoError(err)

	// Verify deletion
	_, err = s.repo.GetOrganization(ctx, org.ID())
	s.Require().Error(err)
	s.ErrorIs(err, domain.ErrOrganizationNotFound)
}

func (s *MongoRepoSuite) TestDeleteOrganizationWithChildren() {
	ctx := context.Background()

	parent, err := s.repo.CreateOrganization(ctx, func() (*domain.Organization, error) {
		return domain.CreateRootOrganization("Parent", "parent.com")
	})
	s.Require().NoError(err)

	_, err = s.repo.CreateOrganization(ctx, func() (*domain.Organization, error) {
		return domain.CreateSubOrganization("Child", "child.com", parent.ID())
	})
	s.Require().NoError(err)

	// Try to delete parent with children
	err = s.repo.DeleteOrganization(ctx, parent.ID())
	s.Require().Error(err)
	s.Contains(err.Error(), "cannot delete organization with children")
}

func (s *MongoRepoSuite) TestGetOrganizationNotFound() {
	ctx := context.Background()
	nonExistentID := uuid.New()

	_, err := s.repo.GetOrganization(ctx, nonExistentID)
	s.Require().Error(err)
	s.ErrorIs(err, domain.ErrOrganizationNotFound)
}

func (s *MongoRepoSuite) TestUpdateOrganizationNotFound() {
	ctx := context.Background()
	nonExistentID := uuid.New()

	_, err := s.repo.UpdateOrganization(ctx, nonExistentID, func(*domain.Organization) (bool, error) {
		return true, nil
	})
	s.Require().Error(err)
	s.ErrorIs(err, domain.ErrOrganizationNotFound)
}

func (s *MongoRepoSuite) TestDeleteOrganizationNotFound() {
	ctx := context.Background()
	nonExistentID := uuid.New()

	err := s.repo.DeleteOrganization(ctx, nonExistentID)
	s.Require().Error(err)
	s.ErrorIs(err, domain.ErrOrganizationNotFound)
}

func (s *MongoRepoSuite) TestHierarchicalOperations() {
	ctx := context.Background()

	// Create complex hierarchy
	//   Corporation
	//   ├── Technology
	//   │   ├── Development
	//   │   │   └── Backend
	//   │   └── QA
	//   └── Business
	//       └── Sales

	corp, err := s.repo.CreateOrganization(ctx, func() (*domain.Organization, error) {
		return domain.CreateRootOrganization("Corporation", "corp.com")
	})
	s.Require().NoError(err)

	tech, err := s.repo.CreateOrganization(ctx, func() (*domain.Organization, error) {
		return domain.CreateSubOrganization("Technology", "tech.corp.com", corp.ID())
	})
	s.Require().NoError(err)

	business, err := s.repo.CreateOrganization(ctx, func() (*domain.Organization, error) {
		return domain.CreateSubOrganization("Business", "business.corp.com", corp.ID())
	})
	s.Require().NoError(err)

	development, err := s.repo.CreateOrganization(ctx, func() (*domain.Organization, error) {
		return domain.CreateSubOrganization("Development", "dev.corp.com", tech.ID())
	})
	s.Require().NoError(err)

	_, err = s.repo.CreateOrganization(ctx, func() (*domain.Organization, error) {
		return domain.CreateSubOrganization("QA", "qa.corp.com", tech.ID())
	})
	s.Require().NoError(err)

	_, err = s.repo.CreateOrganization(ctx, func() (*domain.Organization, error) {
		return domain.CreateSubOrganization("Backend", "backend.corp.com", development.ID())
	})
	s.Require().NoError(err)

	_, err = s.repo.CreateOrganization(ctx, func() (*domain.Organization, error) {
		return domain.CreateSubOrganization("Sales", "sales.corp.com", business.ID())
	})
	s.Require().NoError(err)

	// Test filtering by parent
	techID := tech.ID()
	filter := queries.OrganizationFilter{
		ParentID: &techID,
	}
	techChildren, err := s.repo.ListOrganizations(ctx, filter)
	s.Require().NoError(err)
	s.Len(techChildren, 2) // Development and QA

	// Test root-only filter
	filter = queries.OrganizationFilter{
		IsRootOnly: true,
	}
	rootOrgs, err := s.repo.ListOrganizations(ctx, filter)
	s.Require().NoError(err)
	s.Len(rootOrgs, 1) // Only Corporation
	s.Equal("Corporation", rootOrgs[0].Name())

	// Get full hierarchy
	hierarchy, err := s.repo.GetOrganizationHierarchy(ctx, corp.ID())
	s.Require().NoError(err)
	s.Len(hierarchy.Children, 2) // Technology and Business

	// Verify Technology branch has 2 children
	var techBranch *application.OrganizationTree
	for _, child := range hierarchy.Children {
		if child.Organization.Name() == "Technology" {
			techBranch = child
			break
		}
	}
	s.Require().NotNil(techBranch)
	s.Len(techBranch.Children, 2) // Development and QA
}

func (s *MongoRepoSuite) TestConcurrentOrganizationCreation() {
	ctx := context.Background()

	const numGoroutines = 10
	results := make(chan error, numGoroutines)

	for i := range numGoroutines {
		go func(index int) {
			name := fmt.Sprintf("Concurrent Org %d", index)
			domainName := fmt.Sprintf("concurrent%d.com", index)
			_, createErr := s.repo.CreateOrganization(ctx, func() (*domain.Organization, error) {
				return domain.CreateRootOrganization(name, domainName)
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

	s.Equal(numGoroutines, successCount, "All concurrent organization creations should succeed with unique names")
}

func (s *MongoRepoSuite) TestConcurrentHierarchyOperations() {
	ctx := context.Background()

	// Create root organization
	root, err := s.repo.CreateOrganization(ctx, func() (*domain.Organization, error) {
		return domain.CreateRootOrganization("Concurrent Root", "root.concurrent.com")
	})
	s.Require().NoError(err)

	const numGoroutines = 5
	results := make(chan error, numGoroutines)

	// Concurrent creation of sub-organizations
	for i := range numGoroutines {
		go func(index int) {
			name := fmt.Sprintf("Dept %d", index)
			domainName := fmt.Sprintf("dept%d.concurrent.com", index)
			_, createErr := s.repo.CreateOrganization(ctx, func() (*domain.Organization, error) {
				return domain.CreateSubOrganization(name, domainName, root.ID())
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

	s.Equal(numGoroutines, successCount, "All concurrent child organization creations should succeed")

	// Verify hierarchy
	hierarchy, err := s.repo.GetOrganizationHierarchy(ctx, root.ID())
	s.Require().NoError(err)
	s.Len(hierarchy.Children, numGoroutines)
}

func (s *MongoRepoSuite) TestPerformanceHierarchicalQueries() {
	if testing.Short() {
		s.T().Skip("Skipping performance test in short mode")
	}

	ctx := context.Background()

	// Create hierarchy with depth 4 and breadth 3 at each level
	const depth = 4
	const breadth = 3

	root, err := s.repo.CreateOrganization(ctx, func() (*domain.Organization, error) {
		return domain.CreateRootOrganization("Performance Root", "perf.root.com")
	})
	s.Require().NoError(err)

	// Create organizations level by level
	currentLevel := []*domain.Organization{root}
	totalOrgs := 1

	for d := 1; d < depth; d++ {
		var nextLevel []*domain.Organization
		for _, parent := range currentLevel {
			for b := range breadth {
				name := fmt.Sprintf("L%d-P%s-O%d", d, parent.Name()[len(parent.Name())-10:], b)
				domainName := fmt.Sprintf("l%d.p%d.perf.com", d, b)
				child, childErr := s.repo.CreateOrganization(ctx, func() (*domain.Organization, error) {
					return domain.CreateSubOrganization(name, domainName, parent.ID())
				})
				s.Require().NoError(childErr)
				nextLevel = append(nextLevel, child)
				totalOrgs++
			}
		}
		currentLevel = nextLevel
	}

	s.T().Logf("Created hierarchy with %d organizations", totalOrgs)

	// Test hierarchy retrieval performance
	start := time.Now()
	hierarchy, err := s.repo.GetOrganizationHierarchy(ctx, root.ID())
	s.Require().NoError(err)
	duration := time.Since(start)

	s.T().Logf("Retrieved full hierarchy (%d organizations) in %v", totalOrgs, duration)
	s.Less(duration.Seconds(), 5.0, "Hierarchy retrieval should take less than 5 seconds")
	s.NotNil(hierarchy)
	s.Equal(root.ID(), hierarchy.Organization.ID())
}

func (s *MongoRepoSuite) TestLargeHierarchyOperations() {
	if testing.Short() {
		s.T().Skip("Skipping large hierarchy test in short mode")
	}

	ctx := context.Background()

	// Create a wide hierarchy (many organizations at the same level)
	root, err := s.repo.CreateOrganization(ctx, func() (*domain.Organization, error) {
		return domain.CreateRootOrganization("Large Root", "large.root.com")
	})
	s.Require().NoError(err)

	const numChildren = 100
	children := make([]*domain.Organization, 0, numChildren)

	start := time.Now()
	for i := range numChildren {
		name := fmt.Sprintf("Child %03d", i)
		domainName := fmt.Sprintf("child%d.large.com", i)
		child, childCreateErr := s.repo.CreateOrganization(ctx, func() (*domain.Organization, error) {
			return domain.CreateSubOrganization(name, domainName, root.ID())
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
	filter := queries.OrganizationFilter{
		ParentID: &rootID,
	}
	result, err := s.repo.ListOrganizations(ctx, filter)
	s.Require().NoError(err)
	listDuration := time.Since(start)
	s.Len(result, numChildren)
	s.T().Logf("Listed %d children in %v", numChildren, listDuration)

	// Test hierarchy retrieval
	start = time.Now()
	hierarchy, err := s.repo.GetOrganizationHierarchy(ctx, root.ID())
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
