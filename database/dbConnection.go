package database

import (
	"context"
	"fmt"
	"sync"
	"time"

	"log"

	"github.com/Kisanlink/farmers-module/config"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	clientInstance    *mongo.Client
	clientInstanceErr error
	dbInstance        *mongo.Database
	once              sync.Once
)

// InitializeDatabase initializes the MongoDB connection and sets the global database instance.
func InitializeDatabase() {
	once.Do(func() {
		// Load environment variables
		config.LoadEnv()

		// Get MongoDB connection details
		hostname := config.GetEnv("MONGO_HOSTNAME")
		port := config.GetEnv("MONGO_PORT")
		username := config.GetEnv("MONGO_USERNAME")
		password := config.GetEnv("MONGO_PASSWORD")
		dbName := config.GetEnv("MONGO_DB_NAME")

		// MongoDB URI
		mongoURI := fmt.Sprintf("mongodb://%s:%s@%s:%s/%s", username, password, hostname, port, dbName)

		// Create MongoDB client
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		clientInstance, clientInstanceErr = mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
		if clientInstanceErr != nil {
			log.Fatalf("Failed to connect to MongoDB: %v", clientInstanceErr)
		}

		// Ping the database
		clientInstanceErr = clientInstance.Ping(ctx, nil)
		if clientInstanceErr != nil {
			log.Fatalf("Failed to ping MongoDB: %v", clientInstanceErr)
		}

		// Set the global database instance
		dbInstance = clientInstance.Database(dbName)
		log.Println("Connected to MongoDB successfully")
	})
}

// GetDatabase returns the global database instance.
func GetDatabase() *mongo.Database {
	if dbInstance == nil {
		log.Fatal("Database connection is not initialized. Call InitializeDatabase first.")
	}
	return dbInstance
}

// GetClient returns the global MongoDB client instance.
func GetClient() *mongo.Client {
	if clientInstance == nil {
		log.Fatal("MongoDB client is not initialized. Call InitializeDatabase first.")
	}
	return clientInstance
}
