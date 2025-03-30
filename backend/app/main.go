package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/lllypuk/simpleservicedesk/backend/app/db"
	"github.com/lllypuk/simpleservicedesk/backend/app/handlers"
)

func main() {
	// Initialize MongoDB
	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		mongoURI = "mongodb://localhost:27017"
		log.Println("MONGO_URI not set, using default:", mongoURI)
	}

	// Connect to DB
	client, err := db.ConnectDB(mongoURI)
	if err != nil {
		log.Fatal("Failed to connect to MongoDB:", err)
	}
	// Ensure client disconnects when main function exits
	defer func() {
		if err := client.Disconnect(context.Background()); err != nil {
			log.Printf("Error disconnecting from MongoDB: %v", err)
		} else {
			log.Println("Disconnected from MongoDB.")
		}
	}()
	log.Println("Connected to MongoDB.")

	// Create data store and handler
	todoStore := db.NewMongoTodoStore(client)
	todoHandler := handlers.NewTodoHandler(todoStore)

	// Initialize Chi router
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)    // Logs requests
	r.Use(middleware.Recoverer) // Recovers from panics

	// API routes
	r.Route("/api/todos", func(r chi.Router) {
		r.Get("/", todoHandler.GetTodos)    // GET /api/todos
		r.Post("/", todoHandler.CreateTodo) // POST /api/todos (Changed from /create)
		// Note: Update and Delete handlers need adjustment for URL param ID
		r.Put("/{id}", todoHandler.UpdateTodo)    // PUT /api/todos/{id} (Changed from /update)
		r.Delete("/{id}", todoHandler.DeleteTodo) // DELETE /api/todos/{id} (Changed from /delete)
	})

	// --- Static file serving ---
	// Define the path relative to the 'backend' directory where 'go run' is executed
	frontendDir := "../frontend"
	log.Printf("Serving static files from: %s", frontendDir)

	// Serve static files using chi's FileServer helper
	// Ensure the path exists before trying to serve
	if _, err := os.Stat(frontendDir); os.IsNotExist(err) {
		log.Fatalf("Frontend directory not found at %s. Ensure you are running 'go run app/main.go' from the 'backend' directory.", frontendDir)
	}
	FileServer(r, "/", http.Dir(frontendDir))

	// --- Start Server ---
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081" // Default port
	}

	log.Printf("Server listening on :%s", port)
	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}

// FileServer conveniently sets up a static file server that serves files
// including index.html from the root and handles requests for paths that
// don't exist by serving index.html (useful for SPAs).
func FileServer(r chi.Router, public string, static http.FileSystem) {
	// Ensure the path is absolute or relative to the current working directory
	if strings.ContainsAny(public, "{}*") {
		panic("FileServer does not permit URL parameters.")
	}

	root, _ := static.Open("/")
	if root == nil {
		panic("Static file directory does not exist or is not accessible.")
	}
	root.Close() // We just checked existence

	fs := http.StripPrefix(public, http.FileServer(static))

	if public != "/" && public[len(public)-1] != '/' {
		r.Get(public, http.RedirectHandler(public+"/", http.StatusMovedPermanently).ServeHTTP)
		public += "/"
	}
	r.Get(public+"*", func(w http.ResponseWriter, r *http.Request) {
		// Check if the requested file exists
		file := strings.TrimPrefix(r.URL.Path, public)
		f, err := static.Open(file)
		if os.IsNotExist(err) {
			// File doesn't exist, serve index.html
			http.ServeFile(w, r, filepath.Join(public, "index.html")) // Assuming index.html is in the root of static dir
			return
		} else if err != nil {
			// Other error (e.g., permission denied)
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
		f.Close() // Close the file after checking existence

		// File exists, serve it
		fs.ServeHTTP(w, r)
	})
}
