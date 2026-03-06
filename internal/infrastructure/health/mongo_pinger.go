package health

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
)

// MongoPinger wraps *mongo.Client and implements health.Pinger.
type MongoPinger struct {
	client *mongo.Client
}

// NewMongoPinger creates a MongoPinger for the given mongo client.
func NewMongoPinger(client *mongo.Client) *MongoPinger {
	return &MongoPinger{client: client}
}

// Ping pings MongoDB using the primary read preference.
func (p *MongoPinger) Ping(ctx context.Context) error {
	return p.client.Ping(ctx, nil)
}
