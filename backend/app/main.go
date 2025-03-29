package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

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

	// Serve static files from frontend directory
	fs := http.FileServer(http.Dir("../../frontend"))
	http.Handle("/", http.StripPrefix("/", fs))

	// Handle refresh by serving index.html
	http.HandleFunc("/index.html", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "../../frontend/index.html")
	})

	// API routes
	http.HandleFunc("/api/todos", handlers.GetTodos)
	http.HandleFunc("/api/todos/create", handlers.CreateTodo)
	http.HandleFunc("/api/todos/update", handlers.UpdateTodo)
	http.HandleFunc("/api/todos/delete", handlers.DeleteTodo)

	fmt.Println("Server listening on :8081")
	if err := http.ListenAndServe(":8081", nil); err != nil {
		fmt.Println("Error starting server:", err)
	}
}
