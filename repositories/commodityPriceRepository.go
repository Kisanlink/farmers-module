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
}

func NewCommodityPriceRepository(db *mongo.Database) *CommodityPriceRepository {
	return &CommodityPriceRepository{
		Collection: db.Collection("ComodityPrice"),
	}
}

// GetCommodityPriceByCropID retrieves the commodity price for a given crop ID
func (repo *CommodityPriceRepository) GetCommodityPriceByCropID(ctx context.Context, crop string) (*models.CommodityPrice, error) {
	var price models.CommodityPrice

	// Log the cropID being queried
	log.Printf("DEBUG: Starting query for commodity price for cropID: %s", crop)

	// Query for commodity price by cropID
	err := repo.Collection.FindOne(ctx, bson.M{"comodityName": crop}).Decode(&price)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			log.Printf("DEBUG: No commodity price found for crop: %s", crop)
			return nil, nil // No price found for the crop
		}
		log.Printf("ERROR: Failed to retrieve commodity price for cropID %s: %v", crop, err)
		return nil, err
	}

	// Debug log: Successfully retrieved commodity price
	log.Printf("DEBUG: Successfully retrieved commodity price for cropID: %s, Price: %.2f", crop, price.Price)

	return &price, nil
}
