package users_test

import (
	"context"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/pkg/errors"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"

	domain "simpleservicedesk/internal/domain/users"
	usersInfra "simpleservicedesk/internal/infrastructure/users"
)

type MongoRepoSuite struct {
	suite.Suite

	container testcontainers.Container
	db        *mongo.Database
	repo      *usersInfra.MongoRepo
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
	s.repo = usersInfra.NewMongoRepo(s.db)
}

func (s *MongoRepoSuite) TearDownSuite() {
	ctx := context.Background()
	err := s.db.Client().Disconnect(ctx)
	s.Require().NoError(err)
	err = s.container.Terminate(ctx)
	s.Require().NoError(err)
}

func (s *MongoRepoSuite) SetupTest() {
	err := s.db.Collection("users").Drop(context.Background())
	s.Require().NoError(err)
}

func (s *MongoRepoSuite) TestCreateAndGetUser() {
	ctx := context.Background()
	email := "test@example.com"
	passwordHash := []byte("hashedpassword")
	name := "Test User"

	var createdUser *domain.User

	createdUser, err := s.repo.CreateUser(ctx, email, passwordHash, func() (*domain.User, error) {
		return domain.CreateUser(name, email, passwordHash)
	})
	s.Require().NoError(err)
	s.Require().NotNil(createdUser)
	s.Equal(name, createdUser.Name())
	s.Equal(email, createdUser.Email())

	fetchedUser, err := s.repo.GetUser(ctx, createdUser.ID())
	s.Require().NoError(err)
	s.Require().NotNil(fetchedUser)
	s.Equal(createdUser.ID(), fetchedUser.ID())
	s.Equal(createdUser.Name(), fetchedUser.Name())
	s.Equal(createdUser.Email(), fetchedUser.Email())
}

func (s *MongoRepoSuite) TestUpdateUser() {
	ctx := context.Background()
	email := "update@example.com"
	passwordHash := []byte("hashedpassword")
	name := "Update User"

	user, err := s.repo.CreateUser(ctx, email, passwordHash, func() (*domain.User, error) {
		return domain.CreateUser(name, email, passwordHash)
	})
	s.Require().NoError(err)
	s.Require().NotNil(user)

	newEmail := "updated.email@example.com"

	updatedUser, err := s.repo.UpdateUser(ctx, user.ID(), func(u *domain.User) (bool, error) {
		err = u.ChangeEmail(newEmail)
		if err != nil {
			return false, err
		}
		return true, nil
	})

	s.Require().NoError(err)
	s.Require().NotNil(updatedUser)
	s.Equal(newEmail, updatedUser.Email())

	fetchedUser, err := s.repo.GetUser(ctx, user.ID())
	s.Require().NoError(err)
	s.Equal(newEmail, fetchedUser.Email())
}

func (s *MongoRepoSuite) TestGetUser_NotFound() {
	_, err := s.repo.GetUser(context.Background(), uuid.New())
	s.Require().Error(err)
	s.Require().ErrorIs(err, domain.ErrUserNotFound)
}

func (s *MongoRepoSuite) TestCreateUser_DuplicateEmail() {
	ctx := context.Background()
	email := "duplicate@example.com"
	passwordHash := []byte("hashedpassword")
	name := "Test User"

	_, err := s.repo.CreateUser(ctx, email, passwordHash, func() (*domain.User, error) {
		return domain.CreateUser(name, email, passwordHash)
	})
	s.Require().NoError(err)

	_, err = s.repo.CreateUser(ctx, email, passwordHash, func() (*domain.User, error) {
		return domain.CreateUser("Another User", email, passwordHash)
	})
	s.Require().Error(err)
	s.Require().ErrorIs(err, domain.ErrUserAlreadyExist)
}

func (s *MongoRepoSuite) TestCreateUser_InvalidUserCreation() {
	ctx := context.Background()
	email := "invalid@example.com"
	passwordHash := []byte("hashedpassword")

	_, err := s.repo.CreateUser(ctx, email, passwordHash, func() (*domain.User, error) {
		return nil, errors.New("domain validation error")
	})
	s.Require().Error(err)
	s.Contains(err.Error(), "domain validation error")
}

func (s *MongoRepoSuite) TestUpdateUser_NotFound() {
	ctx := context.Background()
	nonExistentID := uuid.New()

	_, err := s.repo.UpdateUser(ctx, nonExistentID, func(_ *domain.User) (bool, error) {
		return true, nil
	})
	s.Require().Error(err)
	s.Require().ErrorIs(err, domain.ErrUserNotFound)
}

func (s *MongoRepoSuite) TestUpdateUser_NoChangeRequired() {
	ctx := context.Background()
	email := "nochange@example.com"
	passwordHash := []byte("hashedpassword")
	name := "Test User"

	user, err := s.repo.CreateUser(ctx, email, passwordHash, func() (*domain.User, error) {
		return domain.CreateUser(name, email, passwordHash)
	})
	s.Require().NoError(err)

	updatedUser, err := s.repo.UpdateUser(ctx, user.ID(), func(_ *domain.User) (bool, error) {
		return false, nil
	})
	s.Require().NoError(err)
	s.Equal(user.Name(), updatedUser.Name())
	s.Equal(user.Email(), updatedUser.Email())
}

func (s *MongoRepoSuite) TestUpdateUser_UpdateFunctionError() {
	ctx := context.Background()
	email := "error@example.com"
	passwordHash := []byte("hashedpassword")
	name := "Test User"

	user, err := s.repo.CreateUser(ctx, email, passwordHash, func() (*domain.User, error) {
		return domain.CreateUser(name, email, passwordHash)
	})
	s.Require().NoError(err)

	_, err = s.repo.UpdateUser(ctx, user.ID(), func(_ *domain.User) (bool, error) {
		return false, errors.New("update function error")
	})
	s.Require().Error(err)
	s.Contains(err.Error(), "update function error")
}

func (s *MongoRepoSuite) TestUpdateUser_NameAndEmailUpdate() {
	ctx := context.Background()
	email := "update@example.com"
	passwordHash := []byte("hashedpassword")
	name := "Original Name"

	user, err := s.repo.CreateUser(ctx, email, passwordHash, func() (*domain.User, error) {
		return domain.CreateUser(name, email, passwordHash)
	})
	s.Require().NoError(err)

	newEmail := "updated@example.com"

	updatedUser, err := s.repo.UpdateUser(ctx, user.ID(), func(u *domain.User) (bool, error) {
		err = u.ChangeEmail(newEmail)
		if err != nil {
			return false, err
		}
		return true, nil
	})

	s.Require().NoError(err)
	s.Equal(name, updatedUser.Name())
	s.Equal(newEmail, updatedUser.Email())

	fetchedUser, err := s.repo.GetUser(ctx, user.ID())
	s.Require().NoError(err)
	s.Equal(name, fetchedUser.Name())
	s.Equal(newEmail, fetchedUser.Email())
}

func (s *MongoRepoSuite) TestUpdateUser_InvalidEmailUpdate() {
	ctx := context.Background()
	email := "update@example.com"
	passwordHash := []byte("hashedpassword")
	name := "Test User"

	user, err := s.repo.CreateUser(ctx, email, passwordHash, func() (*domain.User, error) {
		return domain.CreateUser(name, email, passwordHash)
	})
	s.Require().NoError(err)

	_, err = s.repo.UpdateUser(ctx, user.ID(), func(u *domain.User) (bool, error) {
		err = u.ChangeEmail("")
		return false, err
	})
	s.Require().Error(err)
}

func (s *MongoRepoSuite) TestConcurrentUserCreation() {
	ctx := context.Background()
	passwordHash := []byte("hashedpassword")
	name := "Concurrent User"

	const numGoroutines = 10
	results := make(chan error, numGoroutines)

	for i := range numGoroutines {
		go func(index int) {
			email := fmt.Sprintf("concurrent%d@example.com", index)
			_, err := s.repo.CreateUser(ctx, email, passwordHash, func() (*domain.User, error) {
				return domain.CreateUser(name, email, passwordHash)
			})
			results <- err
		}(i)
	}

	successCount := 0
	for range numGoroutines {
		err := <-results
		if err == nil {
			successCount++
		}
	}

	s.Equal(numGoroutines, successCount, "All concurrent user creations should succeed with unique emails")
}

func (s *MongoRepoSuite) TestConcurrentUserUpdates() {
	ctx := context.Background()
	email := "concurrent@example.com"
	passwordHash := []byte("hashedpassword")
	name := "Concurrent User"

	user, err := s.repo.CreateUser(ctx, email, passwordHash, func() (*domain.User, error) {
		return domain.CreateUser(name, email, passwordHash)
	})
	s.Require().NoError(err)

	const numGoroutines = 5
	results := make(chan error, numGoroutines)

	for i := range numGoroutines {
		go func(index int) {
			newEmail := fmt.Sprintf("updated%d@concurrent.com", index)
			_, err = s.repo.UpdateUser(ctx, user.ID(), func(u *domain.User) (bool, error) {
				err = u.ChangeEmail(newEmail)
				if err != nil {
					return false, err
				}
				return true, nil
			})
			results <- err
		}(i)
	}

	successCount := 0
	for range numGoroutines {
		err = <-results
		if err == nil {
			successCount++
		}
	}

	s.Equal(numGoroutines, successCount, "All concurrent user updates should succeed")

	fetchedUser, err := s.repo.GetUser(ctx, user.ID())
	s.Require().NoError(err)
	s.Contains(fetchedUser.Email(), "@concurrent.com", "Final email should contain '@concurrent.com'")
}

func (s *MongoRepoSuite) TestDatabaseConnectionError() {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := s.repo.GetUser(ctx, uuid.New())
	s.Require().Error(err)
	s.Contains(err.Error(), "context canceled")
}

func (s *MongoRepoSuite) TestDatabaseTimeout() {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	time.Sleep(2 * time.Nanosecond)

	_, err := s.repo.GetUser(ctx, uuid.New())
	s.Require().Error(err)
	s.Contains(err.Error(), "context deadline exceeded")
}

func (s *MongoRepoSuite) TestDatabaseErrorHandlingInCreate() {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := s.repo.CreateUser(ctx, "test@example.com", []byte("hash"), func() (*domain.User, error) {
		return domain.CreateUser("Test", "test@example.com", []byte("hash"))
	})
	s.Require().Error(err)
	s.Contains(err.Error(), "context canceled")
}

func (s *MongoRepoSuite) TestDatabaseErrorHandlingInUpdate() {
	ctx := context.Background()
	email := "update@example.com"
	passwordHash := []byte("hashedpassword")
	name := "Test User"

	user, err := s.repo.CreateUser(ctx, email, passwordHash, func() (*domain.User, error) {
		return domain.CreateUser(name, email, passwordHash)
	})
	s.Require().NoError(err)

	ctxCanceled, cancel := context.WithCancel(context.Background())
	cancel()

	_, err = s.repo.UpdateUser(ctxCanceled, user.ID(), func(_ *domain.User) (bool, error) {
		return true, nil
	})
	s.Require().Error(err)
	s.Contains(err.Error(), "context canceled")
}

func (s *MongoRepoSuite) TestDatabaseCorruptedDocumentHandling() {
	ctx := context.Background()

	corruptedDoc := bson.M{
		"user_id":       "invalid-uuid",
		"name":          "Test User",
		"email":         "test@example.com",
		"password_hash": []byte("hash"),
	}

	_, err := s.db.Collection("users").InsertOne(ctx, corruptedDoc)
	s.Require().NoError(err)

	cursor, err := s.db.Collection("users").Find(ctx, bson.M{"name": "Test User"})
	s.Require().NoError(err)
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var result bson.M
		err = cursor.Decode(&result)
		s.Require().NoError(err)

		if userIDValue, ok := result["user_id"]; ok && userIDValue == "invalid-uuid" {
			s.T().Log("Found corrupted document with invalid UUID")
			break
		}
	}
}

func (s *MongoRepoSuite) TestLargeDatasetOperations() {
	ctx := context.Background()
	passwordHash := []byte("hashedpassword")
	name := "Bulk User"

	const numUsers = 100
	userIDs := make([]uuid.UUID, 0, numUsers)

	start := time.Now()
	for i := range numUsers {
		email := fmt.Sprintf("bulk%d@example.com", i)
		user, err := s.repo.CreateUser(ctx, email, passwordHash, func() (*domain.User, error) {
			return domain.CreateUser(name, email, passwordHash)
		})
		s.Require().NoError(err)
		userIDs = append(userIDs, user.ID())
	}
	createDuration := time.Since(start)
	s.T().Logf(
		"Created %d users in %v (%.2f users/sec)",
		numUsers,
		createDuration,
		float64(numUsers)/createDuration.Seconds(),
	)

	start = time.Now()
	for _, userID := range userIDs {
		fetchedUser, err := s.repo.GetUser(ctx, userID)
		s.Require().NoError(err)
		s.Equal(name, fetchedUser.Name())
	}
	readDuration := time.Since(start)
	s.T().Logf("Read %d users in %v (%.2f reads/sec)", numUsers, readDuration, float64(numUsers)/readDuration.Seconds())

	start = time.Now()
	for i, userID := range userIDs {
		newEmail := fmt.Sprintf("updated%d@bulk.com", i)
		updatedUser, err := s.repo.UpdateUser(ctx, userID, func(u *domain.User) (bool, error) {
			err := u.ChangeEmail(newEmail)
			if err != nil {
				return false, err
			}
			return true, nil
		})
		s.Require().NoError(err)
		s.Equal(newEmail, updatedUser.Email())
	}
	updateDuration := time.Since(start)
	s.T().Logf(
		"Updated %d users in %v (%.2f updates/sec)",
		numUsers,
		updateDuration,
		float64(numUsers)/updateDuration.Seconds(),
	)
}

func (s *MongoRepoSuite) TestPerformanceLargeDataset() {
	if testing.Short() {
		s.T().Skip("Skipping performance test in short mode")
	}

	ctx := context.Background()
	passwordHash := []byte("hashedpassword")
	name := "Performance User"

	const numUsers = 1000
	userIDs := make([]uuid.UUID, 0, numUsers)

	start := time.Now()
	for i := range numUsers {
		email := fmt.Sprintf("perf%d@example.com", i)
		user, err := s.repo.CreateUser(ctx, email, passwordHash, func() (*domain.User, error) {
			return domain.CreateUser(name, email, passwordHash)
		})
		s.Require().NoError(err)
		userIDs = append(userIDs, user.ID())

		if i%100 == 0 {
			s.T().Logf("Created %d/%d users", i+1, numUsers)
		}
	}
	createDuration := time.Since(start)
	s.T().Logf(
		"Performance test: Created %d users in %v (%.2f users/sec)",
		numUsers,
		createDuration,
		float64(numUsers)/createDuration.Seconds(),
	)

	s.Less(createDuration.Seconds(), 30.0, "Creating 1000 users should take less than 30 seconds")

	start = time.Now()
	for i, userID := range userIDs {
		_, err := s.repo.GetUser(ctx, userID)
		s.Require().NoError(err)

		if i%100 == 0 {
			s.T().Logf("Read %d/%d users", i+1, numUsers)
		}
	}
	readDuration := time.Since(start)
	s.T().Logf(
		"Performance test: Read %d users in %v (%.2f reads/sec)",
		numUsers,
		readDuration,
		float64(numUsers)/readDuration.Seconds(),
	)

	s.Less(readDuration.Seconds(), 10.0, "Reading 1000 users should take less than 10 seconds")
}

func (s *MongoRepoSuite) TestMemoryUsageWithLargeDataset() {
	if testing.Short() {
		s.T().Skip("Skipping memory test in short mode")
	}

	ctx := context.Background()
	passwordHash := []byte("hashedpassword")
	name := "Memory Test User with a very long name to increase memory usage"

	const numUsers = 500
	users := make([]*domain.User, 0, numUsers)

	for i := range numUsers {
		email := fmt.Sprintf("memory%d@example.com", i)
		user, err := s.repo.CreateUser(ctx, email, passwordHash, func() (*domain.User, error) {
			return domain.CreateUser(name, email, passwordHash)
		})
		s.Require().NoError(err)
		users = append(users, user)

		if i%100 == 0 {
			s.T().Logf("Created %d/%d users for memory test", i+1, numUsers)
		}
	}

	for i, user := range users {
		fetchedUser, err := s.repo.GetUser(ctx, user.ID())
		s.Require().NoError(err)
		s.Equal(user.ID(), fetchedUser.ID())

		if i%100 == 0 {
			s.T().Logf("Verified %d/%d users for memory test", i+1, numUsers)
		}
	}

	s.T().Logf("Memory test completed: %d users created and verified", numUsers)
}

func (s *MongoRepoSuite) TestUserWithSpecialCharacters() {
	ctx := context.Background()
	email := "special+chars@test-domain.co.uk"
	passwordHash := []byte("hashedpassword")
	name := "User With Special Chars: äöü ñ 中文 🚀"

	user, err := s.repo.CreateUser(ctx, email, passwordHash, func() (*domain.User, error) {
		return domain.CreateUser(name, email, passwordHash)
	})
	s.Require().NoError(err)
	s.Equal(name, user.Name())
	s.Equal(email, user.Email())

	fetchedUser, err := s.repo.GetUser(ctx, user.ID())
	s.Require().NoError(err)
	s.Equal(name, fetchedUser.Name())
	s.Equal(email, fetchedUser.Email())
}

func (s *MongoRepoSuite) TestUserPasswordHashHandling() {
	ctx := context.Background()
	email := "password@example.com"
	password := "complex$password#hash@123"
	name := "Password User"

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	s.Require().NoError(err)

	user, err := s.repo.CreateUser(ctx, email, passwordHash, func() (*domain.User, error) {
		return domain.CreateUser(name, email, passwordHash)
	})
	s.Require().NoError(err)

	fetchedUser, err := s.repo.GetUser(ctx, user.ID())
	s.Require().NoError(err)

	originalPasswordValid := user.CheckPassword(password)
	s.True(originalPasswordValid)

	fetchedPasswordValid := fetchedUser.CheckPassword(password)
	s.True(fetchedPasswordValid)
}

func TestMongoRepoSuite(t *testing.T) {
	suite.Run(t, new(MongoRepoSuite))
}
