package main

import (
	"context"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoDB struct {
	Client   *mongo.Client
	Database *mongo.Database
}

// InitDB initializes the MongoDB connection and returns a MongoDB wrapper
func InitDB() *MongoDB {
	uri := getEnv("MONGO_URI", "mongodb://localhost:27017")
	dbName := getEnv("DB_NAME", "four_in_a_row")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clientOpts := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		log.Fatalf("❌ Failed to connect to MongoDB: %v", err)
	}

	if err = client.Ping(ctx, nil); err != nil {
		log.Fatalf("❌ MongoDB ping failed: %v", err)
	}

	log.Println("✅ Connected to MongoDB successfully")
	db := client.Database(dbName)
	return &MongoDB{
		Client:   client,
		Database: db,
	}
}

// Disconnect closes the MongoDB connection gracefully
func (db *MongoDB) Disconnect() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := db.Client.Disconnect(ctx); err != nil {
		log.Printf("❌ Error disconnecting MongoDB: %v", err)
	} else {
		log.Println("✅ MongoDB disconnected successfully")
	}
}

// getEnv reads an environment variable or returns fallback if not set
func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
