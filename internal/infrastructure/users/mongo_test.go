package users

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	domain "simpleservicedesk/internal/domain/users"
)

type MongoRepoSuite struct {
	suite.Suite
	container testcontainers.Container
	db        *mongo.Database
	repo      *MongoRepo
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

	uri := fmt.Sprintf("mongodb://%s:%s", host, port.Port())
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	s.Require().NoError(err)

	s.db = client.Database("testdb")
	s.repo = NewMongoRepo(s.db)
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
		err := u.ChangeEmail(newEmail)
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

func TestMongoRepoSuite(t *testing.T) {
	suite.Run(t, new(MongoRepoSuite))
}
