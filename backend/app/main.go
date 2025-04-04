package main

import (
	"context"
	"log"

	// "net/http" // No longer needed for server/routing
	"os"
	// "path/filepath" // No longer needed for FileServer helper
	// "strings" // No longer needed for FileServer helper

	"github.com/gofiber/fiber/v2"                    // Import Fiber
	"github.com/gofiber/fiber/v2/middleware/logger"  // Fiber logger middleware
	"github.com/gofiber/fiber/v2/middleware/recover" // Fiber recover middleware
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

	// Initialize Fiber app
	app := fiber.New()

	// Middleware
	app.Use(recover.New()) // Recovers from panics
	app.Use(logger.New())  // Logs requests (includes request ID)
	// Note: RealIP handling often depends on proxy setup (e.g., X-Forwarded-For)
	// Fiber's Ctx.IP() tries to get the real IP.

	// API routes group
	api := app.Group("/api")     // Group for API routes
	todos := api.Group("/todos") // Group for Todo routes

	todos.Get("/", todoHandler.GetTodos)         // GET /api/todos
	todos.Post("/", todoHandler.CreateTodo)      // POST /api/todos
	todos.Put("/:id", todoHandler.UpdateTodo)    // PUT /api/todos/:id (Fiber uses :param)
	todos.Delete("/:id", todoHandler.DeleteTodo) // DELETE /api/todos/:id (Fiber uses :param)

	// --- Static file serving ---
	// Define the path relative to the 'backend' directory
	frontendDir := "../frontend"
	log.Printf("Serving static files from: %s", frontendDir)

	// Ensure the path exists before trying to serve
	if _, err := os.Stat(frontendDir); os.IsNotExist(err) {
		log.Fatalf("Frontend directory not found at %s. Ensure you are running 'go run app/main.go' from the 'backend' directory.", frontendDir)
	}

	// Serve static files using Fiber's Static middleware
	// The first argument is the prefix, the second is the root directory.
	// "/" prefix means requests like /style.css will look for ../frontend/style.css
	// We also configure it to serve index.html for SPA-like behavior (requests to non-existent files serve index.html)
	app.Static("/", frontendDir, fiber.Static{
		Index: "index.html", // Serve index.html for "/"
		// NotFoundFile: "index.html", // Serve index.html if file not found (for SPA routing)
		Compress: true, // Enable compression
	})

	// --- Start Server ---
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081" // Default port
	}

	log.Printf("Server listening on :%s", port)
	// Use app.Listen to start the Fiber server
	if err := app.Listen(":" + port); err != nil {
		log.Fatalf("Error starting Fiber server: %v", err)
	}
}

// FileServer helper function is no longer needed as Fiber's app.Static handles this.
// Removed the function definition.
