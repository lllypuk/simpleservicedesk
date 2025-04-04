package handlers

import (
	"fmt"
	"log" // Keep for status codes, might remove later if Fiber provides constants
	"strings"

	"github.com/gofiber/fiber/v2" // Import Fiber
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
func (h *TodoHandler) CreateTodo(c *fiber.Ctx) error {
	// Fiber automatically parses the form, access directly
	title := strings.TrimSpace(c.FormValue("title"))
	if title == "" {
		return c.Status(fiber.StatusBadRequest).SendString("Title is required")
	}

	todo := models.Todo{
		Title:     title,
		Completed: false,
		// CreatedAt and UpdatedAt are set by the store
	}

	// Use Fiber context for request context
	createdTodo, err := h.store.Create(c.Context(), &todo)
	if err != nil {
		log.Printf("Error creating todo: %v", err)
		return c.Status(fiber.StatusInternalServerError).SendString("Failed to create todo")
	}

	// Return the HTML fragment for the new todo item for HTMX
	c.Set(fiber.HeaderContentType, fiber.MIMETextHTML) // Set content type
	return c.Status(fiber.StatusCreated).SendString(generateTodoItemHTML(createdTodo))
}

// GetTodos handles retrieving all Todo items and rendering them as HTML fragments.
func (h *TodoHandler) GetTodos(c *fiber.Ctx) error {
	todos, err := h.store.GetAll(c.Context())
	if err != nil {
		log.Printf("Error getting todos: %v", err)
		return c.Status(fiber.StatusInternalServerError).SendString("Failed to retrieve todos")
	}

	var htmlBuilder strings.Builder
	for _, todo := range todos {
		htmlBuilder.WriteString(generateTodoItemHTML(&todo))
	}

	c.Set(fiber.HeaderContentType, fiber.MIMETextHTML)
	return c.Status(fiber.StatusOK).SendString(htmlBuilder.String())
}

// UpdateTodo handles updating a Todo item's completion status.
// It expects the 'id' from the URL path and 'completed' form value.
// It returns an HTML fragment of the updated item.
func (h *TodoHandler) UpdateTodo(c *fiber.Ctx) error {
	// Get ID from URL parameter using Fiber's Params method
	idStr := c.Params("id")
	if idStr == "" {
		// Fiber's routing usually ensures this, but check defensively
		return c.Status(fiber.StatusBadRequest).SendString("ID is required in URL path")
	}

	objID, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Invalid ID format in URL path")
	}

	// Get 'completed' status from form value
	completedStr := c.FormValue("completed")
	// We expect 'completed' to be present for this handler based on the HTML
	if completedStr == "" {
		return c.Status(fiber.StatusBadRequest).SendString("'completed' form value is required")
	}

	// Fetch the existing todo to update its fields
	existingTodo, err := h.store.GetByID(c.Context(), objID)
	if err != nil {
		log.Printf("Error finding todo %s for update: %v", idStr, err)
		if strings.Contains(err.Error(), "not found") {
			return c.Status(fiber.StatusNotFound).SendString("Todo not found")
		}
		return c.Status(fiber.StatusInternalServerError).SendString("Failed to find todo for update")
	}

	// Update the completion status
	existingTodo.Completed = (completedStr == "true")
	// Title is not updated here

	// Update the todo in the store (this also updates UpdatedAt)
	updatedTodo, err := h.store.Update(c.Context(), objID, existingTodo)
	if err != nil {
		log.Printf("Error updating todo %s: %v", idStr, err)
		if strings.Contains(err.Error(), "not found") { // Check again in case of race condition
			return c.Status(fiber.StatusNotFound).SendString("Todo not found during update")
		}
		return c.Status(fiber.StatusInternalServerError).SendString("Failed to update todo")
	}

	// Return the HTML fragment for the updated todo item for HTMX swap
	c.Set(fiber.HeaderContentType, fiber.MIMETextHTML)
	return c.Status(fiber.StatusOK).SendString(generateTodoItemHTML(updatedTodo))
}

// DeleteTodo handles deleting a Todo item.
// It expects the 'id' from the URL path.
func (h *TodoHandler) DeleteTodo(c *fiber.Ctx) error {
	// Get ID from URL parameter
	idStr := c.Params("id")
	if idStr == "" {
		return c.Status(fiber.StatusBadRequest).SendString("ID is required in URL path")
	}

	objID, err := primitive.ObjectIDFromHex(idStr)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Invalid ID format in URL path")
	}

	err = h.store.Delete(c.Context(), objID)
	if err != nil {
		log.Printf("Error deleting todo %s: %v", idStr, err)
		if strings.Contains(err.Error(), "not found") {
			return c.Status(fiber.StatusNotFound).SendString("Todo not found for deletion")
		}
		return c.Status(fiber.StatusInternalServerError).SendString("Failed to delete todo")
	}

	// Return success with empty body. HTMX handles the swap.
	return c.SendStatus(fiber.StatusOK) // Send only status
}

// generateTodoItemHTML generates the HTML fragment string for a single todo item.
// Renamed from writeTodoItemHTML and returns a string instead of writing to a writer.
func generateTodoItemHTML(todo *models.Todo) string {
	checkedAttr := ""
	completedClass := ""
	if todo.Completed {
		checkedAttr = "checked"
		completedClass = "completed"
	}
	idHex := todo.ID.Hex()
	targetID := "todo-" + idHex

	// Use PUT for update, DELETE for delete, targeting the specific item div
	// Note: The hx-vals for completed now sends "true" or "false" as strings,
	// which the UpdateTodo handler expects.
	return fmt.Sprintf(`
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
