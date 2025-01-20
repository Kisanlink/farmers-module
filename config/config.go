package config

import (
	"context"
	"log"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var Client *mongo.Client

// InitDB initializes the MongoDB connection
func InitDB() {
	// MongoDB URI (you can replace this with your Atlas connection string or local MongoDB URI)
	uri := "mongodb+srv://AbhinayVarshith:muYYdUl2DlujEpR0@cluster0.xfxlw.mongodb.net/"
	// Adjust to your MongoDB URI

	// Set client options
	clientOptions := options.Client().ApplyURI(uri)

	// Connect to MongoDB
	var err error
	Client, err = mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		log.Fatal("Error connecting to MongoDB: ", err)
	}

	// Check the connection
	err = Client.Ping(context.Background(), nil)
	if err != nil {
		log.Fatal("Error pinging MongoDB: ", err)
	}

	log.Print("Connected to MongoDB successfully!")
}
