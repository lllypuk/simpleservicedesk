package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/lllypuk/simpleservicedesk/backend/app/db"
	"github.com/lllypuk/simpleservicedesk/backend/app/handlers"
)

func main() {
	// Initialize MongoDB
	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		mongoURI = "mongodb://localhost:27017"
	}

	err := db.ConnectDB(mongoURI)
	if err != nil {
		log.Fatal("Failed to connect to MongoDB:", err)
	}
	defer db.Client.Disconnect(context.Background())

	// Get absolute path to frontend directory
	frontendPath := filepath.Join("..", "frontend")

	// Serve static files
	fs := http.FileServer(http.Dir(frontendPath))
	http.Handle("/", fs)

	// API routes
	http.HandleFunc("/api/todos", handlers.GetTodos)
	http.HandleFunc("/api/todos/create", handlers.CreateTodo)

	fmt.Println("Server listening on :8081")
	if err := http.ListenAndServe(":8081", nil); err != nil {
		fmt.Println("Error starting server:", err)
	}
}
