package handlers

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5" // Import chi
	"github.com/lllypuk/simpleservicedesk/backend/app/db"
	"github.com/lllypuk/simpleservicedesk/backend/app/models"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// TodoHandler holds the dependencies for Todo handlers, like the data store.
type TodoHandler struct {
	store db.TodoStore
}

// NewTodoHandler creates a new TodoHandler with the given TodoStore.
func NewTodoHandler(store db.TodoStore) *TodoHandler {
	return &TodoHandler{store: store}
}

// CreateTodo handles the creation of a new Todo item.
// It expects a 'title' form value and returns an HTML fragment of the new item.
func (h *TodoHandler) CreateTodo(w http.ResponseWriter, r *http.Request) {
	// It's good practice to parse the form first
	if err := r.ParseForm(); err != nil {
		log.Printf("Error parsing form: %v", err)
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	title := strings.TrimSpace(r.FormValue("title"))
	if title == "" {
		http.Error(w, "Title is required", http.StatusBadRequest)
		return
	}

	todo := models.Todo{
		Title:     title,
		Completed: false,
		// CreatedAt and UpdatedAt are set by the store
	}

	createdTodo, err := h.store.Create(r.Context(), &todo)
	if err != nil {
		log.Printf("Error creating todo: %v", err)
		http.Error(w, "Failed to create todo", http.StatusInternalServerError)
		return
	}

	// Return the HTML fragment for the new todo item for HTMX
	w.WriteHeader(http.StatusCreated)
	writeTodoItemHTML(w, createdTodo) // Use helper to generate HTML
}

// GetTodos handles retrieving all Todo items and rendering them as HTML fragments.
func (h *TodoHandler) GetTodos(w http.ResponseWriter, r *http.Request) {
	todos, err := h.store.GetAll(r.Context())
	if err != nil {
		log.Printf("Error getting todos: %v", err)
		http.Error(w, "Failed to retrieve todos", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	for _, todo := range todos {
		writeTodoItemHTML(w, &todo) // Use helper
	}
}

// UpdateTodo handles updating a Todo item's completion status.
// It expects the 'id' from the URL path and 'completed' form value.
// It returns an HTML fragment of the updated item.
func (h *TodoHandler) UpdateTodo(w http.ResponseWriter, r *http.Request) {
	// Get ID from URL parameter
	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		// This shouldn't happen with chi routing, but check anyway
		http.Error(w, "ID is required in URL path", http.StatusBadRequest)
		return
	}

	objID, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		http.Error(w, "Invalid ID format in URL path", http.StatusBadRequest)
		return
	}

	// Parse form to get the 'completed' status
	if err := r.ParseForm(); err != nil {
		log.Printf("Error parsing form for update: %v", err)
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}
	completedStr := r.FormValue("completed")
	// We expect 'completed' to be present for this handler based on the HTML
	if completedStr == "" {
		http.Error(w, "'completed' form value is required", http.StatusBadRequest)
		return
	}

	// Fetch the existing todo to update its fields
	existingTodo, err := h.store.GetByID(r.Context(), objID)
	if err != nil {
		log.Printf("Error finding todo %s for update: %v", idStr, err)
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Todo not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to find todo for update", http.StatusInternalServerError)
		}
		return
	}

	// Update the completion status
	existingTodo.Completed = (completedStr == "true")
	// Title is not updated here, assuming separate mechanism if needed

	// Update the todo in the store (this also updates UpdatedAt)
	updatedTodo, err := h.store.Update(r.Context(), objID, existingTodo)
	if err != nil {
		log.Printf("Error updating todo %s: %v", idStr, err)
		if strings.Contains(err.Error(), "not found") { // Check again in case of race condition
			http.Error(w, "Todo not found during update", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to update todo", http.StatusInternalServerError)
		}
		return
	}

	// Return the HTML fragment for the updated todo item for HTMX swap
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	writeTodoItemHTML(w, updatedTodo)
}

// DeleteTodo handles deleting a Todo item.
// It expects the 'id' from the URL path.
func (h *TodoHandler) DeleteTodo(w http.ResponseWriter, r *http.Request) {
	// Get ID from URL parameter
	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		http.Error(w, "ID is required in URL path", http.StatusBadRequest)
		return
	}

	objID, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		http.Error(w, "Invalid ID format in URL path", http.StatusBadRequest)
		return
	}

	err = h.store.Delete(r.Context(), objID)
	if err != nil {
		log.Printf("Error deleting todo %s: %v", idStr, err)
		if strings.Contains(err.Error(), "not found") {
			http.Error(w, "Todo not found for deletion", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to delete todo", http.StatusInternalServerError)
		}
		return
	}

	// Return success with empty body. HTMX handles the swap.
	w.WriteHeader(http.StatusOK)
}

// writeTodoItemHTML is a helper function to write a single todo item as an HTML fragment.
func writeTodoItemHTML(w http.ResponseWriter, todo *models.Todo) {
	checkedAttr := ""
	completedClass := ""
	if todo.Completed {
		checkedAttr = "checked"
		completedClass = "completed"
	}
	idHex := todo.ID.Hex()
	targetID := "todo-" + idHex

	// Use PUT for update, DELETE for delete, targeting the specific item div
	fmt.Fprintf(w, `
<div class="todo-item" id="%s">
	<input type="checkbox"
		   hx-put="/api/todos/%s"
		   hx-vals='{"completed": "%t"}'
		   hx-target="#%s"
		   hx-swap="outerHTML"
		   %s>
	<span class="todo-title %s">%s</span>
	<button hx-delete="/api/todos/%s"
			hx-target="#%s"
			hx-swap="outerHTML"
			hx-confirm="Are you sure?">
		Delete
	</button>
</div>
`,
		targetID,
		idHex,           // ID in URL for PUT
		!todo.Completed, // Send the opposite value for the toggle in the body
		targetID,        // Target the item itself for replacement
		checkedAttr,
		completedClass,
		todo.Title,
		idHex,    // ID in URL for DELETE
		targetID, // Target the item itself for removal
	)
}
