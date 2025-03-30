package db

import (
	"context"

	"github.com/lllypuk/simpleservicedesk/backend/app/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// TodoStore defines the interface for database operations on Todo items.
type TodoStore interface {
	// Create inserts a new Todo item into the store.
	// It returns the created Todo with its assigned ID or an error.
	Create(ctx context.Context, todo *models.Todo) (*models.Todo, error)

	// GetAll retrieves all Todo items from the store.
	// It returns a slice of Todos or an error.
	GetAll(ctx context.Context) ([]models.Todo, error)

	// GetByID retrieves a single Todo item by its ID.
	// It returns the found Todo or an error (e.g., if not found).
	GetByID(ctx context.Context, id primitive.ObjectID) (*models.Todo, error)

	// Update modifies an existing Todo item identified by its ID.
	// It accepts the updated Todo data and returns the modified Todo or an error.
	Update(ctx context.Context, id primitive.ObjectID, todo *models.Todo) (*models.Todo, error)

	// Delete removes a Todo item identified by its ID from the store.
	// It returns an error if the operation fails.
	Delete(ctx context.Context, id primitive.ObjectID) error
}
