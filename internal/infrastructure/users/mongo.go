package users

import (
	"context"
	"errors"
	"time"

	domain "simpleservicedesk/internal/domain/users"
	"simpleservicedesk/internal/queries"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type mongoUser struct {
	ID             primitive.ObjectID `bson:"_id,omitempty"`
	UserID         uuid.UUID          `bson:"user_id"`
	Name           string             `bson:"name"`
	Email          string             `bson:"email"`
	PasswordHash   []byte             `bson:"password_hash"`
	Role           string             `bson:"role"`
	OrganizationID *uuid.UUID         `bson:"organization_id,omitempty"`
	IsActive       bool               `bson:"is_active"`
	CreatedAt      time.Time          `bson:"created_at"`
	UpdatedAt      time.Time          `bson:"updated_at"`
}

type MongoRepo struct {
	collection *mongo.Collection
}

func NewMongoRepo(db *mongo.Database) *MongoRepo {
	return &MongoRepo{
		collection: db.Collection("users"),
	}
}

func (r *MongoRepo) CreateUser(
	ctx context.Context,
	email string,
	passwordHash []byte,
	createFn func() (*domain.User, error),
) (*domain.User, error) {
	count, err := r.collection.CountDocuments(ctx, bson.M{"email": email})
	if err != nil {
		return nil, err
	}
	if count > 0 {
		return nil, domain.ErrUserAlreadyExist
	}

	u, err := createFn()
	if err != nil {
		return nil, err
	}

	mu := mongoUser{
		UserID:         u.ID(),
		Name:           u.Name(),
		Email:          email,
		PasswordHash:   passwordHash,
		Role:           string(u.Role()),
		OrganizationID: u.OrganizationID(),
		IsActive:       u.IsActive(),
		CreatedAt:      u.CreatedAt(),
		UpdatedAt:      u.UpdatedAt(),
	}
	_, err = r.collection.InsertOne(ctx, mu)
	if err != nil {
		return nil, err
	}

	return u, nil
}

func (r *MongoRepo) GetUser(ctx context.Context, userID uuid.UUID) (*domain.User, error) {
	var mu mongoUser
	err := r.collection.FindOne(ctx, bson.M{"user_id": userID}).Decode(&mu)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, domain.ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}

	role, err := domain.ParseRole(mu.Role)
	if err != nil {
		return nil, err
	}

	user, err := domain.NewUserWithDetails(
		userID, mu.Name, mu.Email, mu.PasswordHash, role,
		mu.OrganizationID, mu.IsActive, mu.CreatedAt, mu.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *MongoRepo) UpdateUser(ctx context.Context,
	userID uuid.UUID,
	updateFn func(*domain.User) (bool, error)) (*domain.User, error) {
	var mu mongoUser
	err := r.collection.FindOne(ctx, bson.M{"user_id": userID}).Decode(&mu)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, domain.ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}

	role, err := domain.ParseRole(mu.Role)
	if err != nil {
		return nil, err
	}

	entity, err := domain.NewUserWithDetails(
		mu.UserID, mu.Name, mu.Email, mu.PasswordHash, role,
		mu.OrganizationID, mu.IsActive, mu.CreatedAt, mu.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	updated, err := updateFn(entity)
	if err != nil {
		return nil, err
	}
	if !updated {
		return entity, nil
	}

	update := bson.M{"$set": bson.M{
		"name":            entity.Name(),
		"email":           entity.Email(),
		"role":            string(entity.Role()),
		"organization_id": entity.OrganizationID(),
		"is_active":       entity.IsActive(),
		"updated_at":      entity.UpdatedAt(),
	}}
	_, err = r.collection.UpdateOne(ctx, bson.M{"user_id": userID}, update)
	if err != nil {
		return nil, err
	}
	return entity, nil
}

func (r *MongoRepo) ListUsers(ctx context.Context, filter queries.UserFilter) ([]*domain.User, error) {
	bsonFilter := bson.M{}

	if filter.Name != nil && *filter.Name != "" {
		bsonFilter["name"] = bson.M{"$regex": *filter.Name, "$options": "i"}
	}
	if filter.Email != nil && *filter.Email != "" {
		bsonFilter["email"] = bson.M{"$regex": *filter.Email, "$options": "i"}
	}
	if filter.Role != nil && *filter.Role != "" {
		bsonFilter["role"] = *filter.Role
	}
	if filter.OrganizationID != nil {
		bsonFilter["organization_id"] = *filter.OrganizationID
	}
	if filter.IsActive != nil {
		bsonFilter["is_active"] = *filter.IsActive
	}

	opts := options.Find()
	opts.SetSort(bson.D{{Key: "created_at", Value: -1}})
	if filter.Limit > 0 {
		opts.SetLimit(int64(filter.Limit))
	}
	if filter.Offset > 0 {
		opts.SetSkip(int64(filter.Offset))
	}

	cursor, err := r.collection.Find(ctx, bsonFilter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var users []*domain.User
	for cursor.Next(ctx) {
		var mu mongoUser
		if decodeErr := cursor.Decode(&mu); decodeErr != nil {
			return nil, decodeErr
		}

		role, parseErr := domain.ParseRole(mu.Role)
		if parseErr != nil {
			return nil, parseErr
		}

		user, userErr := domain.NewUserWithDetails(
			mu.UserID, mu.Name, mu.Email, mu.PasswordHash, role,
			mu.OrganizationID, mu.IsActive, mu.CreatedAt, mu.UpdatedAt,
		)
		if userErr != nil {
			return nil, userErr
		}

		users = append(users, user)
	}

	return users, cursor.Err()
}

func (r *MongoRepo) DeleteUser(ctx context.Context, userID uuid.UUID) error {
	result, err := r.collection.DeleteOne(ctx, bson.M{"user_id": userID})
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		return domain.ErrUserNotFound
	}
	return nil
}

func (r *MongoRepo) CountUsers(ctx context.Context, filter queries.UserFilter) (int64, error) {
	bsonFilter := bson.M{}

	if filter.Name != nil && *filter.Name != "" {
		bsonFilter["name"] = bson.M{"$regex": *filter.Name, "$options": "i"}
	}
	if filter.Email != nil && *filter.Email != "" {
		bsonFilter["email"] = bson.M{"$regex": *filter.Email, "$options": "i"}
	}
	if filter.Role != nil && *filter.Role != "" {
		bsonFilter["role"] = *filter.Role
	}
	if filter.OrganizationID != nil {
		bsonFilter["organization_id"] = *filter.OrganizationID
	}
	if filter.IsActive != nil {
		bsonFilter["is_active"] = *filter.IsActive
	}

	return r.collection.CountDocuments(ctx, bsonFilter)
}
