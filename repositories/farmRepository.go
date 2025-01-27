package repositories

import (
	"context"
	"log"

	"github.com/Kisanlink/farmers-module/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type FarmRepository struct {
	Collection *mongo.Collection
}

func NewFarmRepository(db *mongo.Database) *FarmRepository {
	return &FarmRepository{
		Collection: db.Collection("Farms"),
	}
}

// GetFarms retrieves multiple farms by farmerID
func (repo *FarmRepository) GetFarms(ctx context.Context, farmerid int64) ([]models.Farm, error) {
	// Debug: Log the farmer ID being queried
	log.Printf("DEBUG: Starting query for farms of farmer with ID: %d", farmerid)

	// Perform the query using Find to get multiple farms for the given farmer ID
	cursor, err := repo.Collection.Find(ctx, bson.M{"farmerId": farmerid})
	if err != nil {
		log.Printf("ERROR: Failed to retrieve farms for farmerId %d: %v", farmerid, err)
		return nil, err
	}
	defer cursor.Close(ctx) // Ensure the cursor is closed once done iterating

	// Declare a slice to hold the farms
	var farms []models.Farm

	// Iterate through the cursor and decode each farm into the farms slice
	for cursor.Next(ctx) {
		var farm models.Farm
		if err := cursor.Decode(&farm); err != nil {
			log.Printf("ERROR: Failed to decode farm: %v", err)
			return nil, err
		}
		farms = append(farms, farm)
	}

	// Check for errors while iterating through the cursor
	if err := cursor.Err(); err != nil {
		log.Printf("ERROR: Cursor error: %v", err)
		return nil, err
	}

	// Debug: Successfully retrieved multiple farms
	log.Printf("DEBUG: Successfully retrieved %d farms for farmerId %d", len(farms), farmerid)

	// Return the slice of farms
	return farms, nil
}
