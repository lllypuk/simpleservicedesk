package testhelpers

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
)

type MockCollection struct {
	InsertOneResult *mongo.InsertOneResult
	InsertOneErr    error
	FindResult      *mongo.Cursor
	FindErr         error
	UpdateResult    *mongo.UpdateResult
	UpdateErr       error
	DeleteResult    *mongo.DeleteResult
	DeleteErr       error
}

func (m *MockCollection) InsertOne(ctx context.Context, document interface{}) (*mongo.InsertOneResult, error) {
	return m.InsertOneResult, m.InsertOneErr
}

func (m *MockCollection) Find(ctx context.Context, filter interface{}) (*mongo.Cursor, error) {
	return m.FindResult, m.FindErr
}

func (m *MockCollection) UpdateByID(ctx context.Context, id interface{}, update interface{}) (*mongo.UpdateResult, error) {
	return m.UpdateResult, m.UpdateErr
}

func (m *MockCollection) DeleteOne(ctx context.Context, filter interface{}) (*mongo.DeleteResult, error) {
	return m.DeleteResult, m.DeleteErr
}
