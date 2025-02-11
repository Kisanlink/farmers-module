package repositories

import (
	"context"
"log"
	"github.com/Kisanlink/farmers-module/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type FarmerRepository struct {
	Collection *mongo.Collection
}

func NewFarmerRepository(db *mongo.Database) *FarmerRepository {
	return &FarmerRepository{
		Collection: db.Collection("Farmers"),
	}
}

func (repo *FarmerRepository) CreateFarmer(ctx context.Context, farmer *models.Farmer) error {
	// Insert the new farmer into the collection
	_, err := repo.Collection.InsertOne(ctx, farmer)
	return err
}

func (repo *FarmerRepository) UpdateFarmer(ctx context.Context, id int64, updatedFarmer *models.Farmer) error {
	// Find the farmer by FarmerID and update
	filter := bson.M{"farmerId": id}
	update := bson.M{
		"$set": updatedFarmer, // Update the entire farmer document with new data
	}

	// Perform the update operation
	result := repo.Collection.FindOneAndUpdate(ctx, filter, update)
	if result.Err() != nil {
		return result.Err()
	}

	return nil
}

func (repo *FarmerRepository) SearchFarmers(ctx context.Context, filter string, orderBy string, page int, perPage int) ([]models.Farmer, int64, error) {
	var farmers []models.Farmer
	

	// Build the filter query
	query := bson.M{}
	if filter != "" {
		query["firstName"] = bson.M{"$regex": "^" + filter, "$options": "i"} // Case-insensitive search
	}

	// Count total number of matching farmers
	count, err := repo.Collection.CountDocuments(ctx, query)
	if err != nil {
		return nil, 0, err
	}

	// Get the list of farmers with pagination and sorting
	options := &options.FindOptions{
		Sort: bson.M{orderBy: 1}, // Ascending order by default
		Skip: &[]int64{int64((page - 1) * perPage)}[0],
		Limit: &[]int64{int64(perPage)}[0],
	}

	cursor, err := repo.Collection.Find(ctx, query, options)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	// Decode the result into the farmers slice
	if err := cursor.All(ctx, &farmers); err != nil {
		return nil, 0, err
	}

	return farmers, count, nil
}




// GetFarmerByID retrieves a farmer by their farmerId
func (repo *FarmerRepository) GetFarmerByID(ctx context.Context, id int64) (*models.Farmer, error) {
	var farmer models.Farmer

	// Log the query attempt
	log.Printf("DEBUG: Starting query for farmer with ID: %d", id)

	// Query for farmer using farmerId (note: you might need to adjust the field name if it's different)
	err := repo.Collection.FindOne(ctx, bson.M{"farmerId": id}).Decode(&farmer)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			// If no farmer is found
			log.Printf("DEBUG: No farmer found with ID: %d", id)
			return nil, nil
		}

		log.Printf("ERROR: Failed to retrieve farmer with ID %d: %v", id, err)
		return nil, err
	}

	// Successfully found farmer
	log.Printf("DEBUG: Successfully retrieved farmer: %+v", farmer)

	return &farmer, nil
}
