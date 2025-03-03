package database

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/Kisanlink/farmers-module/config"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	// MongoDB variables
	clientInstance    *mongo.Client
	clientInstanceErr error
	dbInstance        *mongo.Database
	mongoOnce         sync.Once

	// PostgreSQL variables
	pgDB     *gorm.DB
	pgOnce   sync.Once
	pgDBErr  error
)

// InitializeMongoDB initializes MongoDB connection.
func InitializeMongoDB() {
	mongoOnce.Do(func() {
		// Get MongoDB connection details from environment variables
		mongoURI := fmt.Sprintf("mongodb://%s:%s@%s:%s/%s",
			config.GetEnv("MONGO_USERNAME"),
			config.GetEnv("MONGO_PASSWORD"),
			config.GetEnv("MONGO_HOSTNAME"),
			config.GetEnv("MONGO_PORT"),
			config.GetEnv("MONGO_DB_NAME"),
		)

		// Create MongoDB client
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		clientInstance, clientInstanceErr = mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
		if clientInstanceErr != nil {
			log.Fatalf("Failed to connect to MongoDB: %v", clientInstanceErr)
		}

		// Ping MongoDB
		clientInstanceErr = clientInstance.Ping(ctx, nil)
		if clientInstanceErr != nil {
			log.Fatalf("Failed to ping MongoDB: %v", clientInstanceErr)
		}

		// Set global database instance
		dbInstance = clientInstance.Database(config.GetEnv("MONGO_DB_NAME"))
		log.Println("Connected to MongoDB successfully")
	})
}

// GetMongoDatabase returns MongoDB database instance.
func GetMongoDatabase() *mongo.Database {
	if dbInstance == nil {
		log.Fatal("MongoDB connection is not initialized. Call InitializeMongoDB first.")
	}
	return dbInstance
}

// GetMongoClient returns MongoDB client instance.
func GetMongoClient() *mongo.Client {
	if clientInstance == nil {
		log.Fatal("MongoDB client is not initialized. Call InitializeMongoDB first.")
	}
	return clientInstance
}

// InitializePostgres initializes PostgreSQL connection.
func InitializePostgres() {
	pgOnce.Do(func() {
		// Get PostgreSQL connection details from environment variables
		dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
			config.GetEnv("POSTGRES_HOST"),
			config.GetEnv("POSTGRES_PORT"),
			config.GetEnv("POSTGRES_USER"),
			config.GetEnv("POSTGRES_PASSWORD"),
			config.GetEnv("POSTGRES_DB_NAME"),
			config.GetEnv("POSTGRES_SSL_MODE"),
		)

		// Connect to PostgreSQL
		pgDB, pgDBErr = gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if pgDBErr != nil {
			log.Fatalf("Failed to connect to PostgreSQL: %v", pgDBErr)
		}

		log.Println("Connected to PostgreSQL successfully")
	})
}

// GetPGDB returns PostgreSQL database instance.
func GetPGDB() *gorm.DB {
	if pgDB == nil {
		log.Fatal("PostgreSQL connection is not initialized. Call InitializePostgres first.")
	}
	return pgDB
}
