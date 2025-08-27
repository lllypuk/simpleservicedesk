package tickets

import (
	"context"
	"errors"
	"time"

	domain "simpleservicedesk/internal/domain/tickets"
	"simpleservicedesk/internal/queries"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// mongoTicket represents the MongoDB document structure for tickets
type mongoTicket struct {
	ID             primitive.ObjectID `bson:"_id,omitempty"`
	TicketID       uuid.UUID          `bson:"ticket_id"`
	Title          string             `bson:"title"`
	Description    string             `bson:"description"`
	Status         string             `bson:"status"`
	Priority       string             `bson:"priority"`
	OrganizationID uuid.UUID          `bson:"organization_id"`
	CategoryID     *uuid.UUID         `bson:"category_id,omitempty"`
	AuthorID       uuid.UUID          `bson:"author_id"`
	AssigneeID     *uuid.UUID         `bson:"assignee_id,omitempty"`
	Comments       []mongoComment     `bson:"comments"`
	Attachments    []mongoAttachment  `bson:"attachments"`
	CreatedAt      time.Time          `bson:"created_at"`
	UpdatedAt      time.Time          `bson:"updated_at"`
	ResolvedAt     *time.Time         `bson:"resolved_at,omitempty"`
	ClosedAt       *time.Time         `bson:"closed_at,omitempty"`
}

// mongoComment represents the MongoDB subdocument structure for comments
type mongoComment struct {
	ID         uuid.UUID `bson:"id"`
	TicketID   uuid.UUID `bson:"ticket_id"`
	AuthorID   uuid.UUID `bson:"author_id"`
	Content    string    `bson:"content"`
	IsInternal bool      `bson:"is_internal"`
	CreatedAt  time.Time `bson:"created_at"`
}

// mongoAttachment represents the MongoDB subdocument structure for attachments
type mongoAttachment struct {
	ID         uuid.UUID `bson:"id"`
	TicketID   uuid.UUID `bson:"ticket_id"`
	FileName   string    `bson:"file_name"`
	FileSize   int64     `bson:"file_size"`
	MimeType   string    `bson:"mime_type"`
	FilePath   string    `bson:"file_path"`
	UploadedBy uuid.UUID `bson:"uploaded_by"`
	CreatedAt  time.Time `bson:"created_at"`
}

// MongoRepo implements TicketRepository for MongoDB
type MongoRepo struct {
	collection *mongo.Collection
}

// NewMongoRepo creates a new MongoDB repository for tickets
func NewMongoRepo(db *mongo.Database) *MongoRepo {
	collection := db.Collection("tickets")

	// Create indexes for better performance
	ctx := context.Background()
	indexes := []mongo.IndexModel{
		{Keys: bson.D{{"ticket_id", 1}}, Options: options.Index().SetUnique(true)},
		{Keys: bson.D{{"status", 1}}},
		{Keys: bson.D{{"priority", 1}}},
		{Keys: bson.D{{"assignee_id", 1}}},
		{Keys: bson.D{{"author_id", 1}}},
		{Keys: bson.D{{"organization_id", 1}}},
		{Keys: bson.D{{"category_id", 1}}},
		{Keys: bson.D{{"created_at", -1}}},
		{Keys: bson.D{{"updated_at", -1}}},
	}

	_, _ = collection.Indexes().CreateMany(ctx, indexes)

	return &MongoRepo{
		collection: collection,
	}
}

// CreateTicket creates a new ticket in MongoDB
func (r *MongoRepo) CreateTicket(ctx context.Context, createFn func() (*domain.Ticket, error)) (*domain.Ticket, error) {
	ticket, err := createFn()
	if err != nil {
		return nil, err
	}

	mongoDoc := r.domainToMongo(ticket)
	_, err = r.collection.InsertOne(ctx, mongoDoc)
	if err != nil {
		return nil, err
	}

	return ticket, nil
}

// GetTicket retrieves a ticket by ID from MongoDB
func (r *MongoRepo) GetTicket(ctx context.Context, ticketID uuid.UUID) (*domain.Ticket, error) {
	var mongoDoc mongoTicket
	err := r.collection.FindOne(ctx, bson.M{"ticket_id": ticketID}).Decode(&mongoDoc)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, domain.ErrTicketNotFound
	}
	if err != nil {
		return nil, err
	}

	return r.mongoToDomain(&mongoDoc)
}

// UpdateTicket updates an existing ticket in MongoDB
func (r *MongoRepo) UpdateTicket(
	ctx context.Context,
	ticketID uuid.UUID,
	updateFn func(*domain.Ticket) (bool, error),
) (*domain.Ticket, error) {
	var mongoDoc mongoTicket
	err := r.collection.FindOne(ctx, bson.M{"ticket_id": ticketID}).Decode(&mongoDoc)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, domain.ErrTicketNotFound
	}
	if err != nil {
		return nil, err
	}

	ticket, err := r.mongoToDomain(&mongoDoc)
	if err != nil {
		return nil, err
	}

	updated, err := updateFn(ticket)
	if err != nil {
		return nil, err
	}
	if !updated {
		return ticket, nil
	}

	updatedDoc := r.domainToMongo(ticket)
	update := bson.M{"$set": bson.M{
		"title":       updatedDoc.Title,
		"description": updatedDoc.Description,
		"status":      updatedDoc.Status,
		"priority":    updatedDoc.Priority,
		"category_id": updatedDoc.CategoryID,
		"assignee_id": updatedDoc.AssigneeID,
		"comments":    updatedDoc.Comments,
		"attachments": updatedDoc.Attachments,
		"updated_at":  updatedDoc.UpdatedAt,
		"resolved_at": updatedDoc.ResolvedAt,
		"closed_at":   updatedDoc.ClosedAt,
	}}

	_, err = r.collection.UpdateOne(ctx, bson.M{"ticket_id": ticketID}, update)
	if err != nil {
		return nil, err
	}

	return ticket, nil
}

// ListTickets retrieves tickets based on filter criteria
func (r *MongoRepo) ListTickets(ctx context.Context, filter queries.TicketFilter) ([]*domain.Ticket, error) {
	query := r.buildFilterQuery(filter)

	opts := options.Find()

	// Set limit and offset
	if filter.Limit > 0 {
		opts.SetLimit(int64(filter.Limit))
	}
	if filter.Offset > 0 {
		opts.SetSkip(int64(filter.Offset))
	}

	// Set sorting
	sort := r.buildSortOptions(filter)
	if len(sort) > 0 {
		opts.SetSort(sort)
	}

	cursor, err := r.collection.Find(ctx, query, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var tickets []*domain.Ticket
	for cursor.Next(ctx) {
		var mongoDoc mongoTicket
		if err = cursor.Decode(&mongoDoc); err != nil {
			return nil, err
		}

		ticket, domainErr := r.mongoToDomain(&mongoDoc)
		if domainErr != nil {
			return nil, domainErr
		}

		// Apply overdue filter if specified (complex logic requiring business rules)
		if filter.IsOverdue != nil {
			if *filter.IsOverdue != ticket.IsOverdue() {
				continue
			}
		}

		tickets = append(tickets, ticket)
	}

	if err = cursor.Err(); err != nil {
		return nil, err
	}

	// If sorting by priority, we need to re-sort by weight since MongoDB sorted alphabetically
	if filter.SortBy == "priority" {
		r.sortTicketsByPriority(tickets, filter.SortOrder == "asc")
	}

	return tickets, nil
}

// sortTicketsByPriority sorts tickets by priority weight
func (r *MongoRepo) sortTicketsByPriority(tickets []*domain.Ticket, ascending bool) {
	for i := 0; i < len(tickets)-1; i++ {
		for j := 0; j < len(tickets)-i-1; j++ {
			var shouldSwap bool
			if ascending {
				shouldSwap = tickets[j].Priority().Weight() > tickets[j+1].Priority().Weight()
			} else {
				shouldSwap = tickets[j].Priority().Weight() < tickets[j+1].Priority().Weight()
			}

			if shouldSwap {
				tickets[j], tickets[j+1] = tickets[j+1], tickets[j]
			}
		}
	}
}

// DeleteTicket removes a ticket from MongoDB
func (r *MongoRepo) DeleteTicket(ctx context.Context, ticketID uuid.UUID) error {
	result, err := r.collection.DeleteOne(ctx, bson.M{"ticket_id": ticketID})
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		return domain.ErrTicketNotFound
	}
	return nil
}

// Helper methods for conversion between domain and MongoDB models

func (r *MongoRepo) domainToMongo(ticket *domain.Ticket) *mongoTicket {
	var comments []mongoComment
	for _, comment := range ticket.Comments() {
		comments = append(comments, mongoComment{
			ID:         comment.ID,
			TicketID:   comment.TicketID,
			AuthorID:   comment.AuthorID,
			Content:    comment.Content,
			IsInternal: comment.IsInternal,
			CreatedAt:  comment.CreatedAt,
		})
	}

	var attachments []mongoAttachment
	for _, attachment := range ticket.Attachments() {
		attachments = append(attachments, mongoAttachment{
			ID:         attachment.ID,
			TicketID:   attachment.TicketID,
			FileName:   attachment.FileName,
			FileSize:   attachment.FileSize,
			MimeType:   attachment.MimeType,
			FilePath:   attachment.FilePath,
			UploadedBy: attachment.UploadedBy,
			CreatedAt:  attachment.CreatedAt,
		})
	}

	return &mongoTicket{
		TicketID:       ticket.ID(),
		Title:          ticket.Title(),
		Description:    ticket.Description(),
		Status:         string(ticket.Status()),
		Priority:       string(ticket.Priority()),
		OrganizationID: ticket.OrganizationID(),
		CategoryID:     ticket.CategoryID(),
		AuthorID:       ticket.AuthorID(),
		AssigneeID:     ticket.AssigneeID(),
		Comments:       comments,
		Attachments:    attachments,
		CreatedAt:      ticket.CreatedAt(),
		UpdatedAt:      ticket.UpdatedAt(),
		ResolvedAt:     ticket.ResolvedAt(),
		ClosedAt:       ticket.ClosedAt(),
	}
}

func (r *MongoRepo) mongoToDomain(mongoDoc *mongoTicket) (*domain.Ticket, error) {
	priority, err := domain.ParsePriority(mongoDoc.Priority)
	if err != nil {
		return nil, err
	}

	ticket, err := domain.NewTicket(
		mongoDoc.TicketID,
		mongoDoc.Title,
		mongoDoc.Description,
		priority,
		mongoDoc.OrganizationID,
		mongoDoc.AuthorID,
		mongoDoc.CategoryID,
	)
	if err != nil {
		return nil, err
	}

	// Set the timestamps from the database
	ticket.SetCreatedAt(mongoDoc.CreatedAt)
	ticket.SetUpdatedAt(mongoDoc.UpdatedAt)
	ticket.SetResolvedAt(mongoDoc.ResolvedAt)
	ticket.SetClosedAt(mongoDoc.ClosedAt)

	// Restore the status without resetting timestamps
	status, err := domain.ParseStatus(mongoDoc.Status)
	if err != nil {
		return nil, err
	}
	ticket.SetStatus(status)

	// Restore assignee if exists
	if mongoDoc.AssigneeID != nil {
		if err = ticket.AssignTo(*mongoDoc.AssigneeID); err != nil {
			return nil, err
		}
	}

	// Restore comments
	for _, mongoComment := range mongoDoc.Comments {
		if err = ticket.AddComment(mongoComment.AuthorID, mongoComment.Content, mongoComment.IsInternal); err != nil {
			return nil, err
		}
	}

	// Restore attachments
	for _, mongoAttachment := range mongoDoc.Attachments {
		if err = ticket.AddAttachment(
			mongoAttachment.FileName,
			mongoAttachment.FileSize,
			mongoAttachment.MimeType,
			mongoAttachment.FilePath,
			mongoAttachment.UploadedBy,
		); err != nil {
			return nil, err
		}
	}

	return ticket, nil
}

func (r *MongoRepo) buildFilterQuery(filter queries.TicketFilter) bson.M {
	query := bson.M{}

	if filter.Status != nil {
		query["status"] = string(*filter.Status)
	}
	if filter.Priority != nil {
		query["priority"] = string(*filter.Priority)
	}
	if filter.AssigneeID != nil {
		query["assignee_id"] = *filter.AssigneeID
	}
	if filter.AuthorID != nil {
		query["author_id"] = *filter.AuthorID
	}
	if filter.OrganizationID != nil {
		query["organization_id"] = *filter.OrganizationID
	}
	if filter.CategoryID != nil {
		query["category_id"] = *filter.CategoryID
	}

	// Date range filters
	if filter.CreatedAfter != nil || filter.CreatedBefore != nil {
		createdQuery := bson.M{}
		if filter.CreatedAfter != nil {
			createdQuery["$gte"] = *filter.CreatedAfter
		}
		if filter.CreatedBefore != nil {
			createdQuery["$lte"] = *filter.CreatedBefore
		}
		query["created_at"] = createdQuery
	}

	if filter.UpdatedAfter != nil || filter.UpdatedBefore != nil {
		updatedQuery := bson.M{}
		if filter.UpdatedAfter != nil {
			updatedQuery["$gte"] = *filter.UpdatedAfter
		}
		if filter.UpdatedBefore != nil {
			updatedQuery["$lte"] = *filter.UpdatedBefore
		}
		query["updated_at"] = updatedQuery
	}

	return query
}

func (r *MongoRepo) buildSortOptions(filter queries.TicketFilter) bson.D {
	var sort bson.D

	sortBy := "created_at" // default
	if filter.SortBy != "" {
		sortBy = filter.SortBy
	}

	sortOrder := -1 // default: descending
	if filter.SortOrder == "asc" {
		sortOrder = 1
	}

	// For priority sorting, we need to convert string values to weights for proper ordering
	// This is handled in the application logic since MongoDB doesn't know about the priority weights
	// We'll sort by the string field but the application needs to pre-process the priorities
	sort = append(sort, bson.E{Key: sortBy, Value: sortOrder})

	return sort
}

// Clear removes all tickets from MongoDB (useful for testing)
func (r *MongoRepo) Clear(ctx context.Context) error {
	return r.collection.Drop(ctx)
}
