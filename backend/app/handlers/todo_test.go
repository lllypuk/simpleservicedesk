package handlers

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/lllypuk/simpleservicedesk/backend/app/db"
	"github.com/lllypuk/simpleservicedesk/backend/app/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Helper to create a chi context with URL parameters
func chiCtx(params map[string]string) context.Context {
	rctx := chi.NewRouteContext()
	for k, v := range params {
		rctx.URLParams.Add(k, v)
	}
	return context.WithValue(context.Background(), chi.RouteCtxKey, rctx)
}

func TestTodoHandlers(t *testing.T) {
	mockStore := &db.MockTodoStore{}
	handler := NewTodoHandler(mockStore)

	// --- Test CreateTodo ---
	t.Run("CreateTodo_Success", func(t *testing.T) {
		mockID := primitive.NewObjectID()
		mockTodo := &models.Todo{
			ID:        mockID,
			Title:     "Test Create",
			Completed: false,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		mockStore.CreateFn = func(ctx context.Context, todo *models.Todo) (*models.Todo, error) {
			// Basic check on input
			if todo.Title != "Test Create" {
				t.Errorf("Expected title 'Test Create', got '%s'", todo.Title)
			}
			// Return the mock todo with ID and timestamps
			return mockTodo, nil
		}

		formData := url.Values{}
		formData.Set("title", " Test Create ") // Include whitespace to test trimming
		req := httptest.NewRequest("POST", "/api/todos", strings.NewReader(formData.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()

		handler.CreateTodo(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("Expected status %d, got %d", http.StatusCreated, w.Code)
		}
		body := w.Body.String()
		if !strings.Contains(body, `id="todo-`+mockID.Hex()+`"`) {
			t.Errorf("Expected response body to contain todo item HTML with ID %s, got: %s", mockID.Hex(), body)
		}
		if !strings.Contains(body, `Test Create`) {
			t.Errorf("Expected response body to contain todo title 'Test Create', got: %s", body)
		}
	})

	t.Run("CreateTodo_MissingTitle", func(t *testing.T) {
		formData := url.Values{}
		formData.Set("title", "")
		req := httptest.NewRequest("POST", "/api/todos", strings.NewReader(formData.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()

		handler.CreateTodo(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d for missing title, got %d", http.StatusBadRequest, w.Code)
		}
	})

	t.Run("CreateTodo_StoreError", func(t *testing.T) {
		mockStore.CreateFn = func(ctx context.Context, todo *models.Todo) (*models.Todo, error) {
			return nil, errors.New("database error")
		}

		formData := url.Values{}
		formData.Set("title", "Test Error")
		req := httptest.NewRequest("POST", "/api/todos", strings.NewReader(formData.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()

		handler.CreateTodo(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("Expected status %d for store error, got %d", http.StatusInternalServerError, w.Code)
		}
	})

	// --- Test GetTodos ---
	t.Run("GetTodos_Success", func(t *testing.T) {
		mockID1 := primitive.NewObjectID()
		mockID2 := primitive.NewObjectID()
		mockTodos := []models.Todo{
			{ID: mockID1, Title: "Todo 1", Completed: false},
			{ID: mockID2, Title: "Todo 2", Completed: true},
		}
		mockStore.GetAllFn = func(ctx context.Context) ([]models.Todo, error) {
			return mockTodos, nil
		}

		req := httptest.NewRequest("GET", "/api/todos", nil)
		w := httptest.NewRecorder()

		handler.GetTodos(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}
		body := w.Body.String()
		if !strings.Contains(body, `id="todo-`+mockID1.Hex()+`"`) || !strings.Contains(body, `Todo 1`) {
			t.Errorf("Expected response body to contain HTML for Todo 1, got: %s", body)
		}
		if !strings.Contains(body, `id="todo-`+mockID2.Hex()+`"`) || !strings.Contains(body, `Todo 2`) || !strings.Contains(body, `checked`) {
			t.Errorf("Expected response body to contain HTML for Todo 2 (checked), got: %s", body)
		}
	})

	t.Run("GetTodos_StoreError", func(t *testing.T) {
		mockStore.GetAllFn = func(ctx context.Context) ([]models.Todo, error) {
			return nil, errors.New("database error")
		}

		req := httptest.NewRequest("GET", "/api/todos", nil)
		w := httptest.NewRecorder()

		handler.GetTodos(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("Expected status %d for store error, got %d", http.StatusInternalServerError, w.Code)
		}
	})

	// --- Test UpdateTodo ---
	t.Run("UpdateTodo_Success", func(t *testing.T) {
		mockID := primitive.NewObjectID()
		existingTodo := &models.Todo{ID: mockID, Title: "Update Me", Completed: false}
		updatedTodo := &models.Todo{ID: mockID, Title: "Update Me", Completed: true, UpdatedAt: time.Now()} // Store updates timestamp

		mockStore.GetByIDFn = func(ctx context.Context, id primitive.ObjectID) (*models.Todo, error) {
			if id != mockID {
				t.Errorf("GetByIDFn called with wrong ID: %s", id.Hex())
			}
			return existingTodo, nil
		}
		mockStore.UpdateFn = func(ctx context.Context, id primitive.ObjectID, todo *models.Todo) (*models.Todo, error) {
			if id != mockID {
				t.Errorf("UpdateFn called with wrong ID: %s", id.Hex())
			}
			if !todo.Completed { // Check if completed status was updated before passing to store
				t.Errorf("Expected todo.Completed to be true in UpdateFn")
			}
			return updatedTodo, nil
		}

		formData := url.Values{}
		formData.Set("completed", "true")
		req := httptest.NewRequest("PUT", "/api/todos/"+mockID.Hex(), strings.NewReader(formData.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		// Add chi context with URL param
		req = req.WithContext(chiCtx(map[string]string{"id": mockID.Hex()}))
		w := httptest.NewRecorder()

		handler.UpdateTodo(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}
		body := w.Body.String()
		if !strings.Contains(body, `id="todo-`+mockID.Hex()+`"`) || !strings.Contains(body, `checked`) {
			t.Errorf("Expected response body to contain updated todo item HTML (checked), got: %s", body)
		}
	})

	t.Run("UpdateTodo_NotFound", func(t *testing.T) {
		mockID := primitive.NewObjectID()
		mockStore.GetByIDFn = func(ctx context.Context, id primitive.ObjectID) (*models.Todo, error) {
			return nil, errors.New("not found") // Simulate store not finding the item
		}

		formData := url.Values{}
		formData.Set("completed", "true")
		req := httptest.NewRequest("PUT", "/api/todos/"+mockID.Hex(), strings.NewReader(formData.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req = req.WithContext(chiCtx(map[string]string{"id": mockID.Hex()}))
		w := httptest.NewRecorder()

		handler.UpdateTodo(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status %d for not found, got %d", http.StatusNotFound, w.Code)
		}
	})

	t.Run("UpdateTodo_InvalidID", func(t *testing.T) {
		req := httptest.NewRequest("PUT", "/api/todos/invalid-id", nil)
		req = req.WithContext(chiCtx(map[string]string{"id": "invalid-id"}))
		w := httptest.NewRecorder()
		handler.UpdateTodo(w, req)
		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d for invalid ID, got %d", http.StatusBadRequest, w.Code)
		}
	})

	// --- Test DeleteTodo ---
	t.Run("DeleteTodo_Success", func(t *testing.T) {
		mockID := primitive.NewObjectID()
		mockStore.DeleteFn = func(ctx context.Context, id primitive.ObjectID) error {
			if id != mockID {
				t.Errorf("DeleteFn called with wrong ID: %s", id.Hex())
			}
			return nil // Success
		}

		req := httptest.NewRequest("DELETE", "/api/todos/"+mockID.Hex(), nil)
		req = req.WithContext(chiCtx(map[string]string{"id": mockID.Hex()}))
		w := httptest.NewRecorder()

		handler.DeleteTodo(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
		}
		if w.Body.String() != "" {
			t.Errorf("Expected empty body on successful delete, got: %s", w.Body.String())
		}
	})

	t.Run("DeleteTodo_NotFound", func(t *testing.T) {
		mockID := primitive.NewObjectID()
		mockStore.DeleteFn = func(ctx context.Context, id primitive.ObjectID) error {
			return errors.New("not found") // Simulate store error
		}

		req := httptest.NewRequest("DELETE", "/api/todos/"+mockID.Hex(), nil)
		req = req.WithContext(chiCtx(map[string]string{"id": mockID.Hex()}))
		w := httptest.NewRecorder()

		handler.DeleteTodo(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected status %d for not found, got %d", http.StatusNotFound, w.Code)
		}
	})

	t.Run("DeleteTodo_InvalidID", func(t *testing.T) {
		req := httptest.NewRequest("DELETE", "/api/todos/invalid-id", nil)
		req = req.WithContext(chiCtx(map[string]string{"id": "invalid-id"}))
		w := httptest.NewRecorder()
		handler.DeleteTodo(w, req)
		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status %d for invalid ID, got %d", http.StatusBadRequest, w.Code)
		}
	})
}
