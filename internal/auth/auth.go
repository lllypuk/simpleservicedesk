package auth

import (
	"context"
	"encoding/base64"
	"errors"
	"log"
	"strings"
	"time"

	"simpleservicedesk/internal/database"
	"simpleservicedesk/internal/models"

	"github.com/gofiber/fiber/v2" // Changed to v2
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
)

// HashPassword generates a bcrypt hash of the password
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// CheckPasswordHash compares a bcrypt hashed password with its possible plaintext equivalent
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// CreateUser creates a new user in the database
func CreateUser(ctx context.Context, username, password string) (*models.User, error) {
	hashedPassword, err := HashPassword(password)
	if err != nil {
		log.Printf("Error hashing password: %v", err)
		return nil, err
	}

	user := &models.User{
		Username: username,
		Password: hashedPassword,
	}

	collection := database.DB.Collection("users")
	_, err = collection.InsertOne(ctx, user)
	if err != nil {
		// Handle potential duplicate username error (index needed on username field)
		if mongo.IsDuplicateKeyError(err) {
			return nil, errors.New("username already exists")
		}
		log.Printf("Error creating user %s: %v", username, err)
		return nil, err
	}

	// Ideally, return the user object with the ID set by InsertOne,
	// but InsertOneResult doesn't directly return the full object easily.
	// We can fetch it again or just return the input struct (without ID).
	// For simplicity here, we return the input struct.
	log.Printf("User %s created successfully", username)
	return user, nil
}

// GetUserByUsername retrieves a user by username
func GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	var user models.User
	collection := database.DB.Collection("users")
	filter := bson.M{"username": username}

	err := collection.FindOne(ctx, filter).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil // User not found is not necessarily an error in auth flow
		}
		log.Printf("Error finding user %s: %v", username, err)
		return nil, err
	}
	return &user, nil
}

// EnsureUsernameIndex creates a unique index on the username field if it doesn't exist.
// Call this once during application startup.
func EnsureUsernameIndex(ctx context.Context) error {
	collection := database.DB.Collection("users")
	indexModel := mongo.IndexModel{
		Keys:    bson.D{{Key: "username", Value: 1}},
		Options: options.Index().SetUnique(true),
	}
	_, err := collection.Indexes().CreateOne(ctx, indexModel)
	if err != nil {
		log.Printf("Error creating unique index for username: %v", err)
		return err
	}
	log.Println("Unique index for username ensured.")
	return nil
}

// --- Basic Auth Middleware ---

// BasicAuthMiddleware creates a Fiber middleware for Basic Authentication
func BasicAuthMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		auth := c.Get(fiber.HeaderAuthorization)
		if auth == "" {
			return c.Status(fiber.StatusUnauthorized).SendString("Missing Authorization Header")
		}

		// Check if the header starts with "Basic "
		const prefix = "Basic "
		if !strings.HasPrefix(auth, prefix) {
			return c.Status(fiber.StatusUnauthorized).SendString("Invalid Authorization Header format")
		}

		// Decode base64 credentials
		encodedCredentials := auth[len(prefix):]
		decodedBytes, err := base64.StdEncoding.DecodeString(encodedCredentials)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).SendString("Invalid Base64 credentials")
		}
		credentials := string(decodedBytes)

		// Split username and password
		parts := strings.SplitN(credentials, ":", 2)
		if len(parts) != 2 {
			return c.Status(fiber.StatusUnauthorized).SendString("Invalid credentials format")
		}
		username := parts[0]
		password := parts[1]

		// Verify credentials against database
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		user, err := GetUserByUsername(ctx, username)
		if err != nil {
			log.Printf("Error retrieving user %s during auth: %v", username, err)
			return c.Status(fiber.StatusInternalServerError).SendString("Internal Server Error")
		}
		if user == nil || !CheckPasswordHash(password, user.Password) {
			return c.Status(fiber.StatusUnauthorized).SendString("Invalid username or password")
		}

		// Store user information in context for downstream handlers
		c.Locals("user", user)
		log.Printf("User %s authenticated successfully", username)
		return c.Next()
	}
}
