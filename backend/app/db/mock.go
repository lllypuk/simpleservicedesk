package db

import (
	"context"

	"github.com/lllypuk/simpleservicedesk/backend/app/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// MockTodoStore is a mock implementation of the TodoStore interface for testing.
type MockTodoStore struct {
	CreateFn  func(ctx context.Context, todo *models.Todo) (*models.Todo, error)
	GetAllFn  func(ctx context.Context) ([]models.Todo, error)
	GetByIDFn func(ctx context.Context, id primitive.ObjectID) (*models.Todo, error)
	UpdateFn  func(ctx context.Context, id primitive.ObjectID, todo *models.Todo) (*models.Todo, error)
	DeleteFn  func(ctx context.Context, id primitive.ObjectID) error
}

// Ensure MockTodoStore implements TodoStore interface at compile time.
var _ TodoStore = (*MockTodoStore)(nil)

func (m *MockTodoStore) Create(ctx context.Context, todo *models.Todo) (*models.Todo, error) {
	if m.CreateFn == nil {
		panic("MockTodoStore: CreateFn is not set")
	}
	return m.CreateFn(ctx, todo)
}

func (m *MockTodoStore) GetAll(ctx context.Context) ([]models.Todo, error) {
	if m.GetAllFn == nil {
		panic("MockTodoStore: GetAllFn is not set")
	}
	return m.GetAllFn(ctx)
}

func (m *MockTodoStore) GetByID(ctx context.Context, id primitive.ObjectID) (*models.Todo, error) {
	if m.GetByIDFn == nil {
		panic("MockTodoStore: GetByIDFn is not set")
	}
	return m.GetByIDFn(ctx, id)
}

func (m *MockTodoStore) Update(ctx context.Context, id primitive.ObjectID, todo *models.Todo) (*models.Todo, error) {
	if m.UpdateFn == nil {
		panic("MockTodoStore: UpdateFn is not set")
	}
	return m.UpdateFn(ctx, id, todo)
}

func (m *MockTodoStore) Delete(ctx context.Context, id primitive.ObjectID) error {
	if m.DeleteFn == nil {
		panic("MockTodoStore: DeleteFn is not set")
	}
	return m.DeleteFn(ctx, id)
}
