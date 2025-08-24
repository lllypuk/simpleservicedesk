package users

import (
	"context"

	"simpleservicedesk/internal/domain/users"
	infraUsers "simpleservicedesk/internal/infrastructure/users"

	"github.com/google/uuid"
)

type Repository interface {
	CreateUser(ctx context.Context,
		email string,
		passwordHash []byte,
		createFn func() (*users.User, error)) (*users.User, error)
	UpdateUser(ctx context.Context, id uuid.UUID, updateFn func(*users.User) (bool, error)) (*users.User, error)
	GetUser(ctx context.Context, id uuid.UUID) (*users.User, error)
	ListUsers(ctx context.Context, filter infraUsers.UserFilter) ([]*users.User, error)
	DeleteUser(ctx context.Context, id uuid.UUID) error
	CountUsers(ctx context.Context, filter infraUsers.UserFilter) (int64, error)
}

type UserHandlers struct {
	repo Repository
}

func SetupHandlers(repo Repository) UserHandlers {
	return UserHandlers{
		repo: repo,
	}
}
