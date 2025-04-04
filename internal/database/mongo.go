package database

import (
	"context"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var DB *mongo.Database

// ConnectDB initializes the MongoDB connection
func ConnectDB() error {
	// TODO: Use environment variables for connection string and DB name
	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		mongoURI = "mongodb://localhost:27017" // Default for local development
	}
	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		dbName = "simpleservicedesk"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Printf("Error connecting to MongoDB: %v", err)
		return err
	}

	// Ping the primary
	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		log.Printf("Error pinging MongoDB: %v", err)
		return err
	}

	log.Println("Connected to MongoDB!")
	DB = client.Database(dbName)
	return nil
}

// DisconnectDB closes the MongoDB connection
func DisconnectDB(ctx context.Context) error {
	if DB != nil && DB.Client() != nil {
		log.Println("Disconnecting from MongoDB...")
		if err := DB.Client().Disconnect(ctx); err != nil {
			log.Printf("Error disconnecting from MongoDB: %v", err)
			return err
		}
		log.Println("Disconnected from MongoDB.")
	}
	return nil
}
