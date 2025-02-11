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

// GetFarms retrieves multiple farms by farmerID and filters fields based on the provided parameter
func (repo *FarmRepository) GetFarms(ctx context.Context, farmerid int64, fields string) ([]models.Farm, error) {
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

	// Return the slice of farms
	return farms, nil
}

// GetFarmByID retrieves a farm by its farmID
func (repo *FarmRepository) GetFarmByID(ctx context.Context, farmID int64) (*models.Farm, error) {
	var farm models.Farm
	err := repo.Collection.FindOne(ctx, bson.M{"farmId": farmID}).Decode(&farm)
	if err != nil {
		log.Printf("ERROR: Failed to retrieve farm with farmId %d: %v", farmID, err)
		return nil, err
	}
	return &farm, nil
}
