package users

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	domain "simpleservicedesk/internal/domain/users"
)

type mongoUser struct {
	ID           primitive.ObjectID `bson:"_id,omitempty"`
	UserID       uuid.UUID          `bson:"user_id"`
	Name         string             `bson:"name"`
	Email        string             `bson:"email"`
	PasswordHash []byte             `bson:"password_hash"`
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
		UserID:       u.ID(),
		Name:         u.Name(),
		Email:        email,
		PasswordHash: passwordHash,
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

	user, err := domain.NewUser(userID, mu.Name, mu.Email, mu.PasswordHash)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *MongoRepo) UpdateUser(ctx context.Context, userID uuid.UUID, updateFn func(*domain.User) (bool, error)) (*domain.User, error) {
	var mu mongoUser
	err := r.collection.FindOne(ctx, bson.M{"user_id": userID}).Decode(&mu)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return nil, domain.ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}

	entity, err := domain.NewUser(mu.UserID, mu.Name, mu.Email, mu.PasswordHash)
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
		"name":  entity.Name(),
		"email": entity.Email(),
	}}
	_, err = r.collection.UpdateOne(ctx, bson.M{"user_id": userID}, update)
	if err != nil {
		return nil, err
	}
	return entity, nil
}
