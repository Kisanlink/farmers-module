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
	Db         *mongo.Database
}

func NewCommodityPriceRepository(db *mongo.Database) *CommodityPriceRepository {
	return &CommodityPriceRepository{
		Collection: db.Collection("CommodityPrice"),
		Db:         db,
	}
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

// GetPriceByCropID fetches the price for a crop by cropId
func (repo *CommodityPriceRepository) GetPriceByCropID(ctx context.Context, cropID primitive.ObjectID) (*models.CommodityPrice, error) {
	var price models.CommodityPrice
	err := repo.Collection.FindOne(ctx, bson.M{"_id": cropID}).Decode(&price)
	if err != nil {
		log.Printf("ERROR: Failed to retrieve commodity price for cropID %s: %v", cropID.Hex(), err)
		return nil, err
	}
	return &price, nil
}
