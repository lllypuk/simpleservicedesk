package handlers

import (
	"context"
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

	// TODO: Implement proper JSON response
	w.WriteHeader(http.StatusOK)
	for _, todo := range todos {
		w.Write([]byte(todo.Title + "\n"))
	}
}
