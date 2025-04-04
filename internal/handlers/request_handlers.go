package handlers

import (
	"log"
	// time import removed as it's unused here
	"simpleservicedesk/internal/models"
	"simpleservicedesk/internal/repository"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson" // Added
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// RequestHandler holds dependencies for request handlers
type RequestHandler struct {
	repo repository.RequestRepository
}

// NewRequestHandler creates a new RequestHandler
func NewRequestHandler(repo repository.RequestRepository) *RequestHandler {
	return &RequestHandler{repo: repo}
}

// GetRequests handles fetching and displaying a list of requests
func (h *RequestHandler) GetRequests(c *fiber.Ctx) error {
	requests, err := h.repo.GetAll(c.Context())
	if err != nil {
		log.Printf("Error fetching requests: %v", err)
		return c.Status(fiber.StatusInternalServerError).SendString("Error fetching requests")
	}

	// TODO: Render HTMX template for request list
	log.Printf("Handler GetRequests called, found %d requests", len(requests))
	// For now, return simple JSON
	return c.JSON(requests) // Return actual requests
}

// GetRequest handles fetching and displaying details of a single request
func (h *RequestHandler) GetRequest(c *fiber.Ctx) error {
	id := c.Params("id")
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Invalid Request ID format")
	}

	request, err := h.repo.GetByID(c.Context(), objectID)
	if err != nil {
		log.Printf("Error fetching request %s: %v", id, err)
		return c.Status(fiber.StatusInternalServerError).SendString("Error fetching request")
	}
	if request == nil {
		return c.Status(fiber.StatusNotFound).SendString("Request not found")
	}

	// TODO: Render HTMX template for request details
	log.Printf("Handler GetRequest called for ID: %s", id)
	// For now, return simple JSON
	return c.JSON(request) // Return actual request
}

// CreateRequest handles the creation of a new request
func (h *RequestHandler) CreateRequest(c *fiber.Ctx) error {
	var req models.Request
	if err := c.BodyParser(&req); err != nil {
		log.Printf("Error parsing request body: %v", err)
		return c.Status(fiber.StatusBadRequest).SendString("Invalid request body")
	}

	// --- Authentication Disabled ---
	// // Get user from context (set by auth middleware)
	// user, ok := c.Locals("user").(*models.User)
	// if !ok {
	// 	log.Println("Error: User not found in context during request creation")
	// 	return c.Status(fiber.StatusInternalServerError).SendString("Authentication context error")
	// }
	// --- Authentication Disabled ---

	// Basic Validation (Example)
	if req.Title == "" || req.Description == "" || req.Type == "" {
		return c.Status(fiber.StatusBadRequest).SendString("Title, Description, and Type are required")
	}
	// Add more validation for Type enum if needed

	// Set fields managed by the system
	// req.CreatedBy = user.ID // Cannot set CreatedBy without authenticated user
	// Status, CreatedAt, UpdatedAt are set by repository Create method

	createdRequest, err := h.repo.Create(c.Context(), &req)
	if err != nil {
		log.Printf("Error creating request: %v", err)
		return c.Status(fiber.StatusInternalServerError).SendString("Error creating request")
	}

	// TODO: Render HTMX fragment or redirect
	log.Printf("Handler CreateRequest called, created request ID %s", createdRequest.ID.Hex()) // Removed user info
	// For now, return simple JSON
	return c.Status(fiber.StatusCreated).JSON(createdRequest) // Return created request
}

// UpdateRequest handles updating an existing request (e.g., status, description)
func (h *RequestHandler) UpdateRequest(c *fiber.Ctx) error {
	id := c.Params("id")
	objectID, err := primitive.ObjectIDFromHex(id) // Now used
	if err != nil {
		return c.Status(fiber.StatusBadRequest).SendString("Invalid Request ID format")
	}

	// Use bson.M for flexibility, or define specific update structs
	var updateData bson.M
	if err := c.BodyParser(&updateData); err != nil {
		log.Printf("Error parsing update body for request %s: %v", id, err)
		return c.Status(fiber.StatusBadRequest).SendString("Invalid update data")
	}

	// --- Authentication Disabled ---
	// // Get user from context
	// user, ok := c.Locals("user").(*models.User)
	// if !ok {
	// 	log.Println("Error: User not found in context during request update")
	// 	return c.Status(fiber.StatusInternalServerError).SendString("Authentication context error")
	// }
	// --- Authentication Disabled ---

	// Basic Validation & Field Control (Example)
	// Prevent updating certain fields directly
	delete(updateData, "_id")
	delete(updateData, "createdBy")
	delete(updateData, "createdAt")
	// Add validation for allowed fields and values (e.g., Status enum)
	if status, ok := updateData["status"].(string); ok {
		isValidStatus := false
		for _, s := range []models.RequestStatus{models.Open, models.InProgress, models.Resolved, models.Closed} {
			if models.RequestStatus(status) == s {
				isValidStatus = true
				break
			}
		}
		if !isValidStatus {
			return c.Status(fiber.StatusBadRequest).SendString("Invalid status value")
		}
	}

	// UpdatedAt is handled by the repository Update method

	updatedRequest, err := h.repo.Update(c.Context(), objectID, updateData)
	if err != nil {
		log.Printf("Error updating request %s: %v", id, err)
		// Specific error handling can be added here (e.g., check for validation errors from repo)
		return c.Status(fiber.StatusInternalServerError).SendString("Error updating request")
	}
	if updatedRequest == nil { // Repository returns nil if not found
		return c.Status(fiber.StatusNotFound).SendString("Request not found")
	}

	// TODO: Render HTMX fragment or redirect
	log.Printf("Handler UpdateRequest called for ID %s", id) // Removed user info
	// For now, return simple JSON
	return c.JSON(updatedRequest) // Return updated request
}
