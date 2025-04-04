package repository

import (
	"context"
	"errors"
	"log"
	"time"

	"simpleservicedesk/internal/database"
	"simpleservicedesk/internal/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// RequestRepository defines the interface for request data operations
type RequestRepository interface {
	GetAll(ctx context.Context) ([]models.Request, error)
	GetByID(ctx context.Context, id primitive.ObjectID) (*models.Request, error)
	Create(ctx context.Context, request *models.Request) (*models.Request, error)
	Update(ctx context.Context, id primitive.ObjectID, updateData bson.M) (*models.Request, error)
	// Delete(ctx context.Context, id primitive.ObjectID) error // Optional: Add later if needed
}

// mongoRequestRepository implements RequestRepository using MongoDB
type mongoRequestRepository struct {
	collection *mongo.Collection
}

// NewMongoRequestRepository creates a new instance of mongoRequestRepository
func NewMongoRequestRepository() RequestRepository {
	return &mongoRequestRepository{
		collection: database.DB.Collection("requests"),
	}
}

// GetAll retrieves all requests from the database
func (r *mongoRequestRepository) GetAll(ctx context.Context) ([]models.Request, error) {
	var requests []models.Request
	// Find options can be added here for sorting, pagination etc.
	findOptions := options.Find()
	findOptions.SetSort(bson.D{{Key: "createdAt", Value: -1}}) // Sort by newest first

	cursor, err := r.collection.Find(ctx, bson.M{}, findOptions)
	if err != nil {
		log.Printf("Error finding all requests: %v", err)
		return nil, err
	}
	defer cursor.Close(ctx)

	if err = cursor.All(ctx, &requests); err != nil {
		log.Printf("Error decoding requests: %v", err)
		return nil, err
	}

	// Handle case where no documents are found (returns empty slice, not error)
	if requests == nil {
		requests = []models.Request{}
	}

	return requests, nil
}

// GetByID retrieves a single request by its ID
func (r *mongoRequestRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*models.Request, error) {
	var request models.Request
	filter := bson.M{"_id": id}

	err := r.collection.FindOne(ctx, filter).Decode(&request)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil // Return nil, nil if not found
		}
		log.Printf("Error finding request by ID %s: %v", id.Hex(), err)
		return nil, err
	}
	return &request, nil
}

// Create inserts a new request into the database
func (r *mongoRequestRepository) Create(ctx context.Context, request *models.Request) (*models.Request, error) {
	// Ensure timestamps and default status are set
	now := primitive.NewDateTimeFromTime(time.Now())
	if request.CreatedAt == 0 { // Check if primitive.DateTime is zero value
		request.CreatedAt = now
	}
	request.UpdatedAt = now
	if request.Status == "" {
		request.Status = models.Open // Default status
	}

	result, err := r.collection.InsertOne(ctx, request)
	if err != nil {
		log.Printf("Error inserting request: %v", err)
		return nil, err
	}

	// Set the generated ID back to the request object
	if oid, ok := result.InsertedID.(primitive.ObjectID); ok {
		request.ID = oid
	} else {
		log.Printf("Warning: Could not get ObjectID from InsertOne result for request title: %s", request.Title)
	}

	return request, nil
}

// Update modifies an existing request in the database
func (r *mongoRequestRepository) Update(ctx context.Context, id primitive.ObjectID, updateData bson.M) (*models.Request, error) {
	// Ensure updatedAt is always updated
	updateData["updatedAt"] = primitive.NewDateTimeFromTime(time.Now())

	filter := bson.M{"_id": id}
	update := bson.M{"$set": updateData}
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After) // Return the updated document

	var updatedRequest models.Request
	err := r.collection.FindOneAndUpdate(ctx, filter, update, opts).Decode(&updatedRequest)

	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			log.Printf("Request not found for update: ID %s", id.Hex())
			return nil, nil // Return nil, nil if not found
		}
		log.Printf("Error updating request ID %s: %v", id.Hex(), err)
		return nil, err
	}

	return &updatedRequest, nil
}
