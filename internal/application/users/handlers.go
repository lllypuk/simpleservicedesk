package users

import (
	"context"

	"simpleservicedesk/internal/domain/users"
	"simpleservicedesk/internal/queries"

	"github.com/google/uuid"
)

type Repository interface {
	CreateUser(ctx context.Context,
		email string,
		passwordHash []byte,
		createFn func() (*users.User, error)) (*users.User, error)
	UpdateUser(ctx context.Context, id uuid.UUID, updateFn func(*users.User) (bool, error)) (*users.User, error)
	GetUser(ctx context.Context, id uuid.UUID) (*users.User, error)
	ListUsers(ctx context.Context, filter queries.UserFilter) ([]*users.User, error)
	DeleteUser(ctx context.Context, id uuid.UUID) error
	CountUsers(ctx context.Context, filter queries.UserFilter) (int64, error)
}

type UserHandlers struct {
	repo Repository
}

func SetupHandlers(repo Repository) UserHandlers {
	return UserHandlers{
		repo: repo,
	}
}
