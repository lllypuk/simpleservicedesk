package application

import (
	"context"
	"github.com/google/uuid"
	"simpleservicedesk/internal/domain/users"
)

type UserRepository interface {
	CreateUser(ctx context.Context, email string, createFn func() (*users.User, error)) (*users.User, error)
	UpdateUser(ctx context.Context, id uuid.UUID, updateFn func(*users.User) (bool, error)) (*users.User, error)
	GetUser(ctx context.Context, id uuid.UUID) (*users.User, error)
}
