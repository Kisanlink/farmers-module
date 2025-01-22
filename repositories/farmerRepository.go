package repositories

import (
	"context"
	"log"

	"github.com/Kisanlink/farmers-module/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type FarmerRepository struct {
	Collection *mongo.Collection
}

func NewFarmerRepository(db *mongo.Database) *FarmerRepository {
	return &FarmerRepository{
		Collection: db.Collection("Farmers"),
	}
}

func (repo *FarmerRepository) GetFarmerByID(ctx context.Context, id int64) (*models.Farmer, error) {
	var farmer models.Farmer

	// Debug: Log the ID being queried
	log.Printf("DEBUG: Starting query for farmer with ID: %d", id)

	// Perform the query
	err := repo.Collection.FindOne(ctx, bson.M{"farmerId": id}).Decode(&farmer)

	// Debug: Check for errors
	if err != nil {
		if err == mongo.ErrNoDocuments {
			log.Printf("DEBUG: No farmer found with ID: %d", id)
			return nil, nil // No farmer found
		}

		log.Printf("ERROR: Failed to retrieve farmer with ID %d: %v", id, err)
		return nil, err
	}

	// Debug: Successful query
	log.Printf("DEBUG: Successfully retrieved farmer: %+v", farmer)

	return &farmer, nil
}

func (repo *FarmerRepository) CreateFarmer(ctx context.Context, farmer *models.Farmer) error {
	_, err := repo.Collection.InsertOne(ctx, farmer)
	return err
}
