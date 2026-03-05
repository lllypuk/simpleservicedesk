//go:build integration
// +build integration

package shared

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"simpleservicedesk/internal/application"
	authdomain "simpleservicedesk/internal/domain/auth"
	userdomain "simpleservicedesk/internal/domain/users"
	"simpleservicedesk/internal/infrastructure/categories"
	"simpleservicedesk/internal/infrastructure/organizations"
	"simpleservicedesk/internal/infrastructure/tickets"
	userrepo "simpleservicedesk/internal/infrastructure/users"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go/modules/mongodb"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// IntegrationSuite provides common setup for all integration tests with real MongoDB repositories
type IntegrationSuite struct {
	suite.Suite
	HTTPServer        *echo.Echo
	UsersRepo         application.UserRepository
	TicketsRepo       application.TicketRepository
	OrganizationsRepo application.OrganizationRepository
	CategoriesRepo    application.CategoryRepository
	MongoContainer    *mongodb.MongoDBContainer
	MongoDB           *mongo.Database
	MongoClient       *mongo.Client
}

// MongoIntegrationSuite extends IntegrationSuite with MongoDB setup
type MongoIntegrationSuite struct {
	IntegrationSuite
}

const (
	testBypassHeaderKey = "X-Test-Bypass"
	testAuthUserID      = "00000000-0000-0000-0000-000000000001"
)

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

// SetupSuite initializes the integration test suite with real MongoDB repositories
func (s *IntegrationSuite) SetupSuite() {
	var cleanup func()
	s.MongoDB, s.MongoClient, cleanup = SetupMongoTest(s.T())

	// Store cleanup function for TearDownSuite
	s.T().Cleanup(cleanup)

	// Initialize real repositories
	s.UsersRepo = userrepo.NewMongoRepo(s.MongoDB)
	s.TicketsRepo = tickets.NewMongoRepo(s.MongoDB)
	s.OrganizationsRepo = organizations.NewMongoRepo(s.MongoDB)
	s.CategoriesRepo = categories.NewMongoRepo(s.MongoDB)

	// Initialize HTTP server with real repositories
	s.HTTPServer = application.SetupHTTPServer(
		s.UsersRepo,
		s.TicketsRepo,
		s.OrganizationsRepo,
		s.CategoriesRepo,
		"integration-test-jwt-signing-key",
		time.Hour,
	)
	attachDefaultTestAuthHeader(s.HTTPServer, "integration-test-jwt-signing-key", userdomain.RoleAdmin)
}

// SetupTest runs before each test to ensure clean database state
func (s *IntegrationSuite) SetupTest() {
	// Clear all collections for clean test state
	ctx := context.Background()
	collections, err := s.MongoDB.ListCollectionNames(ctx, map[string]interface{}{})
	s.Require().NoError(err)

	for _, collection := range collections {
		err = s.MongoDB.Collection(collection).Drop(ctx)
		s.Require().NoError(err)
	}

	// Re-initialize HTTP server to ensure clean state
	s.HTTPServer = application.SetupHTTPServer(
		s.UsersRepo,
		s.TicketsRepo,
		s.OrganizationsRepo,
		s.CategoriesRepo,
		"integration-test-jwt-signing-key",
		time.Hour,
	)
	attachDefaultTestAuthHeader(s.HTTPServer, "integration-test-jwt-signing-key", userdomain.RoleAdmin)
}

// SetupSuite initializes the MongoDB integration test suite
func (s *MongoIntegrationSuite) SetupSuite() {
	// Call parent SetupSuite which handles MongoDB setup
	s.IntegrationSuite.SetupSuite()
}

// SetupTest runs before each test to ensure clean database state
func (s *MongoIntegrationSuite) SetupTest() {
	// Call parent SetupTest which handles database cleanup
	s.IntegrationSuite.SetupTest()
}

func attachDefaultTestAuthHeader(server *echo.Echo, signingKey string, role userdomain.Role) {
	server.Pre(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if strings.EqualFold(strings.TrimSpace(c.Request().Header.Get(testBypassHeaderKey)), "true") {
				return next(c)
			}

			if strings.TrimSpace(c.Request().Header.Get(echo.HeaderAuthorization)) == "" {
				token, err := createTestToken(signingKey, testAuthUserID, role)
				if err != nil {
					return echo.NewHTTPError(http.StatusInternalServerError, "failed to create test auth token")
				}

				c.Request().Header.Set(echo.HeaderAuthorization, fmt.Sprintf("Bearer %s", token))
			}

			return next(c)
		}
	})
}

func createTestToken(signingKey, userID string, role userdomain.Role) (string, error) {
	issuedAt := time.Now().UTC()
	claims := authdomain.Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(issuedAt),
			ExpiresAt: jwt.NewNumericDate(issuedAt.Add(time.Hour)),
		},
		UserID: userID,
		Role:   role,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(signingKey))
}
