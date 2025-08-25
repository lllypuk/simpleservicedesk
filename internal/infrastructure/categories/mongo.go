package categories

import (
	"context"
	"errors"
	"fmt"
	"time"

	"simpleservicedesk/internal/application"
	domain "simpleservicedesk/internal/domain/categories"
	"simpleservicedesk/internal/queries"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type mongoCategory struct {
	ID             primitive.ObjectID `bson:"_id,omitempty"`
	CategoryID     uuid.UUID          `bson:"category_id"`
	Name           string             `bson:"name"`
	Description    string             `bson:"description"`
	OrganizationID uuid.UUID          `bson:"organization_id"`
	ParentID       *uuid.UUID         `bson:"parent_id,omitempty"`
	IsActive       bool               `bson:"is_active"`
	CreatedAt      time.Time          `bson:"created_at"`
	UpdatedAt      time.Time          `bson:"updated_at"`
}

type MongoRepo struct {
	collection *mongo.Collection
}

func NewMongoRepo(db *mongo.Database) *MongoRepo {
	return &MongoRepo{
		collection: db.Collection("categories"),
	}
}

func (r *MongoRepo) CreateCategory(
	ctx context.Context,
	createFn func() (*domain.Category, error),
) (*domain.Category, error) {
	category, err := createFn()
	if err != nil {
		return nil, err
	}

	// Check if category with same name exists in the same organization
	count, err := r.collection.CountDocuments(ctx, bson.M{
		"name":            category.Name(),
		"organization_id": category.OrganizationID(),
	})
	if err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, domain.ErrCategoryAlreadyExist
	}

	// If category has parent, verify parent exists
	if category.HasParent() {
		parentCount, parentErr := r.collection.CountDocuments(ctx, bson.M{
			"category_id": *category.ParentID(),
		})
		if parentErr != nil {
			return nil, parentErr
		}
		if parentCount == 0 {
			return nil, fmt.Errorf("parent category not found: %v", *category.ParentID())
		}
	}

	mc := mongoCategory{
		CategoryID:     category.ID(),
		Name:           category.Name(),
		Description:    category.Description(),
		OrganizationID: category.OrganizationID(),
		ParentID:       category.ParentID(),
		IsActive:       category.IsActive(),
		CreatedAt:      category.CreatedAt(),
		UpdatedAt:      category.UpdatedAt(),
	}

	_, err = r.collection.InsertOne(ctx, mc)
	if err != nil {
		return nil, err
	}

	return category, nil
}

func (r *MongoRepo) GetCategory(ctx context.Context, categoryID uuid.UUID) (*domain.Category, error) {
	var mc mongoCategory
	err := r.collection.FindOne(ctx, bson.M{"category_id": categoryID}).Decode(&mc)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, domain.ErrCategoryNotFound
	}
	if err != nil {
		return nil, err
	}

	category, err := domain.NewCategory(
		mc.CategoryID,
		mc.Name,
		mc.Description,
		mc.OrganizationID,
		mc.ParentID,
	)
	if err != nil {
		return nil, err
	}

	// Restore activation state if needed
	if !mc.IsActive {
		category.Deactivate()
	}

	return category, nil
}

func (r *MongoRepo) UpdateCategory(
	ctx context.Context,
	categoryID uuid.UUID,
	updateFn func(*domain.Category) (bool, error),
) (*domain.Category, error) {
	var mc mongoCategory
	err := r.collection.FindOne(ctx, bson.M{"category_id": categoryID}).Decode(&mc)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, domain.ErrCategoryNotFound
	}
	if err != nil {
		return nil, err
	}

	category, err := domain.NewCategory(
		mc.CategoryID,
		mc.Name,
		mc.Description,
		mc.OrganizationID,
		mc.ParentID,
	)
	if err != nil {
		return nil, err
	}

	// Restore activation state
	if !mc.IsActive {
		category.Deactivate()
	}

	updated, err := updateFn(category)
	if err != nil {
		return nil, err
	}
	if !updated {
		return category, nil
	}

	// Check for circular reference if parent changed
	if category.HasParent() {
		if circularErr := r.checkCircularReference(ctx, categoryID, *category.ParentID()); circularErr != nil {
			return nil, circularErr
		}
	}

	update := bson.M{"$set": bson.M{
		"name":        category.Name(),
		"description": category.Description(),
		"parent_id":   category.ParentID(),
		"is_active":   category.IsActive(),
		"updated_at":  category.UpdatedAt(),
	}}

	_, err = r.collection.UpdateOne(ctx, bson.M{"category_id": categoryID}, update)
	if err != nil {
		return nil, err
	}

	return category, nil
}

func (r *MongoRepo) ListCategories(
	ctx context.Context,
	filter queries.CategoryFilter,
) ([]*domain.Category, error) {
	query := bson.M{}

	// Apply filters
	if filter.OrganizationID != nil {
		query["organization_id"] = *filter.OrganizationID
	}
	if filter.ParentID != nil {
		query["parent_id"] = *filter.ParentID
	}
	if filter.IsActive != nil {
		query["is_active"] = *filter.IsActive
	}
	if filter.Name != nil {
		query["name"] = bson.M{"$regex": *filter.Name, "$options": "i"}
	}
	if filter.IsRootOnly {
		query["parent_id"] = bson.M{"$exists": false}
	}

	// Set up options
	opts := options.Find()
	if filter.Limit > 0 {
		opts.SetLimit(int64(filter.Limit))
	}
	if filter.Offset > 0 {
		opts.SetSkip(int64(filter.Offset))
	}

	// Set up sorting
	if filter.SortBy != "" {
		sortOrder := 1
		if filter.SortOrder == "desc" {
			sortOrder = -1
		}
		opts.SetSort(bson.M{filter.SortBy: sortOrder})
	} else {
		opts.SetSort(bson.M{"name": 1}) // Default sort by name
	}

	cursor, err := r.collection.Find(ctx, query, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var categories []*domain.Category
	for cursor.Next(ctx) {
		var mc mongoCategory
		if decodeErr := cursor.Decode(&mc); decodeErr != nil {
			return nil, decodeErr
		}

		category, categoryErr := domain.NewCategory(
			mc.CategoryID,
			mc.Name,
			mc.Description,
			mc.OrganizationID,
			mc.ParentID,
		)
		if categoryErr != nil {
			return nil, categoryErr
		}

		// Restore activation state
		if !mc.IsActive {
			category.Deactivate()
		}

		categories = append(categories, category)
	}

	if cursorErr := cursor.Err(); cursorErr != nil {
		return nil, cursorErr
	}

	return categories, nil
}

func (r *MongoRepo) GetCategoryHierarchy(
	ctx context.Context,
	rootID uuid.UUID,
) (*application.CategoryTree, error) {
	root, err := r.GetCategory(ctx, rootID)
	if err != nil {
		return nil, err
	}

	tree := &application.CategoryTree{
		Category: root,
		Children: []*application.CategoryTree{},
	}

	children, err := r.getChildCategories(ctx, rootID)
	if err != nil {
		return nil, err
	}

	for _, child := range children {
		childTree, childErr := r.GetCategoryHierarchy(ctx, child.ID())
		if childErr != nil {
			return nil, childErr
		}
		tree.Children = append(tree.Children, childTree)
	}

	return tree, nil
}

func (r *MongoRepo) DeleteCategory(ctx context.Context, categoryID uuid.UUID) error {
	// Check if category has children
	childCount, err := r.collection.CountDocuments(ctx, bson.M{"parent_id": categoryID})
	if err != nil {
		return err
	}
	if childCount > 0 {
		return errors.New("cannot delete category with children")
	}

	result, err := r.collection.DeleteOne(ctx, bson.M{"category_id": categoryID})
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		return domain.ErrCategoryNotFound
	}

	return nil
}

// checkCircularReference verifies that setting newParentID as parent wouldn't create circular reference
func (r *MongoRepo) checkCircularReference(ctx context.Context, categoryID, newParentID uuid.UUID) error {
	if categoryID == newParentID {
		return domain.ErrCircularReference
	}

	// Traverse up the hierarchy from newParentID to see if we encounter categoryID
	currentID := newParentID
	visited := make(map[uuid.UUID]bool)

	for !visited[currentID] {
		visited[currentID] = true

		if currentID == categoryID {
			return domain.ErrCircularReference
		}

		var mc mongoCategory
		err := r.collection.FindOne(ctx, bson.M{"category_id": currentID}).Decode(&mc)
		if errors.Is(err, mongo.ErrNoDocuments) {
			break
		}
		if err != nil {
			return err
		}

		if mc.ParentID == nil {
			break
		}
		currentID = *mc.ParentID
	}

	return nil
}

// getChildCategories returns immediate children of a category
func (r *MongoRepo) getChildCategories(ctx context.Context, parentID uuid.UUID) ([]*domain.Category, error) {
	cursor, err := r.collection.Find(ctx, bson.M{"parent_id": parentID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var categories []*domain.Category
	for cursor.Next(ctx) {
		var mc mongoCategory
		if decodeErr := cursor.Decode(&mc); decodeErr != nil {
			return nil, decodeErr
		}

		category, categoryErr := domain.NewCategory(
			mc.CategoryID,
			mc.Name,
			mc.Description,
			mc.OrganizationID,
			mc.ParentID,
		)
		if categoryErr != nil {
			return nil, categoryErr
		}

		// Restore activation state
		if !mc.IsActive {
			category.Deactivate()
		}

		categories = append(categories, category)
	}

	if cursorErr := cursor.Err(); cursorErr != nil {
		return nil, cursorErr
	}

	return categories, nil
}
