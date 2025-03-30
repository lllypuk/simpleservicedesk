package models

import (
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestTodoModel(t *testing.T) {
	// Test creating a new todo
	id := primitive.NewObjectID()
	now := time.Now()

	todo := Todo{
		ID:        id,
		Title:     "Test task",
		Completed: false,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if todo.ID != id {
		t.Errorf("Expected ID %v, got %v", id, todo.ID)
	}

	if todo.Title != "Test task" {
		t.Errorf("Expected title 'Test task', got '%s'", todo.Title)
	}

	if todo.Completed != false {
		t.Error("Expected completed=false")
	}

	if !todo.CreatedAt.Equal(now) {
		t.Error("CreatedAt time mismatch")
	}

	if !todo.UpdatedAt.Equal(now) {
		t.Error("UpdatedAt time mismatch")
	}
}
