package db

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/lllypuk/simpleservicedesk/backend/app/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	databaseName   = "todos"
	collectionName = "todos"
)

// MongoTodoStore implements the TodoStore interface using MongoDB.
type MongoTodoStore struct {
	collection *mongo.Collection
}

// NewMongoTodoStore creates a new MongoTodoStore instance.
func NewMongoTodoStore(client *mongo.Client) *MongoTodoStore {
	collection := client.Database(databaseName).Collection(collectionName)
	return &MongoTodoStore{collection: collection}
}

// ConnectDB establishes a connection to MongoDB.
// It returns the mongo client instance or an error.
func ConnectDB(uri string) (*mongo.Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		// Disconnect if ping fails
		_ = client.Disconnect(context.Background())
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	return client, nil
}

// --- TodoStore Interface Implementation ---

// Create inserts a new Todo item into the store.
func (s *MongoTodoStore) Create(ctx context.Context, todo *models.Todo) (*models.Todo, error) {
	now := time.Now()
	todo.ID = primitive.NewObjectID() // Generate new ID
	todo.CreatedAt = now
	todo.UpdatedAt = now

	_, err := s.collection.InsertOne(ctx, todo)
	if err != nil {
		return nil, fmt.Errorf("failed to insert todo: %w", err)
	}
	return todo, nil
}

// GetAll retrieves all Todo items from the store.
func (s *MongoTodoStore) GetAll(ctx context.Context) ([]models.Todo, error) {
	cursor, err := s.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, fmt.Errorf("failed to find todos: %w", err)
	}
	defer cursor.Close(ctx)

	var todos []models.Todo
	if err = cursor.All(ctx, &todos); err != nil {
		return nil, fmt.Errorf("failed to decode todos: %w", err)
	}
	return todos, nil
}

// GetByID retrieves a single Todo item by its ID.
func (s *MongoTodoStore) GetByID(ctx context.Context, id primitive.ObjectID) (*models.Todo, error) {
	var todo models.Todo
	err := s.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&todo)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, fmt.Errorf("todo with ID %s not found", id.Hex()) // Consider a specific error type
		}
		return nil, fmt.Errorf("failed to find todo by ID %s: %w", id.Hex(), err)
	}
	return &todo, nil
}

// Update modifies an existing Todo item identified by its ID.
func (s *MongoTodoStore) Update(ctx context.Context, id primitive.ObjectID, todo *models.Todo) (*models.Todo, error) {
	now := time.Now()
	update := bson.M{
		"$set": bson.M{
			"title":     todo.Title,
			"completed": todo.Completed,
			"updatedAt": now,
		},
	}
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)

	var updatedTodo models.Todo
	err := s.collection.FindOneAndUpdate(ctx, bson.M{"_id": id}, update, opts).Decode(&updatedTodo)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, fmt.Errorf("todo with ID %s not found for update", id.Hex()) // Consider a specific error type
		}
		return nil, fmt.Errorf("failed to update todo with ID %s: %w", id.Hex(), err)
	}
	return &updatedTodo, nil
}

// Delete removes a Todo item identified by its ID from the store.
func (s *MongoTodoStore) Delete(ctx context.Context, id primitive.ObjectID) error {
	result, err := s.collection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return fmt.Errorf("failed to delete todo with ID %s: %w", id.Hex(), err)
	}
	if result.DeletedCount == 0 {
		return fmt.Errorf("todo with ID %s not found for deletion", id.Hex()) // Consider a specific error type
	}
	return nil
}
