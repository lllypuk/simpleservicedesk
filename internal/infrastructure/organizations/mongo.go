package organizations

import (
	"context"
	"errors"
	"fmt"
	"time"

	"simpleservicedesk/internal/application"
	domain "simpleservicedesk/internal/domain/organizations"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type mongoOrganization struct {
	ID        primitive.ObjectID          `bson:"_id,omitempty"`
	OrgID     uuid.UUID                   `bson:"organization_id"`
	Name      string                      `bson:"name"`
	Domain    string                      `bson:"domain"`
	ParentID  *uuid.UUID                  `bson:"parent_id,omitempty"`
	IsActive  bool                        `bson:"is_active"`
	Settings  domain.OrganizationSettings `bson:"settings"`
	CreatedAt time.Time                   `bson:"created_at"`
	UpdatedAt time.Time                   `bson:"updated_at"`
}

type MongoRepo struct {
	collection *mongo.Collection
}

func NewMongoRepo(db *mongo.Database) *MongoRepo {
	return &MongoRepo{
		collection: db.Collection("organizations"),
	}
}

func (r *MongoRepo) CreateOrganization(
	ctx context.Context,
	createFn func() (*domain.Organization, error),
) (*domain.Organization, error) {
	organization, err := createFn()
	if err != nil {
		return nil, err
	}

	// Check if organization with same name exists
	count, err := r.collection.CountDocuments(ctx, bson.M{
		"name": organization.Name(),
	})
	if err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, domain.ErrOrganizationAlreadyExist
	}

	// If organization has parent, verify parent exists
	if organization.HasParent() {
		parentCount, parentErr := r.collection.CountDocuments(ctx, bson.M{
			"organization_id": *organization.ParentID(),
		})
		if parentErr != nil {
			return nil, parentErr
		}
		if parentCount == 0 {
			return nil, fmt.Errorf("parent organization not found: %v", *organization.ParentID())
		}
	}

	mo := mongoOrganization{
		OrgID:     organization.ID(),
		Name:      organization.Name(),
		Domain:    organization.Domain(),
		ParentID:  organization.ParentID(),
		IsActive:  organization.IsActive(),
		Settings:  organization.Settings(),
		CreatedAt: organization.CreatedAt(),
		UpdatedAt: organization.UpdatedAt(),
	}

	_, err = r.collection.InsertOne(ctx, mo)
	if err != nil {
		return nil, err
	}

	return organization, nil
}

func (r *MongoRepo) GetOrganization(ctx context.Context, orgID uuid.UUID) (*domain.Organization, error) {
	var mo mongoOrganization
	err := r.collection.FindOne(ctx, bson.M{"organization_id": orgID}).Decode(&mo)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, domain.ErrOrganizationNotFound
	}
	if err != nil {
		return nil, err
	}

	organization, err := domain.NewOrganization(
		mo.OrgID,
		mo.Name,
		mo.Domain,
		mo.ParentID,
	)
	if err != nil {
		return nil, err
	}

	// Restore settings and activation state
	organization.UpdateSettings(mo.Settings)
	if !mo.IsActive {
		organization.Deactivate()
	}

	return organization, nil
}

func (r *MongoRepo) UpdateOrganization(
	ctx context.Context,
	orgID uuid.UUID,
	updateFn func(*domain.Organization) (bool, error),
) (*domain.Organization, error) {
	var mo mongoOrganization
	err := r.collection.FindOne(ctx, bson.M{"organization_id": orgID}).Decode(&mo)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, domain.ErrOrganizationNotFound
	}
	if err != nil {
		return nil, err
	}

	organization, err := domain.NewOrganization(
		mo.OrgID,
		mo.Name,
		mo.Domain,
		mo.ParentID,
	)
	if err != nil {
		return nil, err
	}

	// Restore settings and activation state
	organization.UpdateSettings(mo.Settings)
	if !mo.IsActive {
		organization.Deactivate()
	}

	updated, err := updateFn(organization)
	if err != nil {
		return nil, err
	}
	if !updated {
		return organization, nil
	}

	// Check for circular reference if parent changed
	if organization.HasParent() {
		if circularErr := r.checkCircularReference(ctx, orgID, *organization.ParentID()); circularErr != nil {
			return nil, circularErr
		}
	}

	update := bson.M{"$set": bson.M{
		"name":       organization.Name(),
		"domain":     organization.Domain(),
		"parent_id":  organization.ParentID(),
		"is_active":  organization.IsActive(),
		"settings":   organization.Settings(),
		"updated_at": organization.UpdatedAt(),
	}}

	_, err = r.collection.UpdateOne(ctx, bson.M{"organization_id": orgID}, update)
	if err != nil {
		return nil, err
	}

	return organization, nil
}

func (r *MongoRepo) ListOrganizations(
	ctx context.Context,
	filter application.OrganizationFilter,
) ([]*domain.Organization, error) {
	query := bson.M{}

	// Apply filters
	if filter.ParentID != nil {
		query["parent_id"] = *filter.ParentID
	}
	if filter.IsActive != nil {
		query["is_active"] = *filter.IsActive
	}
	if filter.Name != nil {
		query["name"] = bson.M{"$regex": *filter.Name, "$options": "i"}
	}
	if filter.Domain != nil {
		query["domain"] = bson.M{"$regex": *filter.Domain, "$options": "i"}
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

	var organizations []*domain.Organization
	for cursor.Next(ctx) {
		var mo mongoOrganization
		if decodeErr := cursor.Decode(&mo); decodeErr != nil {
			return nil, decodeErr
		}

		organization, orgErr := domain.NewOrganization(
			mo.OrgID,
			mo.Name,
			mo.Domain,
			mo.ParentID,
		)
		if orgErr != nil {
			return nil, orgErr
		}

		// Restore settings and activation state
		organization.UpdateSettings(mo.Settings)
		if !mo.IsActive {
			organization.Deactivate()
		}

		organizations = append(organizations, organization)
	}

	if cursorErr := cursor.Err(); cursorErr != nil {
		return nil, cursorErr
	}

	return organizations, nil
}

func (r *MongoRepo) GetOrganizationHierarchy(
	ctx context.Context,
	rootID uuid.UUID,
) (*application.OrganizationTree, error) {
	root, err := r.GetOrganization(ctx, rootID)
	if err != nil {
		return nil, err
	}

	tree := &application.OrganizationTree{
		Organization: root,
		Children:     []*application.OrganizationTree{},
	}

	children, err := r.getChildOrganizations(ctx, rootID)
	if err != nil {
		return nil, err
	}

	for _, child := range children {
		childTree, childErr := r.GetOrganizationHierarchy(ctx, child.ID())
		if childErr != nil {
			return nil, childErr
		}
		tree.Children = append(tree.Children, childTree)
	}

	return tree, nil
}

func (r *MongoRepo) DeleteOrganization(ctx context.Context, orgID uuid.UUID) error {
	// Check if organization has children
	childCount, err := r.collection.CountDocuments(ctx, bson.M{"parent_id": orgID})
	if err != nil {
		return err
	}
	if childCount > 0 {
		return errors.New("cannot delete organization with children")
	}

	result, err := r.collection.DeleteOne(ctx, bson.M{"organization_id": orgID})
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		return domain.ErrOrganizationNotFound
	}

	return nil
}

// checkCircularReference verifies that setting newParentID as parent wouldn't create circular reference
func (r *MongoRepo) checkCircularReference(ctx context.Context, orgID, newParentID uuid.UUID) error {
	if orgID == newParentID {
		return domain.ErrCircularReference
	}

	// Traverse up the hierarchy from newParentID to see if we encounter orgID
	currentID := newParentID
	visited := make(map[uuid.UUID]bool)

	for !visited[currentID] {
		visited[currentID] = true

		if currentID == orgID {
			return domain.ErrCircularReference
		}

		var mo mongoOrganization
		err := r.collection.FindOne(ctx, bson.M{"organization_id": currentID}).Decode(&mo)
		if errors.Is(err, mongo.ErrNoDocuments) {
			break
		}
		if err != nil {
			return err
		}

		if mo.ParentID == nil {
			break
		}
		currentID = *mo.ParentID
	}

	return nil
}

// getChildOrganizations returns immediate children of an organization
func (r *MongoRepo) getChildOrganizations(ctx context.Context, parentID uuid.UUID) ([]*domain.Organization, error) {
	cursor, err := r.collection.Find(ctx, bson.M{"parent_id": parentID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var organizations []*domain.Organization
	for cursor.Next(ctx) {
		var mo mongoOrganization
		if decodeErr := cursor.Decode(&mo); decodeErr != nil {
			return nil, decodeErr
		}

		organization, orgErr := domain.NewOrganization(
			mo.OrgID,
			mo.Name,
			mo.Domain,
			mo.ParentID,
		)
		if orgErr != nil {
			return nil, orgErr
		}

		// Restore settings and activation state
		organization.UpdateSettings(mo.Settings)
		if !mo.IsActive {
			organization.Deactivate()
		}

		organizations = append(organizations, organization)
	}

	if cursorErr := cursor.Err(); cursorErr != nil {
		return nil, cursorErr
	}

	return organizations, nil
}
