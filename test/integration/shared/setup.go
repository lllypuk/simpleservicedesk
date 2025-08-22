//go:build integration
// +build integration

package shared

import (
	"context"
	"testing"

	"simpleservicedesk/internal/application"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go/modules/mongodb"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// IntegrationSuite provides common setup for all integration tests
type IntegrationSuite struct {
	suite.Suite
	HTTPServer *echo.Echo
	UsersRepo  application.UserRepository
}

// MongoIntegrationSuite extends IntegrationSuite with MongoDB setup
type MongoIntegrationSuite struct {
	IntegrationSuite
	MongoContainer *mongodb.MongoDBContainer
	MongoDB        *mongo.Database
	MongoClient    *mongo.Client
}

// SetupMongoTest sets up MongoDB testcontainer for integration tests
func SetupMongoTest(t *testing.T) (*mongo.Database, *mongo.Client, func()) {
	ctx := context.Background()

	// Start MongoDB container
	mongoContainer, err := mongodb.Run(ctx, "mongodb/mongodb-community-server:8.0-ubi8")
	require.NoError(t, err)

	// Get connection string
	endpoint, err := mongoContainer.ConnectionString(ctx)
	require.NoError(t, err)

	// Connect to MongoDB
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(endpoint))
	require.NoError(t, err)

	// Test connection
	err = client.Ping(ctx, nil)
	require.NoError(t, err)

	// Create test database
	db := client.Database("test_integration")

	cleanup := func() {
		client.Disconnect(ctx)
		mongoContainer.Terminate(ctx)
	}

	return db, client, cleanup
}

// SetupSuite initializes the integration test suite with mock repository
func (s *IntegrationSuite) SetupSuite() {
	// Initialize mock repository with fresh state
	s.UsersRepo = newMockUserRepository()

	// Initialize HTTP server with mock repository
	s.HTTPServer = application.SetupHTTPServer(s.UsersRepo)
}

// SetupTest runs before each test to ensure clean state
func (s *IntegrationSuite) SetupTest() {
	// Reset mock repository state for each test
	s.UsersRepo = newMockUserRepository()
	s.HTTPServer = application.SetupHTTPServer(s.UsersRepo)
}

// SetupSuite initializes the MongoDB integration test suite
func (s *MongoIntegrationSuite) SetupSuite() {
	var cleanup func()
	s.MongoDB, s.MongoClient, cleanup = SetupMongoTest(s.T())

	// Store cleanup function for TearDownSuite
	s.T().Cleanup(cleanup)
}

// SetupTest runs before each test to ensure clean database state
func (s *MongoIntegrationSuite) SetupTest() {
	// Clear all collections for clean test state
	ctx := context.Background()
	collections, err := s.MongoDB.ListCollectionNames(ctx, map[string]interface{}{})
	s.Require().NoError(err)

	for _, collection := range collections {
		err = s.MongoDB.Collection(collection).Drop(ctx)
		s.Require().NoError(err)
	}
}
