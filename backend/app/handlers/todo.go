package handlers

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/lllypuk/simpleservicedesk/backend/app/db"
	"github.com/lllypuk/simpleservicedesk/backend/app/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func CreateTodo(w http.ResponseWriter, r *http.Request) {
	title := r.FormValue("title")
	if title == "" {
		http.Error(w, "Title is required", http.StatusBadRequest)
		return
	}

	todo := models.Todo{
		Title:     title,
		Completed: false,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	collection := db.GetCollection("todos")
	result, err := collection.InsertOne(context.Background(), todo)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(result.InsertedID.(primitive.ObjectID).Hex()))
}

func GetTodos(w http.ResponseWriter, r *http.Request) {
	collection := db.GetCollection("todos")
	cursor, err := collection.Find(context.Background(), bson.M{})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer cursor.Close(context.Background())

	var todos []models.Todo
	if err := cursor.All(context.Background(), &todos); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(http.StatusOK)
	for _, todo := range todos {
		fmt.Fprintf(w, `
		<div class="todo-item">
			<input type="checkbox"
				   hx-post="/api/todos/update"
				   hx-vars="id:'%s', completed:this.checked"
				   hx-target="#todo-list"
				   hx-swap="innerHTML"
				   %s>
			<span class="todo-title %s">%s</span>
			<button hx-post="/api/todos/delete"
					hx-vars="id:'%s'"
					hx-target="#todo-list"
					hx-swap="innerHTML"
					hx-confirm="Are you sure you want to delete this todo?">
				Delete
			</button>
		</div>
		`,
			todo.ID.Hex(),
			func() string {
				if todo.Completed {
					return "checked"
				} else {
					return ""
				}
			}(),
			func() string {
				if todo.Completed {
					return "completed"
				} else {
					return ""
				}
			}(),
			todo.Title,
			todo.ID.Hex())
	}
}

func UpdateTodo(w http.ResponseWriter, r *http.Request) {
	id := r.FormValue("id")
	if id == "" {
		http.Error(w, "ID is required", http.StatusBadRequest)
		return
	}

	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		http.Error(w, "Invalid ID format", http.StatusBadRequest)
		return
	}

	completed := r.FormValue("completed")
	update := bson.M{"$set": bson.M{"updatedAt": time.Now()}}

	if completed != "" {
		// Checkbox toggle - only update completion status
		update["$set"].(bson.M)["completed"] = completed == "true"
	} else {
		// Edit operation - update title if provided
		if title := r.FormValue("title"); title != "" {
			update["$set"].(bson.M)["title"] = title
		}
	}

	collection := db.GetCollection("todos")
	_, err = collection.UpdateByID(context.Background(), objID, update)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return updated todo list
	GetTodos(w, r)
}

func DeleteTodo(w http.ResponseWriter, r *http.Request) {
	id := r.FormValue("id")
	if id == "" {
		http.Error(w, "ID is required", http.StatusBadRequest)
		return
	}

	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		http.Error(w, "Invalid ID format", http.StatusBadRequest)
		return
	}

	collection := db.GetCollection("todos")
	_, err = collection.DeleteOne(context.Background(), bson.M{"_id": objID})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
