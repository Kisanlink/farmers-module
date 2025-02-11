package repositories

import (
	"context"
	"log"

	"github.com/Kisanlink/farmers-module/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type CommodityPriceRepository struct {
	Collection *mongo.Collection
	Db         *mongo.Database
}

func NewCommodityPriceRepository(db *mongo.Database) *CommodityPriceRepository {
	return &CommodityPriceRepository{
		Collection: db.Collection("ComodityPrice"),
		Db:         db,
	}
}

// GetCropsByFarmerID fetches all unique crops associated with a farmer
func (repo *CommodityPriceRepository) GetCropsByFarmerID(ctx context.Context, farmerID int) ([]string, error) {
	var crops []string

	// Query the database for farms belonging to the given farmer
	cursor, err := repo.Db.Collection("Farms").Find(ctx, bson.M{"farmerId": farmerID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	// Extract unique crop names
	cropSet := make(map[string]bool)
	for cursor.Next(ctx) {
		var farm models.Farm
		if err := cursor.Decode(&farm); err != nil {
			return nil, err
		}
		if farm.Crop != "" {
			cropSet[farm.Crop] = true
		}
	}

	// Convert map keys to slice
	for crop := range cropSet {
		crops = append(crops, crop)
	}

	log.Printf("DEBUG: Found %d unique crops for farmerID: %d", len(crops), farmerID)
	return crops, nil
}

// GetPricesForCrops fetches prices for multiple crops in a single query
func (repo *CommodityPriceRepository) GetPricesForCrops(ctx context.Context, cropNames []string) ([]models.CommodityPrice, error) {
	var prices []models.CommodityPrice

	// Log crops being queried
	log.Printf("DEBUG: Fetching prices for crops: %v", cropNames)

	// Query the database for prices of multiple crops at once
	cursor, err := repo.Collection.Find(ctx, bson.M{"comodityName": bson.M{"$in": cropNames}})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var price models.CommodityPrice
		if err := cursor.Decode(&price); err != nil {
			return nil, err
		}
		prices = append(prices, price)
	}

	// Log successful retrieval
	log.Printf("DEBUG: Successfully retrieved %d commodity prices", len(prices))

	return prices, nil
}

// GetAllPrices fetches prices for all crops
func (repo *CommodityPriceRepository) GetAllPrices(ctx context.Context) ([]models.CommodityPrice, error) {
	var prices []models.CommodityPrice

	// Query the database for all commodity prices
	cursor, err := repo.Collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var price models.CommodityPrice
		if err := cursor.Decode(&price); err != nil {
			return nil, err
		}
		prices = append(prices, price)
	}

	// Log successful retrieval
	log.Printf("DEBUG: Successfully retrieved %d commodity prices", len(prices))

	return prices, nil
}
