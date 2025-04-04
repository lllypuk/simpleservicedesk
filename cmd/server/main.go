package main

import (
	"context"
	"log"
	"os"
	"time"

	"simpleservicedesk/internal/auth"
	"simpleservicedesk/internal/database"
	"simpleservicedesk/internal/handlers" // Added
	"simpleservicedesk/internal/models"
	"simpleservicedesk/internal/repository" // Added

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/html/v2" // Added
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func main() {
	// Connect to Database
	if err := database.ConnectDB(); err != nil {
		log.Fatalf("Could not connect to the database: %v", err)
	}
	// Ensure unique index for usernames
	ctxIdx, cancelIdx := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelIdx()
	if err := auth.EnsureUsernameIndex(ctxIdx); err != nil {
		log.Fatalf("Could not ensure username index: %v", err)
	}

	// Create default admin user if none exists (for development)
	createDefaultAdminUser()

	// Disconnect from database on shutdown
	defer func() {
		ctxDisconnect, cancelDisconnect := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancelDisconnect()
		if err := database.DisconnectDB(ctxDisconnect); err != nil {
			log.Printf("Error disconnecting from database: %v", err)
		}
	}()

	// --- Setup Repositories ---
	requestRepo := repository.NewMongoRequestRepository()

	// --- Setup Handlers ---
	requestHandler := handlers.NewRequestHandler(requestRepo)

	// --- Setup Fiber App with HTML Template Engine ---
	engine := html.New("./views", ".html")
	// Add reload capability for development (optional)
	// engine.Reload(true)
	app := fiber.New(fiber.Config{
		Views: engine,
	})

	// Public route
	app.Get("/", func(c *fiber.Ctx) error {
		// Render index template
		return c.Render("index", fiber.Map{
			"Title": "Simple Service Desk",
		})
	})

	// API group (Authentication Disabled)
	api := app.Group("/api") // Removed auth.BasicAuthMiddleware()

	// Example protected route
	api.Get("/me", func(c *fiber.Ctx) error {
		user, ok := c.Locals("user").(*models.User) // Get user from context
		if !ok {
			// This should ideally not happen if middleware is correct
			return c.SendStatus(fiber.StatusInternalServerError)
		}
		return c.JSON(fiber.Map{
			"message":  "Authenticated!",
			"username": user.Username,
			"userId":   user.ID,
		})
	})

	// --- Request Routes (Now Public) ---
	api.Get("/requests", requestHandler.GetRequests)
	api.Post("/requests", requestHandler.CreateRequest)
	api.Get("/requests/:id", requestHandler.GetRequest)
	api.Put("/requests/:id", requestHandler.UpdateRequest) // Using PUT for updates

	log.Println("Starting server on port 3000...")
	log.Fatal(app.Listen(":3000"))
}

// createDefaultAdminUser checks if any user exists and creates an admin user if not.
func createDefaultAdminUser() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	collection := database.DB.Collection("users")
	count, err := collection.CountDocuments(ctx, bson.M{})
	if err != nil && err != mongo.ErrNoDocuments {
		log.Printf("Error checking for existing users: %v", err)
		return // Don't stop startup, but log the error
	}

	if count == 0 {
		log.Println("No users found. Creating default admin user (admin/password)...")
		// Use environment variables or config in a real app
		adminUser := os.Getenv("ADMIN_USER")
		adminPass := os.Getenv("ADMIN_PASS")
		if adminUser == "" {
			adminUser = "admin"
		}
		if adminPass == "" {
			adminPass = "password"
		}

		_, err := auth.CreateUser(ctx, adminUser, adminPass)
		if err != nil {
			log.Printf("Failed to create default admin user: %v", err)
		} else {
			log.Println("Default admin user created successfully.")
		}
	} else {
		log.Printf("%d user(s) found in the database.", count)
	}
}
