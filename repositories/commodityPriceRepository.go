package repositories

import (
	"context"
	"log"

	"github.com/Kisanlink/farmers-module/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type CommodityPriceRepository struct {
	Collection *mongo.Collection
}

func NewCommodityPriceRepository(db *mongo.Database) *CommodityPriceRepository {
	return &CommodityPriceRepository{
		Collection: db.Collection("ComodityPrice"),
	}
}

// GetAllPrices fetches prices for all crops
func (repo *CommodityPriceRepository) GetAllPrices(ctx context.Context) ([]models.CommodityPrice, error) {
	var prices []models.CommodityPrice

	// Query the database for all commodity prices
	cursor, err := repo.Collection.Find(ctx, bson.M{})
	if err != nil {
		log.Printf("ERROR: Failed to execute query: %v", err)
		return nil, err
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var price models.CommodityPrice
		if err := cursor.Decode(&price); err != nil {
			log.Printf("ERROR: Failed to decode document: %v", err)
			return nil, err
		}
		prices = append(prices, price)
	}

	// Check for errors during cursor iteration
	if err := cursor.Err(); err != nil {
		log.Printf("ERROR: Cursor iteration error: %v", err)
		return nil, err
	}

	// Log successful retrieval
	log.Printf("DEBUG: Successfully retrieved %d commodity prices", len(prices))

	// If prices are empty, log a debug message
	if len(prices) == 0 {
		log.Println("DEBUG: No commodity prices found in the database.")
	}

	return prices, nil
}

// GetPriceByName fetches the price for a specific crop by name
func (repo *CommodityPriceRepository) GetPriceByName(ctx context.Context, cropName string) (*models.CommodityPrice, error) {
	var price models.CommodityPrice

	// Query the database for the commodity price by crop name
	err := repo.Collection.FindOne(ctx, bson.M{"comodityName": cropName}).Decode(&price)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			log.Printf("DEBUG: No commodity price found for crop name: %s", cropName)
			return nil, nil
		}
		log.Printf("ERROR: Failed to execute query: %v", err)
		return nil, err
	}

	// Log successful retrieval
	log.Printf("DEBUG: Successfully retrieved commodity price for crop name: %s", cropName)

	return &price, nil
}

// GetPriceByID fetches the price for a specific crop by ID
func (repo *CommodityPriceRepository) GetPriceByID(ctx context.Context, id string) (*models.CommodityPrice, error) {
	var price models.CommodityPrice

	// Convert the ID string to an ObjectID
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		log.Printf("ERROR: Invalid ID format: %v", err)
		return nil, err
	}

	// Query the database for the commodity price by ID
	err = repo.Collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&price)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			log.Printf("DEBUG: No commodity price found for ID: %s", id)
			return nil, nil
		}
		log.Printf("ERROR: Failed to execute query: %v", err)
		return nil, err
	}

	// Log successful retrieval
	log.Printf("DEBUG: Successfully retrieved commodity price for ID: %s", id)

	return &price, nil
}
